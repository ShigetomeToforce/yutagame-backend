package admin

import (
	"context"
	"errors"
	"yutagame-backend/application/usecase"
	"yutagame-backend/domain/model"
	"yutagame-backend/infrastructure/database"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// =========================================================================
// 構造体＆コンストラクタ
// =========================================================================

// UserListFilter ユーザー検索用構造体
type UserListFilter struct {
	SearchWord string
}

// UserUseCase ユーザーの認証処理およびアカウント管理のビジネスロジックを担当するユースケース
type UserUseCase struct {
	userRepo *database.UserRepository
}

// NewUserUseCase UserUseCaseの新しいインスタンスを生成するコンストラクタ
func NewUserUseCase(userRepo *database.UserRepository) *UserUseCase {
	return &UserUseCase{userRepo: userRepo}
}

// =========================================================================
// 🛠️ User Management CRUD (管理者管理ロジック) - ルーティングのガード内側で利用
// =========================================================================

// -------------------------------------------------------------------------
// C: Create (作成)
// -------------------------------------------------------------------------

// CreateUser 重複チェックとパスワードのハッシュ化を行い、新しいユーザー情報を作成する
func (u *UserUseCase) CreateUser(ctx context.Context, name, email, password string) (*model.User, error) {
	// 1. メールアドレスの重複チェック
	existing, err := u.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("email already registered")
	}

	// 2. パスワードを安全にハッシュ化
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		return nil, err
	}

	// 3. モデルの組み立てと永続化
	user := &model.User{
		Name:     name,
		Email:    email,
		Password: string(hashedPassword),
	}

	if err := u.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

// -------------------------------------------------------------------------
// R: Read (取得)
// -------------------------------------------------------------------------

// GetUserByID 管理者IDを指定して、該当するユーザー情報を1件取得する
func (u *UserUseCase) GetUserByID(ctx context.Context, id int64) (*model.User, error) {
	return u.userRepo.FindByID(ctx, id)
}

// GetAllUsers 登録されているすべてのユーザー情報を取得する（ページングなしの全件マスターデータ用）
func (u *UserUseCase) GetAllUsers(ctx context.Context) ([]model.User, error) {
	return u.userRepo.FindAll(ctx)
}

// GetUsersWithPagination 指定されたページ、件数、検索キーワードに基づいて、ページング・検索適用済みのユーザー情報を取得する
func (u *UserUseCase) GetUsersWithPagination(
	ctx context.Context,
	page, limit int,
	filter UserListFilter,
) ([]model.User, int64, int, error) {
	var whereQuery func(*gorm.DB) *gorm.DB
	if filter.SearchWord != "" {
		whereQuery = func(db *gorm.DB) *gorm.DB {
			likeQuery := "%" + filter.SearchWord + "%"
			return db.Where("name LIKE ? OR email LIKE ?", likeQuery, likeQuery)
		}
	}

	return usecase.ExecutePaginatedSearch(
		ctx, page, limit, whereQuery,
		u.userRepo.CountAll,
		u.userRepo.FindAllWithPagination,
	)
}

// -------------------------------------------------------------------------
// U: Update (更新)
// -------------------------------------------------------------------------

// UpdateUser 既存の管理者情報を更新する（パスワードが空文字の場合は変更なしとして扱う）
func (u *UserUseCase) UpdateUser(ctx context.Context, id int64, name, email, password string) (*model.User, error) {
	// 1. 更新対象のアカウントが存在するか確認
	user, err := u.userRepo.FindByID(ctx, id)
	if err != nil || user == nil {
		return nil, errors.New("user not found")
	}

	// 2. メールアドレスが変更される場合のみ、他アカウントとの重複チェック
	if user.Email != email {
		existing, _ := u.userRepo.FindByEmail(ctx, email)
		if existing != nil {
			return nil, errors.New("email already in use")
		}
	}

	// 3. フィールドの書き換え
	user.Name = name
	user.Email = email

	// 4. パスワードが入力されている場合のみ、新しくハッシュ化して更新
	if password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 10)
		if err != nil {
			return nil, err
		}
		user.Password = string(hashedPassword)
	}

	if err := u.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

// -------------------------------------------------------------------------
// D: Delete (削除)
// -------------------------------------------------------------------------

// DeleteUser ユーザーIDを指定して、該当するユーザーアカウントを削除する
func (u *UserUseCase) DeleteUser(ctx context.Context, id int64) error {
	return u.userRepo.Delete(ctx, id)
}
