package model

import "time"

// Machine は machines テーブルを表すモデルです
type Machine struct {
	ID             int64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name           string    `gorm:"column:name;not null" json:"name"`
	Kana           string    `gorm:"column:kana;not null" json:"kana"`
	Overview       string    `gorm:"column:overview;not null" json:"overview"`
	Code           string    `gorm:"column:code;not null;unique" json:"code"`
	Abbreviation   string    `gorm:"column:abbreviation;not null" json:"abbreviation"`
	ManufacturerID int64     `gorm:"column:manufacturer_id;not null" json:"manufacturerId"`
	MachineType    string    `gorm:"column:machine_type;not null" json:"machineType"`
	ReleaseDate    time.Time `gorm:"column:release_date;not null" json:"releaseDate"`
	SortOrder      int32     `gorm:"column:sort_order;not null" json:"sortOrder"`
	CreatedAt      time.Time `gorm:"column:created_at;not null;<-:create" json:"createdAt"`
	UpdatedAt      time.Time `gorm:"column:updated_at;not null" json:"updatedAt"`

	// 💡 依存関係の解消：GORMに「メーカー情報も連動してね」と教えるリレーション定義
	Manufacturer *Manufacturer `gorm:"foreignKey:ManufacturerID;constraint:OnDelete:RESTRICT;" json:"manufacturer,omitempty"`
}
