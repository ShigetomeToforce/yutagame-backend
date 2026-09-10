package model

import "time"

// GameRankingActiveEntry は公開中のランキングを管理するテーブル。
type GameRankingActiveEntry struct {
	ID        int64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	GameID    int64     `gorm:"column:game_id;not null;index:idx_game_ranking_active_game,unique" json:"gameId"`
	Rank      int       `gorm:"column:display_rank;not null" json:"rank"`
	CreatedAt time.Time `gorm:"column:created_at;not null;<-:create" json:"createdAt"`
	UpdatedAt time.Time `gorm:"column:updated_at;not null" json:"updatedAt"`

	Game *Game `gorm:"foreignKey:GameID" json:"game,omitempty"`
}

func (GameRankingActiveEntry) TableName() string {
	return "game_ranking_active_entries"
}

// GameRankingDraftEntry は一時保存用のランキングを管理するテーブル。
type GameRankingDraftEntry struct {
	ID        int64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	GameID    int64     `gorm:"column:game_id;not null;index:idx_game_ranking_draft_game,unique" json:"gameId"`
	Rank      int       `gorm:"column:display_rank;not null" json:"rank"`
	CreatedAt time.Time `gorm:"column:created_at;not null;<-:create" json:"createdAt"`
	UpdatedAt time.Time `gorm:"column:updated_at;not null" json:"updatedAt"`

	Game *Game `gorm:"foreignKey:GameID" json:"game,omitempty"`
}

func (GameRankingDraftEntry) TableName() string {
	return "game_ranking_draft_entries"
}

// GameRankingPreviousEntry は直前に公開されていたランキングを保持するテーブル。
type GameRankingPreviousEntry struct {
	ID        int64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	GameID    int64     `gorm:"column:game_id;not null;index:idx_game_ranking_previous_game,unique" json:"gameId"`
	Rank      int       `gorm:"column:display_rank;not null" json:"rank"`
	CreatedAt time.Time `gorm:"column:created_at;not null;<-:create" json:"createdAt"`
	UpdatedAt time.Time `gorm:"column:updated_at;not null" json:"updatedAt"`
}

func (GameRankingPreviousEntry) TableName() string {
	return "game_ranking_previous_entries"
}
