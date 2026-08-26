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

// HandleListOrPagination 全件取得とページング取得（検索対応版）を自動判別してレスポンスを返す汎用関数
func HandleListOrPagination[T any, F any](
	c echo.Context,
	getAllFn func(ctx context.Context) ([]T, error),
	getPageFn func(ctx context.Context, page, limit int, filter F) ([]T, int64, int, error),
	buildFilterFn func(c echo.Context) F,
) error {
	ctx := c.Request().Context()

	// 1. page パラメータがない場合は、全データを配列で返す
	pageStr := c.QueryParam("page")
	if pageStr == "" {
		data, err := getAllFn(ctx)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, ErrorResponse{Message: err.Error()})
		}
		return c.JSON(http.StatusOK, data)
	}

	// 2. page パラメータがある場合はページング（＋検索）処理
	page, _ := strconv.Atoi(pageStr)
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit != 10 && limit != 30 && limit != 50 {
		limit = 10
	}

	filter := buildFilterFn(c)

	data, totalCount, totalPages, err := getPageFn(ctx, page, limit, filter)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Message: err.Error()})
	}

	return c.JSON(http.StatusOK, PaginatedResponse[T]{
		Data:       data,
		TotalCount: totalCount,
		TotalPages: totalPages,
	})
}
