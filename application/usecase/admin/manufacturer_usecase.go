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

// ManufactureListFilter メーカー検索用構造体
type ManufactureListFilter struct {
	SearchWord string
}

// ManufacturerUseCase メーカー管理のビジネスロジックを担当するユースケース
type ManufacturerUseCase struct {
	manufacturerRepo *database.ManufacturerRepository
}

// NewManufacturerUseCase ManufacturerUseCaseの新しいインスタンスを生成するコンストラクタ
func NewManufacturerUseCase(manufacturerRepo *database.ManufacturerRepository) *ManufacturerUseCase {
	return &ManufacturerUseCase{manufacturerRepo: manufacturerRepo}
}

// ============================================================================
// Manufacturer Management CRUD (メーカー管理ロジック) - ルーティングのガード内側で利用
// ============================================================================

// ----------------------------------------------------------------------------
// C: Create (作成)
// ----------------------------------------------------------------------------

// CreateGenre 新しいメーカーを作成する
func (u *ManufacturerUseCase) CreateManufacturer(ctx context.Context, g *model.Manufacturer) error {
	return u.manufacturerRepo.Create(ctx, g)
}

// ----------------------------------------------------------------------------
// R: Read (取得)
// ----------------------------------------------------------------------------

// GetManufacturerByID メーカーIDを指定して、該当するメーカー情報を1件取得する
func (u *ManufacturerUseCase) GetManufacturerByID(ctx context.Context, id int64) (*model.Manufacturer, error) {
	return u.manufacturerRepo.FindByID(ctx, id)
}

// GetAllManufacturers 登録されているすべてのメーカー情報を取得する（ページングなしの全件マスターデータ用）
func (u *ManufacturerUseCase) GetAllManufacturers(ctx context.Context) ([]model.Manufacturer, error) {
	return u.manufacturerRepo.FindAll(ctx)
}

// GetManufacturersWithPagination 指定されたページ、件数、検索キーワードに基づいて、ページング・検索適用済みのメーカー情報を取得する
func (u *ManufacturerUseCase) GetManufacturersWithPagination(
	ctx context.Context,
	page, limit int,
	filter ManufactureListFilter,
) ([]model.Manufacturer, int64, int, error) {
	var whereQuery func(*gorm.DB) *gorm.DB
	if filter.SearchWord != "" {
		whereQuery = func(db *gorm.DB) *gorm.DB {
			likeQuery := "%" + filter.SearchWord + "%"
			return db.Where("name LIKE ? OR kana LIKE ?", likeQuery, likeQuery)
		}
	}

	return usecase.ExecutePaginatedSearch(
		ctx, page, limit, whereQuery,
		u.manufacturerRepo.CountAll,
		u.manufacturerRepo.FindAllWithPagination,
	)
}

// ----------------------------------------------------------------------------
// U: Update (更新)
// ----------------------------------------------------------------------------

// UpdateManufacturer 既存のメーカー情報を更新する
func (u *ManufacturerUseCase) UpdateManufacturer(ctx context.Context, g *model.Manufacturer) error {
	return u.manufacturerRepo.Update(ctx, g)
}

// ----------------------------------------------------------------------------
// D: Delete (削除)
// ----------------------------------------------------------------------------

// DeleteManufacturer メーカーIDを指定して、該当するメーカー情報を削除する
func (u *ManufacturerUseCase) DeleteManufacturer(ctx context.Context, id int64) error {
	return u.manufacturerRepo.Delete(ctx, id)
}
