package model

import "time"

// GameFavorite は推し投票の記録を保持するモデルです。
type GameFavorite struct {
	ID           int64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	GameID       int64     `gorm:"column:game_id;not null;index:idx_game_favorites_game_id;uniqueIndex:ux_game_favorites_game_visitor_date,priority:1" json:"gameId"`
	VisitorID    string    `gorm:"column:visitor_id;not null;size:64;index:idx_game_favorites_visitor_id;uniqueIndex:ux_game_favorites_game_visitor_date,priority:2" json:"visitorId"`
	FavoriteDate string    `gorm:"column:favorite_date;type:date;not null;index:idx_game_favorites_favorite_date;uniqueIndex:ux_game_favorites_game_visitor_date,priority:3" json:"favoriteDate"`
	CreatedAt    time.Time `gorm:"column:created_at;not null;<-:create" json:"createdAt"`
	UpdatedAt    time.Time `gorm:"column:updated_at;not null" json:"updatedAt"`

	Game *Game `gorm:"-:migration;foreignKey:GameID;constraint:OnDelete:CASCADE;" json:"game,omitempty"`
}
