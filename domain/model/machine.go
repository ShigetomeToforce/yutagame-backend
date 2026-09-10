package model

import "time"

// Machine は machines テーブルを表すモデルです
type Machine struct {
	ID             int64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name           string    `gorm:"column:name;type:varchar(256);not null" json:"name"`
	Kana           string    `gorm:"column:kana;type:varchar(256);not null" json:"kana"`
	Overview       string    `gorm:"column:overview;type:varchar(2000);not null" json:"overview"`
	Code           string    `gorm:"column:code;type:varchar(30);not null;unique" json:"code"`
	ImageKey       *string   `gorm:"column:image_key;type:varchar(255)" json:"imageKey"`
	Abbreviation   string    `gorm:"column:abbreviation;type:varchar(10);not null" json:"abbreviation"`
	ManufacturerID int64     `gorm:"column:manufacturer_id;not null" json:"manufacturerId"`
	MachineType    string    `gorm:"column:machine_type;type:enum('STATIONARY','PORTABLE','BOTH');not null" json:"machineType"`
	ReleaseDate    time.Time `gorm:"column:release_date;not null" json:"releaseDate"`
	SortOrder      int32     `gorm:"column:sort_order;not null" json:"sortOrder"`
	CreatedAt      time.Time `gorm:"column:created_at;not null;<-:create" json:"createdAt"`
	UpdatedAt      time.Time `gorm:"column:updated_at;not null" json:"updatedAt"`

	// 💡 依存関係の解消：GORMに「メーカー情報も連動してね」と教えるリレーション定義
	Manufacturer *Manufacturer `gorm:"-:migration;foreignKey:ManufacturerID;constraint:OnDelete:RESTRICT;" json:"manufacturer,omitempty"`
}
