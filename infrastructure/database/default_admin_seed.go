package database

import (
	"os"
	"strings"
	"time"
	"yutagame-backend/domain/model"

	"gorm.io/gorm"
)

const defaultAdminPasswordHash = "$2a$10$53xmr3o2m1Tuxo0IQFfko.Suq4Z426P16q4eStnZ9u4acD0WlyeAq"

func envOrDefault(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func EnsureDefaultAdmin(db *gorm.DB) error {
	admin := model.Admin{
		ID:        1,
		Email:     envOrDefault("DEFAULT_ADMIN_EMAIL", "admin@example.com"),
		Password:  envOrDefault("DEFAULT_ADMIN_PASSWORD_HASH", defaultAdminPasswordHash),
		Name:      envOrDefault("DEFAULT_ADMIN_NAME", "デフォルト管理者アカウント"),
		RoleType:  envOrDefault("DEFAULT_ADMIN_ROLE", "ADMIN"),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	return db.Where("id = ?", admin.ID).Attrs(admin).FirstOrCreate(&model.Admin{}).Error
}
