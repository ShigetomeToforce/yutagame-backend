package database

import (
	"context"
	"errors"
	"yutagame-backend/domain/model"

	"gorm.io/gorm"
)

// =========================================================================
// 構造体＆コンストラクタ
// =========================================================================

// ManufacturerRepository メーカー情報に関するデータベース操作を担当するリポジトリ
type ManufacturerRepository struct {
	db *gorm.DB
}

// NewManufacturerRepository ManufacturerRepositoryの新しいインスタンスを生成するコンストラクタ
func NewManufacturerRepository(db *gorm.DB) *ManufacturerRepository {
	return &ManufacturerRepository{db: db}
}

// =========================================================================
// C: Create (作成)
// =========================================================================

// Create 新しいメーカー情報をデータベースに登録する
func (r *ManufacturerRepository) Create(ctx context.Context, genre *model.Manufacturer) error {
	return r.db.WithContext(ctx).Create(genre).Error
}

// =========================================================================
// R: Read (取得)
// =========================================================================

// FindByID メーカーID（主キー）を指定して、該当するメーカー情報を1件取得する
func (r *ManufacturerRepository) FindByID(ctx context.Context, id int64) (*model.Manufacturer, error) {
	var manufacturer model.Manufacturer
	err := r.db.WithContext(ctx).First(&manufacturer, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &manufacturer, err
}

// FindAll 登録されているすべてのメーカー情報をID昇順で取得する（ページングなし）
func (r *ManufacturerRepository) FindAll(ctx context.Context) ([]model.Manufacturer, error) {
	var manufacturers []model.Manufacturer
	err := r.db.WithContext(ctx).Order("id asc").Find(&manufacturers).Error
	return manufacturers, err
}

// FindAllWithPagination 指定された件数（limit）と開始位置（offset）に応じて、メーカー情報をID昇順で取得する
func (r *ManufacturerRepository) FindAllWithPagination(
	ctx context.Context,
	limit, offset int,
	whereQueries ...func(*gorm.DB) *gorm.DB,
) ([]model.Manufacturer, error) {
	return ExecuteFindWithPagination[model.Manufacturer](ctx, r.db, limit, offset, "id asc", nil, whereQueries...)
}

// CountAll ページングの総ページ数計算のため、条件に合致するメーカー情報の総件数を取得する
func (r *ManufacturerRepository) CountAll(ctx context.Context, whereQueries ...func(*gorm.DB) *gorm.DB) (int64, error) {
	return ExecuteCount[model.Manufacturer](ctx, r.db, whereQueries...)
}

// =========================================================================
// U: Update (更新)
// =========================================================================

// Update 既存のメーカー情報（名前、説明など）を更新する
func (r *ManufacturerRepository) Update(ctx context.Context, genre *model.Manufacturer) error {
	return r.db.WithContext(ctx).Save(genre).Error
}

// =========================================================================
// D: Delete (削除)
// =========================================================================

// Delete メーカーIDを指定して、該当するメーカー情報を物理削除する
func (r *ManufacturerRepository) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&model.Manufacturer{}, id).Error
}
