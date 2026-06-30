package model

import "time"

// User は users テーブルを表すモデルです
type User struct {
	ID        int64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Email     string    `gorm:"column:email;not null;unique" json:"email"`
	Password  string    `gorm:"column:password;not null" json:"-"` // 💡 JSON変換時は非表示
	Name      string    `gorm:"column:name;not null" json:"name"`
	CreatedAt time.Time `gorm:"column:created_at;not null" json:"createdAt"`
	UpdatedAt time.Time `gorm:"column:updated_at;not null" json:"updatedAt"`
}
