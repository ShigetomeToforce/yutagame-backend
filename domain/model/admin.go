package model

import "time"

// Admin は admins テーブルを表すモデルです
type Admin struct {
	ID        int64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Email     string    `gorm:"column:email;type:varchar(256);not null;unique" json:"email"`
	Password  string    `gorm:"column:password;type:varchar(128);not null" json:"-"` // 💡 セキュリティのため、JSON変換時はパスワードを隠蔽します
	Name      string    `gorm:"column:name;type:varchar(32);not null" json:"name"`
	RoleType  string    `gorm:"column:role_type;type:enum('ADMIN','USER');not null" json:"roleType"`
	CreatedAt time.Time `gorm:"column:created_at;not null;<-:create" json:"createdAt"`
	UpdatedAt time.Time `gorm:"column:updated_at;not null" json:"updatedAt"`
}
