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
