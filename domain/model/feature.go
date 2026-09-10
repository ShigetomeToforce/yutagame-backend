package model

import "time"

type Feature struct {
	ID                int64         `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Code              string        `gorm:"column:code;not null;unique" json:"code"`
	Title             string        `gorm:"column:title;not null" json:"title"`
	Excerpt           string        `gorm:"column:excerpt;not null" json:"excerpt"`
	BodyHTML          string        `gorm:"column:body_html;not null" json:"bodyHtml"`
	ThumbnailImageKey *string       `gorm:"column:thumbnail_image_key" json:"thumbnailImageKey,omitempty"`
	Status            string        `gorm:"column:status;not null" json:"status"`
	PublishedAt       *time.Time    `gorm:"column:published_at" json:"publishedAt,omitempty"`
	PublishStartAt    *time.Time    `gorm:"column:publish_start_at" json:"publishStartAt,omitempty"`
	PublishEndAt      *time.Time    `gorm:"column:publish_end_at" json:"publishEndAt,omitempty"`
	DisplayOrder      int           `gorm:"column:display_order;not null;default:0" json:"displayOrder"`
	AccessCount       int64         `gorm:"-" json:"accessCount"`
	CreatedAt         time.Time     `gorm:"column:created_at;not null;<-:create" json:"createdAt"`
	UpdatedAt         time.Time     `gorm:"column:updated_at;not null" json:"updatedAt"`
	FeatureGames      []FeatureGame `gorm:"foreignKey:FeatureID;constraint:OnDelete:CASCADE;" json:"featureGames,omitempty"`
	Games             []Game        `gorm:"-" json:"games,omitempty"`
}

type FeatureGame struct {
	ID           int64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	FeatureID    int64     `gorm:"column:feature_id;not null;index:idx_feature_games_feature_order,priority:1" json:"featureId"`
	GameID       int64     `gorm:"column:game_id;not null;index" json:"gameId"`
	DisplayOrder int       `gorm:"column:display_order;not null;index:idx_feature_games_feature_order,priority:2" json:"displayOrder"`
	CreatedAt    time.Time `gorm:"column:created_at;not null;<-:create" json:"createdAt"`
	UpdatedAt    time.Time `gorm:"column:updated_at;not null" json:"updatedAt"`
	Game         *Game     `gorm:"-:migration;foreignKey:GameID;constraint:OnDelete:CASCADE;" json:"game,omitempty"`
}
