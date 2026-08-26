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

// UserRepository ユーザーデータに関するデータベース操作を担当するリポジトリ
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository UserRepositoryの新しいインスタンスを生成するコンストラクタ
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// =========================================================================
// C: Create (作成)
// =========================================================================

// Create 新しいユーザーアカウントをデータベースに登録する
func (r *UserRepository) Create(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

// =========================================================================
// R: Read (取得)
// =========================================================================

// FindByID ユーザーID（主キー）を指定して、該当するユーザー情報を1件取得する
func (r *UserRepository) FindByID(ctx context.Context, id int64) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).First(&user, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &user, err
}

// FindByEmail メールアドレスを指定して、該当するユーザー情報を1件取得する（ログイン・重複チェック用）
func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &user, err
}

// FindAll 登録されているすべてのユーザー情報をID昇順で取得する（ページングなし）
func (r *UserRepository) FindAll(ctx context.Context) ([]model.User, error) {
	var users []model.User
	err := r.db.WithContext(ctx).Order("id asc").Find(&users).Error
	return users, err
}

// FindAllWithPagination 指定された件数（limit）と開始位置（offset）に応じて、ユーザー情報をID昇順で取得する
func (r *UserRepository) FindAllWithPagination(
	ctx context.Context,
	limit, offset int,
	whereQueries ...func(*gorm.DB) *gorm.DB,
) ([]model.User, error) {
	return ExecuteFindWithPagination[model.User](ctx, r.db, limit, offset, "id asc", nil, whereQueries...)
}

// CountAll ページングの総ページ数計算のため、条件に合致するユーザー情報の総件数を取得する
func (r *UserRepository) CountAll(ctx context.Context, whereQueries ...func(*gorm.DB) *gorm.DB) (int64, error) {
	return ExecuteCount[model.User](ctx, r.db, whereQueries...)
}

// =========================================================================
// U: Update (更新)
// =========================================================================

// Update 既存の管理アカウント情報（メール、パスワードなど）を更新する
func (r *UserRepository) Update(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

// =========================================================================
// D: Delete (削除)
// =========================================================================

// Delete ユーザーIDを指定して、該当するユーザー情報を物理削除する
func (r *UserRepository) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&model.User{}, id).Error
}
