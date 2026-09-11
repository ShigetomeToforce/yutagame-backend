package middleware

import (
	"fmt"
	"runtime/debug"
	"strconv"
	"strings"
	"time"
	"yutagame-backend/infrastructure/filelog"

	"github.com/labstack/echo/v4"
)

func resolveLogScope(path string) string {
	if strings.HasPrefix(path, "/api/admin") || strings.HasPrefix(path, "/admin") {
		return "admin"
	}
	return "app"
}

func resolveLogKind(path string) string {
	if strings.HasPrefix(path, "/api/") {
		return "api"
	}
	return "access"
}

func resolveLogLevel(status int) string {
	if status >= 500 {
		return "error"
	}
	if status >= 400 {
		return "warn"
	}
	return "info"
}

// resolveErrStatus はエラーからHTTPステータスを推定する。
// レスポンスがまだ書き込まれていない段階（エラーがハンドラチェーンを伝播中）でも
// c.Response().Status は 0(既定値) のままのため、echo.HTTPError からコードを取り出す。
func resolveErrStatus(c echo.Context, err error) int {
	if committed := c.Response().Committed; committed {
		if s := c.Response().Status; s != 0 {
			return s
		}
	}
	if he, ok := err.(*echo.HTTPError); ok && he.Code != 0 {
		return he.Code
	}
	if s := c.Response().Status; s != 0 {
		return s
	}
	return 500
}

func RequestFileLog(logger *filelog.DailyLogger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			start := time.Now()

			// パニックはこの内側で回収し、発生位置を含むスタックトレースをそのままログに残す。
			// echo標準の Recover はコンソールにしかスタックを出さず、管理画面のログに残らないため。
			defer func() {
				if r := recover(); r != nil {
					stack := string(debug.Stack())
					path := c.Request().URL.Path
					if path == "" {
						path = c.Path()
					}
					scope := resolveLogScope(path)
					logger.Log(scope, "error", "error", "panic recovered", map[string]interface{}{
						"method":     c.Request().Method,
						"path":       path,
						"statusCode": 500,
						"error":      fmt.Sprint(r),
						"stack":      stack,
					})
					err = echo.NewHTTPError(500, "internal server error")
				}
			}()

			err = next(c)

			path := c.Request().URL.Path
			if path == "" {
				path = c.Path()
			}

			status := status200IfNoErr(c, err)

			scope := resolveLogScope(path)
			kind := resolveLogKind(path)
			level := resolveLogLevel(status)
			fields := map[string]interface{}{
				"method":     c.Request().Method,
				"path":       path,
				"statusCode": status,
				"ip":         c.RealIP(),
				"userAgent":  c.Request().UserAgent(),
				"latencyMs":  strconv.FormatInt(time.Since(start).Milliseconds(), 10),
			}

			if q := c.Request().URL.RawQuery; q != "" {
				fields["query"] = q
			}

			logger.Log(scope, kind, level, "request completed", fields)

			// ハンドラが handler.RespondError() 経由でエラーを返した場合、
			// 発生時点のスタックトレースが echo.Context に載っているのでそれを記録する。
			if stack, ok := c.Get("errStack").(string); ok && stack != "" {
				msg, _ := c.Get("errMessage").(string)
				logger.Log(scope, "error", "error", "request failed", map[string]interface{}{
					"method":     c.Request().Method,
					"path":       path,
					"statusCode": status,
					"error":      msg,
					"stack":      stack,
				})
			} else if err != nil {
				logger.Log(scope, "error", "error", "request failed", map[string]interface{}{
					"method":     c.Request().Method,
					"path":       path,
					"statusCode": status,
					"error":      err.Error(),
				})
			}

			return err
		}
	}
}

func status200IfNoErr(c echo.Context, err error) int {
	if err != nil {
		return resolveErrStatus(c, err)
	}
	if s := c.Response().Status; s != 0 {
		return s
	}
	return 200
}
