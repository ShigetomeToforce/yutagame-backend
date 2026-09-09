package model

import "time"

type Announcement struct {
	ID          int64      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Title       string     `gorm:"column:title;not null" json:"title"`
	Excerpt     string     `gorm:"column:excerpt;not null" json:"excerpt"`
	BodyHTML    string     `gorm:"column:body_html;not null" json:"bodyHtml"`
	Status      string     `gorm:"column:status;not null" json:"status"`
	PublishedAt *time.Time `gorm:"column:published_at" json:"publishedAt,omitempty"`
	CreatedAt   time.Time  `gorm:"column:created_at;not null;<-:create" json:"createdAt"`
	UpdatedAt   time.Time  `gorm:"column:updated_at;not null" json:"updatedAt"`
}
