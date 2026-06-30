package model

import "time"

// Manufacturer は manufacturers テーブルを表すモデルです
type Manufacturer struct {
	ID        int64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name      string    `gorm:"column:name;not null" json:"name"`
	Kana      string    `gorm:"column:kana;not null" json:"kana"`
	Overview  string    `gorm:"column:overview;not null" json:"overview"`
	Code      string    `gorm:"column:code;not null;unique" json:"code"`
	CreatedAt time.Time `gorm:"column:created_at;not null" json:"createdAt"`
	UpdatedAt time.Time `gorm:"column:updated_at;not null" json:"updatedAt"`
}
