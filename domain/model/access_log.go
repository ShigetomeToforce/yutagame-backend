package model

import "time"

// AccessLog はアクセス解析イベントの生ログを保持するモデルです。
type AccessLog struct {
	ID                uint      `gorm:"primaryKey" json:"id"`
	EventType         string    `gorm:"type:varchar(32);index" json:"eventType"`
	EventSource       string    `gorm:"type:varchar(32);index" json:"eventSource"`
	Path              string    `gorm:"type:varchar(255);index" json:"path"`
	Method            string    `gorm:"type:varchar(16)" json:"method"`
	StatusCode        int       `gorm:"index" json:"statusCode"`
	VisitorID         string    `gorm:"type:varchar(64);index" json:"visitorId"`
	IPHash            string    `gorm:"type:varchar(64);index" json:"ipHash"`
	UserAgent         string    `gorm:"type:varchar(512)" json:"userAgent"`
	Referrer          string    `gorm:"type:varchar(512)" json:"referrer"`
	MachineCode       string    `gorm:"type:varchar(64);index" json:"machineCode"`
	ManufacturerCode  string    `gorm:"type:varchar(64);index" json:"manufacturerCode"`
	GenreCode         string    `gorm:"type:varchar(64);index" json:"genreCode"`
	KeywordCode       string    `gorm:"type:varchar(64);index" json:"keywordCode"`
	SearchWord        string    `gorm:"type:varchar(255);index" json:"searchWord"`
	GameCode          string    `gorm:"type:varchar(64);index" json:"gameCode"`
	AffiliateCategory string    `gorm:"type:varchar(64);index" json:"affiliateCategory"`
	CreatedAt         time.Time `gorm:"index" json:"createdAt"`
}
