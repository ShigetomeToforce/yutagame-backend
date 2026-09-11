package model

import "time"

// PageViewLog は同じ訪問者・同じページを1日1行だけ保持するアクセス集計用モデルです。
type PageViewLog struct {
	ID          int64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	VisitorHash string    `gorm:"column:visitor_hash;type:char(64);not null;uniqueIndex:uq_page_view_daily,priority:1" json:"-"`
	PagePath    string    `gorm:"column:page_path;type:varchar(191);not null;uniqueIndex:uq_page_view_daily,priority:2;index" json:"pagePath"`
	ViewedOn    string    `gorm:"column:viewed_on;type:date;not null;uniqueIndex:uq_page_view_daily,priority:3;index" json:"viewedOn"`
	CreatedAt   time.Time `gorm:"column:created_at;not null;<-:create;index" json:"createdAt"`
}

func (PageViewLog) TableName() string {
	return "page_view_logs"
}
