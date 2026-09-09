package model

import "time"

// GameAffiliate はゲームの購入導線URLを管理するモデルです。
type GameAffiliate struct {
	ID        int64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	GameID    int64     `gorm:"column:game_id;not null" json:"gameId"`
	Category  string    `gorm:"column:category;not null" json:"category"`
	URL       string    `gorm:"column:url;not null" json:"url"`
	CreatedAt time.Time `gorm:"column:created_at;not null;<-:create" json:"createdAt"`
	UpdatedAt time.Time `gorm:"column:updated_at;not null" json:"updatedAt"`
}
