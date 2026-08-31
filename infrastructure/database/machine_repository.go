package database

import (
	"context"
	"errors"
	"sort"
	"yutagame-backend/domain/model"

	"gorm.io/gorm"
)

// =========================================================================
// 構造体＆コンストラクタ
// =========================================================================

// MachineRepository 機種情報に関するデータベース操作を担当するリポジトリ
type MachineRepository struct {
	db *gorm.DB
}

type MachineWithGameCount struct {
	Code      string  `json:"code"`
	Name      string  `json:"name"`
	ImageKey  *string `json:"imageKey"`
	GameCount int64   `json:"gameCount"`
}

// NewMachineRepository MachineRepositoryの新しいインスタンスを生成するコンストラクタ
func NewMachineRepository(db *gorm.DB) *MachineRepository {
	return &MachineRepository{db: db}
}

// =========================================================================
// C: Create (作成)
// =========================================================================

// Create 新しい機種情報をデータベースに登録する
func (r *MachineRepository) Create(ctx context.Context, machine *model.Machine) error {
	return r.db.WithContext(ctx).Create(machine).Error
}

// =========================================================================
// R: Read (取得)
// =========================================================================

// FindByID 機種ID（主キー）を指定して、該当する機種情報を1件取得する
func (r *MachineRepository) FindByID(ctx context.Context, id int64) (*model.Machine, error) {
	var machine model.Machine
	err := r.db.WithContext(ctx).Preload("Manufacturer").First(&machine, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &machine, err
}

// FindByCode コードを指定して、該当する機種情報を1件取得する
func (r *MachineRepository) FindByCode(ctx context.Context, code string) (*model.Machine, error) {
	var machine model.Machine
	err := r.db.WithContext(ctx).Preload("Manufacturer").Where("code = ?", code).First(&machine).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &machine, err
}

// FindAll 登録されているすべての機種情報をID昇順で取得する（ページングなし）
func (r *MachineRepository) FindAll(ctx context.Context) ([]model.Machine, error) {
	var machines []model.Machine
	err := r.db.WithContext(ctx).Preload("Manufacturer").Order("sort_order asc, id asc").Find(&machines).Error
	if err != nil {
		return nil, err
	}
	return machines, nil
}

// FindByIDs 指定されたID群に一致する機種をまとめて取得する
func (r *MachineRepository) FindByIDs(ctx context.Context, ids []int64) ([]model.Machine, error) {
	if len(ids) == 0 {
		return []model.Machine{}, nil
	}

	var machines []model.Machine
	err := r.db.WithContext(ctx).
		Where("id IN ?", ids).
		Order("id asc").
		Find(&machines).Error
	if err != nil {
		return nil, err
	}

	sort.Slice(machines, func(i, j int) bool {
		return machines[i].ID < machines[j].ID
	})

	return machines, nil
}

// FindAllWithPagination 指定された件数（limit）と開始位置（offset）に応じて、機種情報をID昇順で取得する
func (r *MachineRepository) FindAllWithPagination(
	ctx context.Context,
	limit, offset int,
	whereQueries ...func(*gorm.DB) *gorm.DB,
) ([]model.Machine, error) {
	modifier := func(db *gorm.DB) *gorm.DB {
		return db.Preload("Manufacturer")
	}
	return ExecuteFindWithPagination[model.Machine](ctx, r.db, limit, offset, "id asc", modifier, whereQueries...)
}

// CountAll ページングの総ページ数計算のため、条件に合致する機種情報の総件数を取得する
func (r *MachineRepository) CountAll(ctx context.Context, whereQueries ...func(*gorm.DB) *gorm.DB) (int64, error) {
	return ExecuteCount[model.Machine](ctx, r.db, whereQueries...)
}

// FindAllWithGameCount 公開画面向けに機種ごとのゲーム件数を集計して返す
func (r *MachineRepository) FindAllWithGameCount(ctx context.Context) ([]MachineWithGameCount, error) {
	var items []MachineWithGameCount
	err := r.db.WithContext(ctx).
		Table("machines").
		Select(`
			machines.code,
			machines.name,
			machines.image_key,
			COUNT(games.id) AS game_count
		`).
		Joins("LEFT JOIN games ON games.machine_id = machines.id").
		Group("machines.id").
		Order("machines.sort_order asc, machines.id asc").
		Find(&items).Error
	return items, err
}

// =========================================================================
// U: Update (更新)
// =========================================================================

// Update 既存の機種情報（名前、説明など）を更新する
func (r *MachineRepository) Update(ctx context.Context, m *model.Machine) error {
	return r.db.WithContext(ctx).Save(m).Error
}

// UpdateFieldsByID 指定IDの機種に対し、指定カラムだけを更新する
func (r *MachineRepository) UpdateFieldsByID(ctx context.Context, id int64, updates map[string]any) error {
	if len(updates) == 0 {
		return nil
	}

	return r.db.WithContext(ctx).
		Model(&model.Machine{}).
		Where("id = ?", id).
		Updates(updates).Error
}

// UpdateImageKey は機種画像キーのみを更新する
func (r *MachineRepository) UpdateImageKey(ctx context.Context, id int64, imageKey *string) error {
	return r.db.WithContext(ctx).
		Model(&model.Machine{}).
		Where("id = ?", id).
		Update("image_key", imageKey).Error
}

// =========================================================================
// D: Delete (削除)
// =========================================================================

// Delete 機種IDを指定して、該当する機種情報を物理削除する
func (r *MachineRepository) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&model.Machine{}, id).Error
}
