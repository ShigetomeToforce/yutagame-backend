package model

import "time"

// User は users テーブルを表すモデルです
type User struct {
	ID        int64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Email     string    `gorm:"column:email;type:varchar(256);not null;unique" json:"email"`
	Password  string    `gorm:"column:password;type:varchar(128);not null" json:"-"` // 💡 JSON変換時は非表示
	Name      string    `gorm:"column:name;type:varchar(32);not null" json:"name"`
	CreatedAt time.Time `gorm:"column:created_at;not null;<-:create" json:"createdAt"`
	UpdatedAt time.Time `gorm:"column:updated_at;not null" json:"updatedAt"`
}
