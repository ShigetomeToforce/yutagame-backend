package handler

import (
	"context"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

// PaginatedResponse 任意の型 T に対応する汎用ページングレスポンス
type PaginatedResponse[T any] struct {
	Data       []T   `json:"data"` // admins や items などのキー名に依存しないよう、汎用的に "data" とします
	TotalCount int64 `json:"totalCount"`
	TotalPages int   `json:"totalPages"`
}

// HandleListOrPagination 全件取得とページング取得を自動判別してレスポンスを返す汎用関数
// - getAllFn: 引数に context を取り、全件のスライスを返す関数
// - getPageFn: 引数に context, page, limit を取り、スライス、総件数、総ページ数を返す関数
func HandleListOrPagination[T any](
	c echo.Context,
	getAllFn func(ctx context.Context) ([]T, error),
	getPageFn func(ctx context.Context, page, limit int) ([]T, int64, int, error),
) error {
	ctx := c.Request().Context()
	pageStr := c.QueryParam("page")

	// 💡 1. page パラメータがない場合は、従来の全件全データを配列（スライス）で返す
	if pageStr == "" {
		data, err := getAllFn(ctx)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, ErrorResponse{Message: err.Error()})
		}
		return c.JSON(http.StatusOK, data)
	}

	// 💡 2. page パラメータがある場合はページング処理に切り替え
	page, _ := strconv.Atoi(pageStr)
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	// 共通仕様として 10, 30, 50 以外は 10 に固定（必要に応じてカスタム可能にしてもOK）
	if limit != 10 && limit != 30 && limit != 50 {
		limit = 10
	}

	data, totalCount, totalPages, err := getPageFn(ctx, page, limit)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Message: err.Error()})
	}

	return c.JSON(http.StatusOK, PaginatedResponse[T]{
		Data:       data,
		TotalCount: totalCount,
		TotalPages: totalPages,
	})
}
