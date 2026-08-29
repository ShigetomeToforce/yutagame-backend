package handler

import (
	"strconv"

	"github.com/labstack/echo/v4"
)

// 文字列から整数値の配列に変換
func ParseInt64Array(c echo.Context, key string) []int64 {
	values := c.QueryParams()[key]
	if len(values) == 0 {
		return nil
	}

	result := make([]int64, 0, len(values))
	for _, v := range values {
		id, err := strconv.ParseInt(v, 10, 64)
		if err == nil {
			result = append(result, id)
		}
	}
	return result
}

// 文字列から真偽値を変換し、未指定時は nil を返す。
// これにより "指定しない / true / false" の3状態を表現できる。
func ParseOptionalBool(c echo.Context, key string) *bool {
	values := c.QueryParams()[key]
	if len(values) == 0 || values[0] == "" {
		return nil
	}

	parsed, err := strconv.ParseBool(values[0])
	if err != nil {
		return nil
	}
	return &parsed
}
