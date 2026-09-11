package handler

import (
	"runtime/debug"

	"github.com/labstack/echo/v4"
)

// ErrorResponse はHTTP通信において、フロントエンドに一律でエラーを伝えるための器
type ErrorResponse struct {
	Message string `json:"message"`
}

// RespondError はエラーレスポンスを返しつつ、発生時点のスタックトレースを
// echo.Context に載せておく。ミドルウェア(RequestFileLog)がリクエスト終了時に
// これを取り出し、管理画面から確認できるエラーログへ発生箇所ごと記録する。
func RespondError(c echo.Context, status int, err error) error {
	c.Set("errStack", string(debug.Stack()))
	c.Set("errMessage", err.Error())
	return c.JSON(status, ErrorResponse{Message: err.Error()})
}
