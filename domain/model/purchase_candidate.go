package model

import "time"

type PurchaseCandidate struct {
	ID              int64         `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name            string        `gorm:"column:name;not null" json:"name"`
	Kana            string        `gorm:"column:kana;not null" json:"kana"`
	ImageKey        *string       `gorm:"column:image_key" json:"imageKey,omitempty"`
	Code            string        `gorm:"column:code;not null;unique" json:"code"`
	ListPrice       *int32        `gorm:"column:list_price" json:"listPrice,omitempty"`
	OfficialSiteURL string        `gorm:"column:official_site_url;not null;default:''" json:"officialSiteUrl"`
	YouTubeURL      string        `gorm:"column:youtube_url;not null;default:''" json:"youtubeUrl"`
	ReleaseDateText string        `gorm:"column:release_date_text;not null;default:''" json:"releaseDateText"`
	ManufacturerID  int64         `gorm:"column:manufacturer_id;not null" json:"manufacturerId"`
	MachineID       int64         `gorm:"column:machine_id;not null" json:"machineId"`
	GenreID         *int64        `gorm:"column:genre_id" json:"genreId,omitempty"`
	IsPurchased     bool          `gorm:"column:is_purchased;not null;default:false;index" json:"isPurchased"`
	DisplayOrder    int           `gorm:"column:display_order;not null;default:0;index" json:"displayOrder"`
	CreatedAt       time.Time     `gorm:"column:created_at;not null;<-:create" json:"createdAt"`
	UpdatedAt       time.Time     `gorm:"column:updated_at;not null" json:"updatedAt"`
	Manufacturer    *Manufacturer `gorm:"-:migration;foreignKey:ManufacturerID;" json:"manufacturer,omitempty"`
	Machine         *Machine      `gorm:"-:migration;foreignKey:MachineID;" json:"machine,omitempty"`
	Genre           *Genre        `gorm:"-:migration;foreignKey:GenreID;" json:"genre,omitempty"`
}

func (PurchaseCandidate) TableName() string {
	return "purchase_candidates"
}
