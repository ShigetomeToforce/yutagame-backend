package model

import "time"

type Banner struct {
	ID           int64      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Title        string     `gorm:"column:title;not null" json:"title"`
	Placement    string     `gorm:"column:placement;not null;index" json:"placement"`
	ImageKey     string     `gorm:"column:image_key;not null" json:"imageKey"`
	LinkURL      string     `gorm:"column:link_url" json:"linkUrl"`
	OpenInNewTab bool       `gorm:"column:open_in_new_tab;not null" json:"openInNewTab"`
	StartsAt     *time.Time `gorm:"column:starts_at" json:"startsAt,omitempty"`
	EndsAt       *time.Time `gorm:"column:ends_at" json:"endsAt,omitempty"`
	DisplayOrder int        `gorm:"column:display_order;not null;index" json:"displayOrder"`
	ClickCount   int64      `gorm:"column:click_count;not null" json:"clickCount"`
	AccessCount  int64      `gorm:"-" json:"accessCount"`
	CreatedAt    time.Time  `gorm:"column:created_at;not null;<-:create" json:"createdAt"`
	UpdatedAt    time.Time  `gorm:"column:updated_at;not null" json:"updatedAt"`
}
