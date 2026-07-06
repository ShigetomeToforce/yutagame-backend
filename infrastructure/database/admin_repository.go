package database

import (
	"context"
	"errors"
	"yutagame-backend/domain/model"

	"gorm.io/gorm"
)

// AdminRepository 管理者データに関するデータベース操作を担当するリポジトリ
type AdminRepository struct {
	db *gorm.DB
}

// NewAdminRepository AdminRepositoryの新しいインスタンスを生成するコンストラクタ
func NewAdminRepository(db *gorm.DB) *AdminRepository {
	return &AdminRepository{db: db}
}

// =========================================================================
// C: Create (作成)
// =========================================================================

// Create 新しい管理者アカウントをデータベースに登録する
func (r *AdminRepository) Create(ctx context.Context, admin *model.Admin) error {
	return r.db.WithContext(ctx).Create(admin).Error
}

// =========================================================================
// R: Read (取得)
// =========================================================================

// FindByID 管理者ID（主キー）を指定して、該当する管理者情報を1件取得する
func (r *AdminRepository) FindByID(ctx context.Context, id int64) (*model.Admin, error) {
	var admin model.Admin
	err := r.db.WithContext(ctx).First(&admin, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil // レコードが見つからない場合はエラーにせずnilを返す
	}
	return &admin, err
}

// FindByEmail メールアドレスを指定して、該当する管理者情報を1件取得する（ログイン・重複チェック用）
func (r *AdminRepository) FindByEmail(ctx context.Context, email string) (*model.Admin, error) {
	var admin model.Admin
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&admin).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil // レコードが見つからない場合はエラーにせずnilを返す
	}
	return &admin, err
}

// FindAll 登録されているすべての管理者情報をID昇順で取得する（ページングなし）
func (r *AdminRepository) FindAll(ctx context.Context) ([]model.Admin, error) {
	var admins []model.Admin
	err := r.db.WithContext(ctx).Order("id asc").Find(&admins).Error
	return admins, err
}

// FindAllWithPagination 指定された件数（limit）と開始位置（offset）に応じて、管理者情報をID昇順で取得する
func (r *AdminRepository) FindAllWithPagination(ctx context.Context, limit, offset int) ([]model.Admin, error) {
	var admins []model.Admin
	err := r.db.WithContext(ctx).Order("id asc").Limit(limit).Offset(offset).Find(&admins).Error
	return admins, err
}

// CountAll ページングの総ページ数計算のため、登録されているすべての管理者アカウントの総件数を取得する
func (r *AdminRepository) CountAll(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Admin{}).Count(&count).Error
	return count, err
}

// =========================================================================
// U: Update (更新)
// =========================================================================

// Update 既存の管理者情報（名前、メール、パスワード、権限など）を更新する
func (r *AdminRepository) Update(ctx context.Context, admin *model.Admin) error {
	return r.db.WithContext(ctx).Save(admin).Error
}

// =========================================================================
// D: Delete (削除)
// =========================================================================

// Delete 管理者IDを指定して、該当する管理者アカウントを物理削除する
func (r *AdminRepository) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&model.Admin{}, id).Error
}
