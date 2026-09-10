package model

import "time"

type Announcement struct {
	ID             int64      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Title          string     `gorm:"column:title;not null" json:"title"`
	Excerpt        string     `gorm:"column:excerpt;not null" json:"excerpt"`
	BodyHTML       string     `gorm:"column:body_html;not null" json:"bodyHtml"`
	Status         string     `gorm:"column:status;not null" json:"status"`
	PublishedAt    *time.Time `gorm:"column:published_at" json:"publishedAt,omitempty"`
	PublishStartAt *time.Time `gorm:"column:publish_start_at" json:"publishStartAt,omitempty"`
	PublishEndAt   *time.Time `gorm:"column:publish_end_at" json:"publishEndAt,omitempty"`
	// DisplayOrder は公開中のお知らせ同士の表示優先度（1以上が手動指定、0は未指定）
	DisplayOrder int       `gorm:"column:display_order;not null;default:0" json:"displayOrder"`
	AccessCount  int64     `gorm:"-" json:"accessCount"`
	CreatedAt    time.Time `gorm:"column:created_at;not null;<-:create" json:"createdAt"`
	UpdatedAt    time.Time `gorm:"column:updated_at;not null" json:"updatedAt"`
}
