package model

import "time"

// Keyword は keywords テーブルを表すモデルです
type Keyword struct {
	ID          int64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name        string    `gorm:"column:name;type:varchar(256);not null" json:"name"`
	Kana        string    `gorm:"column:kana;type:varchar(256);not null" json:"kana"`
	Overview    string    `gorm:"column:overview;type:varchar(2000);not null" json:"overview"`
	Code        string    `gorm:"column:code;type:varchar(30);not null;unique" json:"code"`
	KeywordType string    `gorm:"column:keyword_type;type:enum('SERIES','SYSTEM','MACHINE','OTHER');not null" json:"keywordType"`
	SortOrder   int32     `gorm:"column:sort_order;not null" json:"sortOrder"`
	CreatedAt   time.Time `gorm:"column:created_at;not null;<-:create" json:"createdAt"`
	UpdatedAt   time.Time `gorm:"column:updated_at;not null" json:"updatedAt"`
}
