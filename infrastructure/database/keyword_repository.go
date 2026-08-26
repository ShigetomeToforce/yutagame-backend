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

// KeywordRepository キーワード情報に関するデータベース操作を担当するリポジトリ
type KeywordRepository struct {
	db *gorm.DB
}

// NewKeywordRepository KeywordRepositoryの新しいインスタンスを生成するコンストラクタ
func NewKeywordRepository(db *gorm.DB) *KeywordRepository {
	return &KeywordRepository{db: db}
}

// =========================================================================
// C: Create (作成)
// =========================================================================

// Create 新しいキーワード情報をデータベースに登録する
func (r *KeywordRepository) Create(ctx context.Context, k *model.Keyword) error {
	return r.db.WithContext(ctx).Create(k).Error
}

// =========================================================================
// R: Read (取得)
// =========================================================================

// FindByID キーワードID（主キー）を指定して、該当するキーワード情報を1件取得する
func (r *KeywordRepository) FindByID(ctx context.Context, id int64) (*model.Keyword, error) {
	var keyword model.Keyword
	err := r.db.WithContext(ctx).First(&keyword, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &keyword, err
}

// FindAll 登録されているすべてのキーワード情報をソート・ID昇順で取得する（ページングなし）
func (r *KeywordRepository) FindAll(ctx context.Context) ([]model.Keyword, error) {
	var keywords []model.Keyword
	err := r.db.WithContext(ctx).Order("sort_order asc, id asc").Find(&keywords).Error
	return keywords, err
}

// FindAllWithPagination 指定された件数（limit）と開始位置（offset）に応じて、キーワード情報をソート・ID昇順で取得する
func (r *KeywordRepository) FindAllWithPagination(
	ctx context.Context,
	limit, offset int,
	whereQueries ...func(*gorm.DB) *gorm.DB,
) ([]model.Keyword, error) {
	return ExecuteFindWithPagination[model.Keyword](ctx, r.db, limit, offset, "sort_order asc, id asc", nil, whereQueries...)
}

// CountAll ページングの総ページ数計算のため、条件に合致するキーワード情報の総件数を取得する
func (r *KeywordRepository) CountAll(ctx context.Context, whereQueries ...func(*gorm.DB) *gorm.DB) (int64, error) {
	return ExecuteCount[model.Keyword](ctx, r.db, whereQueries...)
}

// =========================================================================
// U: Update (更新)
// =========================================================================

// Update は既存のキーワード情報を更新します
func (r *KeywordRepository) Update(ctx context.Context, k *model.Keyword) error {
	return r.db.WithContext(ctx).Save(k).Error
}

// =========================================================================
// D: Delete (削除)
// =========================================================================

// Delete ジャンルIDを指定して、該当するジャンル情報を物理削除する。
// データベース側の ON DELETE CASCADE 設定により、中間テーブル（game_keywords）の紐付けデータもMySQLが自動で連動削除
func (r *KeywordRepository) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&model.Keyword{}, id).Error
}
