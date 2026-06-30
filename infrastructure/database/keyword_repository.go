package database

import (
	"context"
	"errors"
	"yutagame-backend/domain/model"

	"gorm.io/gorm"
)

type KeywordRepository struct {
	db *gorm.DB
}

func NewKeywordRepository(db *gorm.DB) *KeywordRepository {
	return &KeywordRepository{db: db}
}

// FindAll はすべてのキーワードを取得します（ソート順）
func (r *KeywordRepository) FindAll(ctx context.Context) ([]model.Keyword, error) {
	var keywords []model.Keyword
	err := r.db.WithContext(ctx).Order("sort_order asc, id asc").Find(&keywords).Error
	if err != nil {
		return nil, err
	}
	return keywords, nil
}

// FindByID は指定されたIDのキーワードを1件取得します
func (r *KeywordRepository) FindByID(ctx context.Context, id int64) (*model.Keyword, error) {
	var keyword model.Keyword
	err := r.db.WithContext(ctx).First(&keyword, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &keyword, nil
}

// Create は新しいキーワードを登録します
func (r *KeywordRepository) Create(ctx context.Context, k *model.Keyword) error {
	return r.db.WithContext(ctx).Create(k).Error
}

// Update は既存のキーワード情報を更新します
func (r *KeywordRepository) Update(ctx context.Context, k *model.Keyword) error {
	return r.db.WithContext(ctx).Save(k).Error
}

// Delete はキーワードを削除します
// 💡 データベース側の ON DELETE CASCADE 設定により、
// 中間テーブル（game_keywords）の紐付けデータもMySQLが自動で連動削除してくれます！
func (r *KeywordRepository) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&model.Keyword{}, id).Error
}
