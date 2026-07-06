package database

import (
	"context"

	"gorm.io/gorm"
)

// ExecuteCount 条件に合致する指定モデルの総件数を取得する汎用関数
func ExecuteCount[T any](ctx context.Context, db *gorm.DB, whereQueries ...func(*gorm.DB) *gorm.DB) (int64, error) {
	var count int64
	var model T

	tx := db.WithContext(ctx).Model(&model)
	for _, query := range whereQueries {
		if query != nil {
			tx = query(tx)
		}
	}

	err := tx.Count(&count).Error
	return count, err
}

// ExecuteFindWithPagination 条件・ページングを適用して、指定モデルの配列を取得する汎用関数
func ExecuteFindWithPagination[T any](ctx context.Context, db *gorm.DB, limit, offset int, order string, whereQueries ...func(*gorm.DB) *gorm.DB) ([]T, error) {
	var results []T

	tx := db.WithContext(ctx)
	for _, query := range whereQueries {
		if query != nil {
			tx = query(tx)
		}
	}

	err := tx.Order(order).Limit(limit).Offset(offset).Find(&results).Error
	return results, err
}
