package model

import "time"

// Game は games テーブルを表すモデルです（GORM ＆ 多対多対応版）
type Game struct {
	ID              int64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name            string    `gorm:"column:name;not null" json:"name"`
	Kana            string    `gorm:"column:kana;not null" json:"kana"`
	Overview        string    `gorm:"column:overview;not null" json:"overview"`
	Code            string    `gorm:"column:code;not null;unique" json:"code"`
	ManufacturerID  int64     `gorm:"column:manufacturer_id;not null" json:"manufacturerId"`
	MachineID       int64     `gorm:"column:machine_id;not null" json:"machineId"`
	GenreID         int64     `gorm:"column:genre_id;not null" json:"genreId"`
	SubGenre        string    `gorm:"column:sub_genre;not null" json:"subGenre"`
	CatchCopy       string    `gorm:"column:catch_copy;not null" json:"catchCopy"`
	SubCatch        string    `gorm:"column:sub_catch;not null" json:"subCatch"`
	ListPrice       int32     `gorm:"column:list_price;not null" json:"listPrice"`
	ReleaseDate     time.Time `gorm:"column:release_date;not null" json:"releaseDate"`
	OfficialSiteURL string    `gorm:"column:official_site_url;not null" json:"officialSiteUrl"`
	YouTubeURL      string    `gorm:"column:youtube_url;not null" json:"youtubeUrl"`
	IsPlay          bool      `gorm:"column:is_play;not null" json:"isPlay"`
	IsClear         bool      `gorm:"column:is_clear;not null" json:"isClear"`
	IsFavourite     bool      `gorm:"column:is_favourite;not null" json:"isFavourite"`
	CreatedAt       time.Time `gorm:"column:created_at;not null;<-:create" json:"createdAt"`
	UpdatedAt       time.Time `gorm:"column:updated_at;not null" json:"updatedAt"`

	// 💡 依存関係の解消：GORMの1対多（Belongs To）リレーション群
	Manufacturer *Manufacturer `gorm:"foreignKey:ManufacturerID;" json:"manufacturer,omitempty"`
	Machine      *Machine      `gorm:"foreignKey:MachineID;" json:"machine,omitempty"`
	Genre        *Genre        `gorm:"foreignKey:GenreID;" json:"genre,omitempty"`

	// 💡 凄まじいポイント：GORMに「game_keywordsという中間テーブルを使って、紐づくKeywordsを全自動でガッチャンコしてね」と1行で命令します
	Keywords []Keyword `gorm:"many2many:game_keywords;" json:"keywords"`
}
