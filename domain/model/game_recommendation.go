package model

import "time"

type GameRecommendation struct {
	ID        int64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	GameName  string    `gorm:"column:game_name;not null" json:"gameName"`
	Reason    string    `gorm:"column:reason;not null" json:"reason"`
	Status    string    `gorm:"column:status;not null" json:"status"`
	AdminNote string    `gorm:"column:admin_note;not null" json:"adminNote"`
	CreatedAt time.Time `gorm:"column:created_at;not null;<-:create" json:"createdAt"`
	UpdatedAt time.Time `gorm:"column:updated_at;not null" json:"updatedAt"`
}
