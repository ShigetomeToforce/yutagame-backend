package model

import "time"

type SearchLog struct {
	ID               int64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	VisitorID        string    `gorm:"column:visitor_id;type:varchar(128);not null;default:'';index" json:"visitorId"`
	IPHash           string    `gorm:"column:ip_hash;type:char(64);not null;default:'';index" json:"ipHash"`
	MachineCode      string    `gorm:"column:machine_code;type:varchar(64);not null;default:'';index" json:"machineCode"`
	ManufacturerCode string    `gorm:"column:manufacturer_code;type:varchar(64);not null;default:'';index" json:"manufacturerCode"`
	GenreCode        string    `gorm:"column:genre_code;type:varchar(64);not null;default:'';index" json:"genreCode"`
	KeywordCode      string    `gorm:"column:keyword_code;type:varchar(64);not null;default:'';index" json:"keywordCode"`
	SearchWord       string    `gorm:"column:search_word;type:varchar(255);not null;default:'';index" json:"searchWord"`
	CreatedAt        time.Time `gorm:"column:created_at;not null;<-:create;index" json:"createdAt"`
}

func (SearchLog) TableName() string {
	return "search_logs"
}
