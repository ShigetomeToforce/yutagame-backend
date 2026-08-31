package admin

import (
	"context"
	"fmt"
	"strings"
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

// CreateManufacturer 新しいメーカーを作成する
func (u *ManufacturerUseCase) CreateManufacturer(ctx context.Context, g *model.Manufacturer) error {
	if err := validateDuplicateCode(ctx, g.Code, 0, func(ctx context.Context, code string) (*model.Manufacturer, error) {
		return u.manufacturerRepo.FindByCode(ctx, code)
	}); err != nil {
		return err
	}

	return u.manufacturerRepo.Create(ctx, g)
}

// ----------------------------------------------------------------------------
// R: Read (取得)
// ----------------------------------------------------------------------------

// GetManufacturerByID メーカーIDを指定して、該当するメーカー情報を1件取得する
func (u *ManufacturerUseCase) GetManufacturerByID(ctx context.Context, id int64) (*model.Manufacturer, error) {
	return u.manufacturerRepo.FindByID(ctx, id)
}

// GetManufacturerByCode メーカーコードを指定して、該当するメーカー情報を1件取得する
func (u *ManufacturerUseCase) GetManufacturerByCode(ctx context.Context, code string) (*model.Manufacturer, error) {
	return u.manufacturerRepo.FindByCode(ctx, code)
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
	if err := validateDuplicateCode(ctx, g.Code, g.ID, func(ctx context.Context, code string) (*model.Manufacturer, error) {
		return u.manufacturerRepo.FindByCode(ctx, code)
	}); err != nil {
		return err
	}

	return u.manufacturerRepo.Update(ctx, g)
}

// SetManufacturerImageKey はメーカー画像キーを更新する
func (u *ManufacturerUseCase) SetManufacturerImageKey(ctx context.Context, id int64, imageKey *string) error {
	return u.manufacturerRepo.UpdateImageKey(ctx, id, imageKey)
}

type ManufacturerCSVRowInput struct {
	RowNumber int64   `json:"rowNumber"`
	ID        *int64  `json:"id"`
	Name      *string `json:"name"`
	Kana      *string `json:"kana"`
	Overview  *string `json:"overview"`
	Code      *string `json:"code"`
}

type ManufacturerCSVFieldDiff struct {
	Field string `json:"field"`
	From  string `json:"from"`
	To    string `json:"to"`
}

type ManufacturerCSVOperation struct {
	RowNumber  int64                      `json:"rowNumber"`
	Action     string                     `json:"action"`
	ID         *int64                     `json:"id"`
	Selectable bool                       `json:"selectable"`
	Reason     string                     `json:"reason"`
	Diffs      []ManufacturerCSVFieldDiff `json:"diffs"`
	Payload    ManufacturerCSVRowInput    `json:"payload"`
}

type ManufacturerCSVPreview struct {
	Operations    []ManufacturerCSVOperation `json:"operations"`
	Creatable     int                        `json:"creatable"`
	Updatable     int                        `json:"updatable"`
	Skipped       int                        `json:"skipped"`
	SelectableAll int                        `json:"selectableAll"`
}

// BuildCSVPreview CSV入力行から、実行可能な更新・登録内容の差分プレビューを作成する
func (u *ManufacturerUseCase) BuildCSVPreview(ctx context.Context, rows []ManufacturerCSVRowInput) (*ManufacturerCSVPreview, error) {
	ids := make([]int64, 0, len(rows))
	for _, row := range rows {
		if row.ID != nil && *row.ID > 0 {
			ids = append(ids, *row.ID)
		}
	}

	existingList, err := u.manufacturerRepo.FindByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}

	existingByID := map[int64]model.Manufacturer{}
	for _, item := range existingList {
		existingByID[item.ID] = item
	}

	preview := &ManufacturerCSVPreview{Operations: make([]ManufacturerCSVOperation, 0, len(rows))}

	for _, row := range rows {
		op := ManufacturerCSVOperation{
			RowNumber:  row.RowNumber,
			ID:         row.ID,
			Action:     "skip",
			Selectable: false,
			Diffs:      make([]ManufacturerCSVFieldDiff, 0),
			Payload:    row,
		}

		if row.ID == nil || *row.ID < 1 {
			op.Reason = "ID列が空、または不正なためスキップ"
			preview.Skipped++
			preview.Operations = append(preview.Operations, op)
			continue
		}

		existing, found := existingByID[*row.ID]
		if !found {
			missing := missingRequiredForManufacturerCreate(row)
			if len(missing) > 0 {
				op.Reason = "新規登録に必要な列が不足: " + strings.Join(missing, ", ")
				preview.Skipped++
				preview.Operations = append(preview.Operations, op)
				continue
			}

			op.Action = "create"
			op.Selectable = true
			op.Reason = "IDに一致する既存レコードがないため新規登録"
			preview.Creatable++
			preview.SelectableAll++
			preview.Operations = append(preview.Operations, op)
			continue
		}

		updates, diffs := buildManufacturerUpdateMapAndDiff(existing, row)
		if len(updates) == 0 {
			op.Reason = "更新対象の差分がないためスキップ"
			preview.Skipped++
			preview.Operations = append(preview.Operations, op)
			continue
		}

		op.Action = "update"
		op.Selectable = true
		op.Diffs = diffs
		op.Reason = "差分あり"
		preview.Updatable++
		preview.SelectableAll++
		preview.Operations = append(preview.Operations, op)
	}

	return preview, nil
}

// ApplyCSVOperations チェックされた差分のみを適用する
func (u *ManufacturerUseCase) ApplyCSVOperations(ctx context.Context, operations []ManufacturerCSVOperation) (int, int, error) {
	created := 0
	updated := 0

	for _, op := range operations {
		if !op.Selectable {
			continue
		}
		if op.ID == nil || *op.ID < 1 {
			continue
		}

		switch op.Action {
		case "create":
			missing := missingRequiredForManufacturerCreate(op.Payload)
			if len(missing) > 0 {
				continue
			}

			code := strings.TrimSpace(valueOrEmpty(op.Payload.Code))
			if err := validateDuplicateCode(ctx, code, 0, func(ctx context.Context, code string) (*model.Manufacturer, error) {
				return u.manufacturerRepo.FindByCode(ctx, code)
			}); err != nil {
				return created, updated, fmt.Errorf("row %d: %w", op.RowNumber, err)
			}

			item := &model.Manufacturer{
				ID:       *op.ID,
				Name:     strings.TrimSpace(valueOrEmpty(op.Payload.Name)),
				Kana:     strings.TrimSpace(valueOrEmpty(op.Payload.Kana)),
				Overview: strings.TrimSpace(valueOrEmpty(op.Payload.Overview)),
				Code:     code,
			}

			if err := u.manufacturerRepo.Create(ctx, item); err != nil {
				return created, updated, fmt.Errorf("row %d create failed: %w", op.RowNumber, err)
			}
			created++

		case "update":
			existing, err := u.manufacturerRepo.FindByID(ctx, *op.ID)
			if err != nil {
				return created, updated, fmt.Errorf("row %d: %w", op.RowNumber, err)
			}
			if existing == nil {
				continue
			}

			updates, _ := buildManufacturerUpdateMapAndDiff(*existing, op.Payload)
			if len(updates) == 0 {
				continue
			}

			if codeRaw, ok := updates["code"]; ok {
				code, _ := codeRaw.(string)
				if err := validateDuplicateCode(ctx, code, *op.ID, func(ctx context.Context, code string) (*model.Manufacturer, error) {
					return u.manufacturerRepo.FindByCode(ctx, code)
				}); err != nil {
					return created, updated, fmt.Errorf("row %d: %w", op.RowNumber, err)
				}
			}

			if err := u.manufacturerRepo.UpdateFieldsByID(ctx, *op.ID, updates); err != nil {
				return created, updated, fmt.Errorf("row %d update failed: %w", op.RowNumber, err)
			}
			updated++
		}
	}

	return created, updated, nil
}

func missingRequiredForManufacturerCreate(row ManufacturerCSVRowInput) []string {
	missing := make([]string, 0, 4)
	if strings.TrimSpace(valueOrEmpty(row.Name)) == "" {
		missing = append(missing, "name")
	}
	if strings.TrimSpace(valueOrEmpty(row.Kana)) == "" {
		missing = append(missing, "kana")
	}
	if strings.TrimSpace(valueOrEmpty(row.Overview)) == "" {
		missing = append(missing, "overview")
	}
	if strings.TrimSpace(valueOrEmpty(row.Code)) == "" {
		missing = append(missing, "code")
	}
	return missing
}

func buildManufacturerUpdateMapAndDiff(existing model.Manufacturer, row ManufacturerCSVRowInput) (map[string]any, []ManufacturerCSVFieldDiff) {
	updates := map[string]any{}
	diffs := make([]ManufacturerCSVFieldDiff, 0, 4)

	if row.Name != nil {
		next := strings.TrimSpace(*row.Name)
		if next != "" && normalizeComparableValue(next) != normalizeComparableValue(existing.Name) {
			updates["name"] = next
			diffs = append(diffs, ManufacturerCSVFieldDiff{Field: "name", From: existing.Name, To: next})
		}
	}

	if row.Kana != nil {
		next := strings.TrimSpace(*row.Kana)
		if next != "" && normalizeComparableValue(next) != normalizeComparableValue(existing.Kana) {
			updates["kana"] = next
			diffs = append(diffs, ManufacturerCSVFieldDiff{Field: "kana", From: existing.Kana, To: next})
		}
	}

	if row.Overview != nil {
		next := strings.TrimSpace(*row.Overview)
		if next != "" && normalizeComparableValue(next) != normalizeComparableValue(existing.Overview) {
			updates["overview"] = next
			diffs = append(diffs, ManufacturerCSVFieldDiff{Field: "overview", From: existing.Overview, To: next})
		}
	}

	if row.Code != nil {
		next := strings.TrimSpace(*row.Code)
		if next != "" && normalizeComparableValue(next) != normalizeComparableValue(existing.Code) {
			updates["code"] = next
			diffs = append(diffs, ManufacturerCSVFieldDiff{Field: "code", From: existing.Code, To: next})
		}
	}

	return updates, diffs
}

// ----------------------------------------------------------------------------
// D: Delete (削除)
// ----------------------------------------------------------------------------

// DeleteManufacturer メーカーIDを指定して、該当するメーカー情報を削除する
func (u *ManufacturerUseCase) DeleteManufacturer(ctx context.Context, id int64) error {
	return u.manufacturerRepo.Delete(ctx, id)
}
