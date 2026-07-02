package admin

import (
	"context"
	"errors"
	"os"
	"time"
	"yutagame-backend/domain/model"
	"yutagame-backend/infrastructure/database"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AdminAuthUseCase struct {
	adminRepo *database.AdminRepository
}

func NewAdminAuthUseCase(adminRepo *database.AdminRepository) *AdminAuthUseCase {
	return &AdminAuthUseCase{adminRepo: adminRepo}
}

// 🔑 ログイン処理 (これはガードの外側で使います)
func (u *AdminAuthUseCase) Login(ctx context.Context, email, password string) (string, error) {
	admin, err := u.adminRepo.FindByEmail(ctx, email)
	if err != nil {
		return "", err
	}
	if admin == nil {
		return "", errors.New("invalid email or password")
	}

	err = bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(password))
	if err != nil {
		return "", errors.New("invalid email or password")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"admin_id": admin.ID,
		"name":     admin.Name,
		"role":     admin.RoleType,
		"exp":      time.Now().Add(time.Hour * 72).Unix(),
	})

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "yutagame-fallback-secret-key"
	}

	return token.SignedString([]byte(secret))
}

// 🛠️ 以下、ガードの内側で行う管理者CRUDロジック

func (u *AdminAuthUseCase) GetAllAdmins(ctx context.Context) ([]model.Admin, error) {
	return u.adminRepo.FindAll(ctx)
}

func (u *AdminAuthUseCase) GetAdminByID(ctx context.Context, id int64) (*model.Admin, error) {
	return u.adminRepo.FindByID(ctx, id)
}

func (u *AdminAuthUseCase) CreateAdmin(ctx context.Context, name, email, password, roleType string) (*model.Admin, error) {
	existing, err := u.adminRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("email already registered")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		return nil, err
	}

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

func (u *AdminAuthUseCase) UpdateAdmin(ctx context.Context, id int64, name, email, password, roleType string) (*model.Admin, error) {
	admin, err := u.adminRepo.FindByID(ctx, id)
	if err != nil || admin == nil {
		return nil, errors.New("admin not found")
	}

	// メールアドレス変更時の重複チェック
	if admin.Email != email {
		existing, _ := u.adminRepo.FindByEmail(ctx, email)
		if existing != nil {
			return nil, errors.New("email already in use")
		}
	}

	admin.Name = name
	admin.Email = email
	admin.RoleType = roleType

	// 💡 パスワードが入力されている場合のみ、新しくハッシュ化して更新
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

func (u *AdminAuthUseCase) DeleteAdmin(ctx context.Context, id int64) error {
	return u.adminRepo.Delete(ctx, id)
}
