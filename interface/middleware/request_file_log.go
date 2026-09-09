package middleware

import (
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

func RequestFileLog(logger *filelog.DailyLogger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			err := next(c)

			path := c.Request().URL.Path
			if path == "" {
				path = c.Path()
			}

			status := c.Response().Status
			if status == 0 {
				if err != nil {
					status = 500
				} else {
					status = 200
				}
			}

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
			logger.Log(scope, "error", level, "request completed", fields)

			if err != nil {
				errFields := map[string]interface{}{
					"method":     c.Request().Method,
					"path":       path,
					"statusCode": status,
					"error":      err.Error(),
				}
				logger.Log(scope, "error", "error", "request failed", errFields)
			}

			return err
		}
	}
}
