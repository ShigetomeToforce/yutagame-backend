package model

import "time"

// Keyword は keywords テーブルを表すモデルです
type Keyword struct {
	ID        int64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name      string    `gorm:"column:name;not null" json:"name"`
	Kana      string    `gorm:"column:kana;not null" json:"kana"`
	Code      string    `gorm:"column:code;not null;unique" json:"code"`
	Category  string    `gorm:"column:category;not null" json:"category"`
	SortOrder int32     `gorm:"column:sort_order;not null" json:"sortOrder"`
	CreatedAt time.Time `gorm:"column:created_at;not null;<-:create" json:"createdAt"`
	UpdatedAt time.Time `gorm:"column:updated_at;not null" json:"updatedAt"`
}
