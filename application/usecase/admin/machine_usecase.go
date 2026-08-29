package admin

import (
	"context"
	"yutagame-backend/application/usecase"
	"yutagame-backend/domain/model"
	"yutagame-backend/infrastructure/database"

	"gorm.io/gorm"
)

// =========================================================================
// 構造体＆コンストラクタ
// =========================================================================

// MachineListFilter 機種検索用構造体
type MachineListFilter struct {
	SearchWord      string
	ManufacturerIDs []int64
}

// MachineUseCase 機種管理のビジネスロジックを担当するユースケース
type MachineUseCase struct {
	machineRepo *database.MachineRepository
}

// NewMachineUseCase MachineUseCaseの新しいインスタンスを生成するコンストラクタ
func NewMachineUseCase(machineRepo *database.MachineRepository) *MachineUseCase {
	return &MachineUseCase{machineRepo: machineRepo}
}

// =========================================================================
// Machine Management CRUD (機種管理ロジック) - ルーティングのガード内側で利用
// =========================================================================

// -------------------------------------------------------------------------
// C: Create (作成)
// -------------------------------------------------------------------------

// CreateMachine 新しい機種を作成する
func (u *MachineUseCase) CreateMachine(ctx context.Context, m *model.Machine) error {
	if err := validateDuplicateCode(ctx, m.Code, 0, func(ctx context.Context, code string) (*model.Machine, error) {
		return u.machineRepo.FindByCode(ctx, code)
	}); err != nil {
		return err
	}

	return u.machineRepo.Create(ctx, m)
}

// -------------------------------------------------------------------------
// R: Read (取得)
// -------------------------------------------------------------------------

// GetMachineByID 機種IDを指定して、該当する機種情報を1件取得する
func (u *MachineUseCase) GetMachineByID(ctx context.Context, id int64) (*model.Machine, error) {
	return u.machineRepo.FindByID(ctx, id)
}

// GetMachineByCode 機種コードを指定して、該当する機種情報を1件取得する
func (u *MachineUseCase) GetMachineByCode(ctx context.Context, code string) (*model.Machine, error) {
	return u.machineRepo.FindByCode(ctx, code)
}

// GetAllMachines 登録されているすべての機種情報を取得する（ページングなしの全件マスターデータ用）
func (u *MachineUseCase) GetAllMachines(ctx context.Context) ([]model.Machine, error) {
	return u.machineRepo.FindAll(ctx)
}

// GetMachinesWithPagination 指定されたページ、件数、検索キーワードに基づいて、ページング・検索適用済みの機種情報を取得する
func (u *MachineUseCase) GetMachinesWithPagination(
	ctx context.Context,
	page, limit int,
	filter MachineListFilter,
) ([]model.Machine, int64, int, error) {
	var whereQuery func(*gorm.DB) *gorm.DB

	if filter.SearchWord != "" || len(filter.ManufacturerIDs) > 0 {
		whereQuery = func(db *gorm.DB) *gorm.DB {
			if filter.SearchWord != "" {
				likeQuery := "%" + filter.SearchWord + "%"
				db = db.Where("name LIKE ? OR kana LIKE ?", likeQuery, likeQuery)
			}

			if len(filter.ManufacturerIDs) > 0 {
				db = db.Where("manufacturer_id IN ?", filter.ManufacturerIDs)
			}

			return db
		}
	}

	return usecase.ExecutePaginatedSearch(
		ctx, page, limit, whereQuery,
		u.machineRepo.CountAll,
		u.machineRepo.FindAllWithPagination,
	)
}

// -------------------------------------------------------------------------
// U: Update (更新)
// -------------------------------------------------------------------------

// UpdateMachine 既存の機種情報を更新する
func (u *MachineUseCase) UpdateMachine(ctx context.Context, m *model.Machine) error {
	if err := validateDuplicateCode(ctx, m.Code, m.ID, func(ctx context.Context, code string) (*model.Machine, error) {
		return u.machineRepo.FindByCode(ctx, code)
	}); err != nil {
		return err
	}

	return u.machineRepo.Update(ctx, m)
}

// SetMachineImageKey は機種画像キーを更新する
func (u *MachineUseCase) SetMachineImageKey(ctx context.Context, id int64, imageKey *string) error {
	return u.machineRepo.UpdateImageKey(ctx, id, imageKey)
}

// -------------------------------------------------------------------------
// D: Delete (削除)
// -------------------------------------------------------------------------

// DeleteMachine 機種IDを指定して、該当する機種情報を削除する
func (u *MachineUseCase) DeleteMachine(ctx context.Context, id int64) error {
	return u.machineRepo.Delete(ctx, id)
}
