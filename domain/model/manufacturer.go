package model

import "time"

// Manufacturer は manufacturers テーブルを表すモデルです
type Manufacturer struct {
	ID        int64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name      string    `gorm:"column:name;type:varchar(256);not null" json:"name"`
	Kana      string    `gorm:"column:kana;type:varchar(256);not null" json:"kana"`
	Overview  string    `gorm:"column:overview;type:varchar(2000);not null" json:"overview"`
	Code      string    `gorm:"column:code;type:varchar(30);not null;unique" json:"code"`
	ImageKey  *string   `gorm:"column:image_key;type:varchar(255)" json:"imageKey"`
	CreatedAt time.Time `gorm:"column:created_at;not null;<-:create" json:"createdAt"`
	UpdatedAt time.Time `gorm:"column:updated_at;not null" json:"updatedAt"`
}
