package database

import (
	"context"

	"gorm.io/gorm"
)

// ExecuteCount 条件に合致する指定モデルの総件数を取得する汎用関数
func ExecuteCount[T any](
	ctx context.Context,
	db *gorm.DB,
	whereQueries ...func(*gorm.DB) *gorm.DB,
) (int64, error) {
	var count int64
	var model T

	query := db.WithContext(ctx).Model(&model)
	for _, where := range whereQueries {
		if where != nil {
			query = where(query)
		}
	}

	err := query.Count(&count).Error
	return count, err
}

// ExecuteFindWithPagination 条件・ページングを適用して、指定モデルの配列を取得する汎用関数
func ExecuteFindWithPagination[T any](
	ctx context.Context,
	db *gorm.DB,
	limit, offset int,
	order string,
	modifier func(*gorm.DB) *gorm.DB,
	whereQueries ...func(*gorm.DB) *gorm.DB,
) ([]T, error) {
	var items []T
	query := db.WithContext(ctx)

	// Preloadなどの追加処理を適用
	if modifier != nil {
		query = modifier(query)
	}

	for _, where := range whereQueries {
		query = where(query)
	}

	err := query.Limit(limit).Offset(offset).Order(order).Find(&items).Error
	return items, err
}
