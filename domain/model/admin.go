package model

import "time"

// Admin は admins テーブルを表すモデルです
type Admin struct {
	ID        int64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Email     string    `gorm:"column:email;not null;unique" json:"email"`
	Password  string    `gorm:"column:password;not null" json:"-"` // 💡 セキュリティのため、JSON変換時はパスワードを隠蔽します
	Name      string    `gorm:"column:name;not null" json:"name"`
	RoleType  string    `gorm:"column:role_type;not null" json:"roleType"`
	CreatedAt time.Time `gorm:"column:created_at;not null" json:"createdAt"`
	UpdatedAt time.Time `gorm:"column:updated_at;not null" json:"updatedAt"`
}
