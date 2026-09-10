package model

import "time"

type GameViewLog struct {
	ID        int64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	VisitorID string    `gorm:"column:visitor_id;type:varchar(128);not null;default:'';index" json:"visitorId"`
	IPHash    string    `gorm:"column:ip_hash;type:char(64);not null;default:'';index" json:"ipHash"`
	GameCode  string    `gorm:"column:game_code;type:varchar(64);not null;default:'';index" json:"gameCode"`
	CreatedAt time.Time `gorm:"column:created_at;not null;<-:create;index" json:"createdAt"`
}

func (GameViewLog) TableName() string {
	return "game_view_logs"
}
