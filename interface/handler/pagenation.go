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
func HandleListOrPagination[T any](
	c echo.Context,
	getAllFn func(ctx context.Context) ([]T, error),
	getPageFn func(ctx context.Context, page, limit int, search string) ([]T, int64, int, error), // 💡 引数に search を追加
) error {
	ctx := c.Request().Context()
	pageStr := c.QueryParam("page")
	searchWord := c.QueryParam("q") // 💡 URLの ?q=xxx を取得

	// 1. page パラメータがない場合は、全データを配列で返す
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

	// 💡 UseCaseの関数に searchWord を流し込む
	data, totalCount, totalPages, err := getPageFn(ctx, page, limit, searchWord)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Message: err.Error()})
	}

	return c.JSON(http.StatusOK, PaginatedResponse[T]{
		Data:       data,
		TotalCount: totalCount,
		TotalPages: totalPages,
	})
}
