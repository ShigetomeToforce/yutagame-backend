package database

import (
	"context"
	"errors"
	"yutagame-backend/domain/model"

	"gorm.io/gorm"
)

// GenreRepository ジャンルデータに関するデータベース操作を担当するリポジトリ
type GenreRepository struct {
	db *gorm.DB
}

// NewGenreRepository GenreRepositoryの新しいインスタンスを生成するコンストラクタ
func NewGenreRepository(db *gorm.DB) *GenreRepository {
	return &GenreRepository{db: db}
}

// =========================================================================
// C: Create (作成)
// =========================================================================

// Create 新しいジャンルをデータベースに登録する
func (r *GenreRepository) Create(ctx context.Context, genre *model.Genre) error {
	return r.db.WithContext(ctx).Create(genre).Error
}

// =========================================================================
// R: Read (取得)
// =========================================================================

// FindByID ジャンルID（主キー）を指定して、該当するジャンル情報を1件取得する
func (r *GenreRepository) FindByID(ctx context.Context, id int64) (*model.Genre, error) {
	var genre model.Genre
	err := r.db.WithContext(ctx).First(&genre, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil // レコードが見つからない場合はエラーにせずnilを返す
	}
	return &genre, err
}

// FindAll 登録されているすべてのジャンル情報をID昇順で取得する（ページングなし）
func (r *GenreRepository) FindAll(ctx context.Context) ([]model.Genre, error) {
	var genres []model.Genre
	err := r.db.WithContext(ctx).Order("id asc").Find(&genres).Error
	return genres, err
}

// FindAllWithPagination 指定された件数（limit）と開始位置（offset）に応じて、ジャンル情報をID昇順で取得する
func (r *GenreRepository) FindAllWithPagination(ctx context.Context, limit, offset int, whereQueries ...func(*gorm.DB) *gorm.DB) ([]model.Genre, error) {
	return ExecuteFindWithPagination[model.Genre](ctx, r.db, limit, offset, "id asc", whereQueries...)
}

// CountAll ページングの総ページ数計算のため、条件に合致するジャンルの総件数を取得する
func (r *GenreRepository) CountAll(ctx context.Context, whereQueries ...func(*gorm.DB) *gorm.DB) (int64, error) {
	return ExecuteCount[model.Genre](ctx, r.db, whereQueries...)
}

// =========================================================================
// U: Update (更新)
// =========================================================================

// Update 既存のジャンル情報（名前、説明など）を更新する
func (r *GenreRepository) Update(ctx context.Context, genre *model.Genre) error {
	return r.db.WithContext(ctx).Save(genre).Error
}

// =========================================================================
// D: Delete (削除)
// =========================================================================

// Delete ジャンルIDを指定して、該当するジャンルを物理削除する
func (r *GenreRepository) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&model.Genre{}, id).Error
}
