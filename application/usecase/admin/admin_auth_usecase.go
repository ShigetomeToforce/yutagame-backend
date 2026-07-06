package admin

import (
	"context"
	"errors"
	"os"
	"time"
	"yutagame-backend/application/usecase"
	"yutagame-backend/domain/model"
	"yutagame-backend/infrastructure/database"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// AdminAuthUseCase 管理者の認証処理およびアカウント管理のビジネスロジックを担当するユースケース
type AdminAuthUseCase struct {
	adminRepo *database.AdminRepository
}

// NewAdminAuthUseCase AdminAuthUseCaseの新しいインスタンスを生成するコンストラクタ
func NewAdminAuthUseCase(adminRepo *database.AdminRepository) *AdminAuthUseCase {
	return &AdminAuthUseCase{adminRepo: adminRepo}
}

// =========================================================================
// 🔑 Authentication (認証処理) - ルーティングのガード外側で利用
// =========================================================================

// Login メールアドレスとパスワードを検証し、認証成功時にJWTトークンを発行する
func (u *AdminAuthUseCase) Login(ctx context.Context, email, password string) (string, error) {
	// 1. メールアドレスから管理者アカウントを特定
	admin, err := u.adminRepo.FindByEmail(ctx, email)
	if err != nil {
		return "", err
	}
	if admin == nil {
		return "", errors.New("invalid email or password")
	}

	// 2. パスワードのハッシュ値を比較検証
	err = bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(password))
	if err != nil {
		return "", errors.New("invalid email or password")
	}

	// 3. JWTトークンのクレーム（格納データ）を定義
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"admin_id": admin.ID,
		"name":     admin.Name,
		"role":     admin.RoleType,
		"exp":      time.Now().Add(time.Hour * 72).Unix(), // 有効期限: 72時間
	})

	// 4. 環境変数からシークレットキーを取得して署名
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "yutagame-fallback-secret-key"
	}

	return token.SignedString([]byte(secret))
}

// =========================================================================
// 🛠️ Admin Management CRUD (管理者管理ロジック) - ルーティングのガード内側で利用
// =========================================================================

// -------------------------------------------------------------------------
// C: Create (作成)
// -------------------------------------------------------------------------

// CreateAdmin 重複チェックとパスワードのハッシュ化を行い、新しい管理者アカウントを作成する
func (u *AdminAuthUseCase) CreateAdmin(ctx context.Context, name, email, password, roleType string) (*model.Admin, error) {
	// 1. メールアドレスの重複チェック
	existing, err := u.adminRepo.FindByEmail(ctx, email)
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
	admin := &model.Admin{
		Name:     name,
		Email:    email,
		Password: string(hashedPassword),
		RoleType: roleType,
	}

	if err := u.adminRepo.Create(ctx, admin); err != nil {
		return nil, err
	}
	return admin, nil
}

// -------------------------------------------------------------------------
// R: Read (取得)
// -------------------------------------------------------------------------

// GetAdminByID 管理者IDを指定して、該当する管理者情報を1件取得する
func (u *AdminAuthUseCase) GetAdminByID(ctx context.Context, id int64) (*model.Admin, error) {
	return u.adminRepo.FindByID(ctx, id)
}

// GetAllAdmins 登録されているすべての管理者情報を取得する（ページングなしの全件マスターデータ用）
func (u *AdminAuthUseCase) GetAllAdmins(ctx context.Context) ([]model.Admin, error) {
	return u.adminRepo.FindAll(ctx)
}

// GetAdminsWithPagination 指定されたページと件数に基づいて、ページング適用済みの管理者一覧およびメタデータを取得する
func (u *AdminAuthUseCase) GetAdminsWithPagination(ctx context.Context, page, limit int) ([]model.Admin, int64, int, error) {
	// 1. 全件数の取得
	totalCount, err := u.adminRepo.CountAll(ctx)
	if err != nil {
		return nil, 0, 0, err
	}

	// 2. 💡 共通関数を使ってページングの計算とバリデーションを一撃で行う
	p := usecase.CalculatePagination(totalCount, page, limit)

	// 3. 該当ページのデータ取得（計算された安全な Offset と Limit を適用）
	admins, err := u.adminRepo.FindAllWithPagination(ctx, limit, p.Offset)
	if err != nil {
		return nil, 0, 0, err
	}

	return admins, totalCount, p.TotalPages, nil
}

// -------------------------------------------------------------------------
// U: Update (更新)
// -------------------------------------------------------------------------

// UpdateAdmin 既存の管理者情報を更新する（パスワードが空文字の場合は変更なしとして扱う）
func (u *AdminAuthUseCase) UpdateAdmin(ctx context.Context, id int64, name, email, password, roleType string) (*model.Admin, error) {
	// 1. 更新対象のアカウントが存在するか確認
	admin, err := u.adminRepo.FindByID(ctx, id)
	if err != nil || admin == nil {
		return nil, errors.New("admin not found")
	}

	// 2. メールアドレスが変更される場合のみ、他アカウントとの重複チェック
	if admin.Email != email {
		existing, _ := u.adminRepo.FindByEmail(ctx, email)
		if existing != nil {
			return nil, errors.New("email already in use")
		}
	}

	// 3. フィールドの書き換え
	admin.Name = name
	admin.Email = email
	admin.RoleType = roleType

	// 4. パスワードが入力されている場合のみ、新しくハッシュ化して更新
	if password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 10)
		if err != nil {
			return nil, err
		}
		admin.Password = string(hashedPassword)
	}

	if err := u.adminRepo.Update(ctx, admin); err != nil {
		return nil, err
	}
	return admin, nil
}

// -------------------------------------------------------------------------
// D: Delete (削除)
// -------------------------------------------------------------------------

// DeleteAdmin 管理者IDを指定して、該当する管理者アカウントを削除する
func (u *AdminAuthUseCase) DeleteAdmin(ctx context.Context, id int64) error {
	return u.adminRepo.Delete(ctx, id)
}
