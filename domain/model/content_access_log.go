package model

import "time"

type ContentAccessLog struct {
	ID          int64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	ContentType string    `gorm:"column:content_type;type:varchar(32);not null;index:idx_content_access_target_day,priority:1" json:"contentType"`
	ContentKey  string    `gorm:"column:content_key;type:varchar(191);not null;index:idx_content_access_target_day,priority:2" json:"contentKey"`
	VisitorID   string    `gorm:"column:visitor_id;type:varchar(128);not null;default:'';index" json:"visitorId"`
	IPHash      string    `gorm:"column:ip_hash;type:char(64);not null;default:'';index" json:"ipHash"`
	CreatedAt   time.Time `gorm:"column:created_at;not null;<-:create;index:idx_content_access_target_day,priority:3" json:"createdAt"`
}

func (ContentAccessLog) TableName() string {
	return "content_access_logs"
}
