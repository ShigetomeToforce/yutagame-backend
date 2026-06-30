package database

import (
	"context"
	"errors"
	"yutagame-backend/domain/model"

	"gorm.io/gorm"
)

type MachineRepository struct {
	db *gorm.DB
}

func NewMachineRepository(db *gorm.DB) *MachineRepository {
	return &MachineRepository{db: db}
}

// FindAll は機種一覧を取得します
func (r *MachineRepository) FindAll(ctx context.Context) ([]model.Machine, error) {
	var machines []model.Machine
	// 💡 生SQLは一切なし！GORMが自動的に「SELECT * FROM machines ORDER BY sort_order ASC」を発行します
	err := r.db.WithContext(ctx).Preload("Manufacturer").Order("sort_order asc, id asc").Find(&machines).Error
	if err != nil {
		return nil, err
	}
	return machines, nil
}

// FindByID は指定されたIDの機種を1件取得します
func (r *MachineRepository) FindByID(ctx context.Context, id int64) (*model.Machine, error) {
	var machine model.Machine
	err := r.db.WithContext(ctx).Preload("Manufacturer").First(&machine, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil // レコードが見つからない場合はnilを返す
	}
	if err != nil {
		return nil, err
	}
	return &machine, nil
}

// Create は新しい機種を登録します
func (r *MachineRepository) Create(ctx context.Context, m *model.Machine) error {
	// 💡 挿入処理もこれだけ。CreatedAt, UpdatedAtの初期化やLastInsertIdの取得もGORMが裏で自動でやります
	return r.db.WithContext(ctx).Create(m).Error
}

// Update は既存の機種情報を更新します
func (r *MachineRepository) Update(ctx context.Context, m *model.Machine) error {
	// 💡 構造体のIDを基準に、全カラムを自動でUPDATEします
	return r.db.WithContext(ctx).Save(m).Error
}

// Delete は機種を削除します
func (r *MachineRepository) Delete(ctx context.Context, id int64) error {
	// 💡 IDを指定して削除
	return r.db.WithContext(ctx).Delete(&model.Machine{}, id).Error
}
