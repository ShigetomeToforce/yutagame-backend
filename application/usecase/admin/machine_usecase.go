package admin

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"
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

type MachineCSVRowInput struct {
	RowNumber      int64   `json:"rowNumber"`
	ID             *int64  `json:"id"`
	Name           *string `json:"name"`
	Kana           *string `json:"kana"`
	Overview       *string `json:"overview"`
	Code           *string `json:"code"`
	Abbreviation   *string `json:"abbreviation"`
	ManufacturerID *int64  `json:"manufacturerId"`
	MachineType    *string `json:"machineType"`
	ReleaseDate    *string `json:"releaseDate"`
	SortOrder      *int32  `json:"sortOrder"`
}

type MachineCSVFieldDiff struct {
	Field string `json:"field"`
	From  string `json:"from"`
	To    string `json:"to"`
}

type MachineCSVOperation struct {
	RowNumber  int64                 `json:"rowNumber"`
	Action     string                `json:"action"`
	ID         *int64                `json:"id"`
	Selectable bool                  `json:"selectable"`
	Reason     string                `json:"reason"`
	Diffs      []MachineCSVFieldDiff `json:"diffs"`
	Payload    MachineCSVRowInput    `json:"payload"`
}

type MachineCSVPreview struct {
	Operations    []MachineCSVOperation `json:"operations"`
	Creatable     int                   `json:"creatable"`
	Updatable     int                   `json:"updatable"`
	Skipped       int                   `json:"skipped"`
	SelectableAll int                   `json:"selectableAll"`
}

// BuildCSVPreview CSV入力行から、実行可能な更新・登録内容の差分プレビューを作成する
func (u *MachineUseCase) BuildCSVPreview(ctx context.Context, rows []MachineCSVRowInput) (*MachineCSVPreview, error) {
	ids := make([]int64, 0, len(rows))
	for _, row := range rows {
		if row.ID != nil && *row.ID > 0 {
			ids = append(ids, *row.ID)
		}
	}

	existingList, err := u.machineRepo.FindByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}

	existingByID := map[int64]model.Machine{}
	for _, item := range existingList {
		existingByID[item.ID] = item
	}

	preview := &MachineCSVPreview{Operations: make([]MachineCSVOperation, 0, len(rows))}

	for _, row := range rows {
		op := MachineCSVOperation{
			RowNumber:  row.RowNumber,
			ID:         row.ID,
			Action:     "skip",
			Selectable: false,
			Diffs:      make([]MachineCSVFieldDiff, 0),
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
			missing := missingRequiredForMachineCreate(row)
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

		updates, diffs, buildErr := buildMachineUpdateMapAndDiff(existing, row)
		if buildErr != nil {
			op.Reason = buildErr.Error()
			preview.Skipped++
			preview.Operations = append(preview.Operations, op)
			continue
		}

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
func (u *MachineUseCase) ApplyCSVOperations(ctx context.Context, operations []MachineCSVOperation) (int, int, error) {
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
			missing := missingRequiredForMachineCreate(op.Payload)
			if len(missing) > 0 {
				continue
			}

			code := strings.TrimSpace(valueOrEmpty(op.Payload.Code))
			if err := validateDuplicateCode(ctx, code, 0, func(ctx context.Context, code string) (*model.Machine, error) {
				return u.machineRepo.FindByCode(ctx, code)
			}); err != nil {
				return created, updated, fmt.Errorf("row %d: %w", op.RowNumber, err)
			}

			releaseDate, err := parseCSVDate(valueOrEmpty(op.Payload.ReleaseDate))
			if err != nil {
				return created, updated, fmt.Errorf("row %d: release_date: %w", op.RowNumber, err)
			}

			sortOrder := int32(0)
			if op.Payload.SortOrder != nil {
				sortOrder = *op.Payload.SortOrder
			}

			item := &model.Machine{
				ID:             *op.ID,
				Name:           strings.TrimSpace(valueOrEmpty(op.Payload.Name)),
				Kana:           strings.TrimSpace(valueOrEmpty(op.Payload.Kana)),
				Overview:       strings.TrimSpace(valueOrEmpty(op.Payload.Overview)),
				Code:           code,
				Abbreviation:   strings.TrimSpace(valueOrEmpty(op.Payload.Abbreviation)),
				ManufacturerID: valueInt64OrZero(op.Payload.ManufacturerID),
				MachineType:    strings.TrimSpace(valueOrEmpty(op.Payload.MachineType)),
				ReleaseDate:    releaseDate,
				SortOrder:      sortOrder,
			}

			if err := u.machineRepo.Create(ctx, item); err != nil {
				return created, updated, fmt.Errorf("row %d create failed: %w", op.RowNumber, err)
			}
			created++

		case "update":
			existing, err := u.machineRepo.FindByID(ctx, *op.ID)
			if err != nil {
				return created, updated, fmt.Errorf("row %d: %w", op.RowNumber, err)
			}
			if existing == nil {
				continue
			}

			updates, _, buildErr := buildMachineUpdateMapAndDiff(*existing, op.Payload)
			if buildErr != nil {
				return created, updated, fmt.Errorf("row %d: %w", op.RowNumber, buildErr)
			}
			if len(updates) == 0 {
				continue
			}

			if codeRaw, ok := updates["code"]; ok {
				code, _ := codeRaw.(string)
				if err := validateDuplicateCode(ctx, code, *op.ID, func(ctx context.Context, code string) (*model.Machine, error) {
					return u.machineRepo.FindByCode(ctx, code)
				}); err != nil {
					return created, updated, fmt.Errorf("row %d: %w", op.RowNumber, err)
				}
			}

			if err := u.machineRepo.UpdateFieldsByID(ctx, *op.ID, updates); err != nil {
				return created, updated, fmt.Errorf("row %d update failed: %w", op.RowNumber, err)
			}
			updated++
		}
	}

	return created, updated, nil
}

func missingRequiredForMachineCreate(row MachineCSVRowInput) []string {
	missing := make([]string, 0, 9)
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
	if strings.TrimSpace(valueOrEmpty(row.Abbreviation)) == "" {
		missing = append(missing, "abbreviation")
	}
	if row.ManufacturerID == nil || *row.ManufacturerID < 1 {
		missing = append(missing, "manufacturer_id")
	}
	if strings.TrimSpace(valueOrEmpty(row.MachineType)) == "" {
		missing = append(missing, "machine_type")
	}
	if strings.TrimSpace(valueOrEmpty(row.ReleaseDate)) == "" {
		missing = append(missing, "release_date")
	}
	if row.SortOrder == nil {
		missing = append(missing, "sort_order")
	}
	return missing
}

func buildMachineUpdateMapAndDiff(existing model.Machine, row MachineCSVRowInput) (map[string]any, []MachineCSVFieldDiff, error) {
	updates := map[string]any{}
	diffs := make([]MachineCSVFieldDiff, 0, 9)

	if row.Name != nil {
		next := strings.TrimSpace(*row.Name)
		if next != "" && normalizeComparableValue(next) != normalizeComparableValue(existing.Name) {
			updates["name"] = next
			diffs = append(diffs, MachineCSVFieldDiff{Field: "name", From: existing.Name, To: next})
		}
	}

	if row.Kana != nil {
		next := strings.TrimSpace(*row.Kana)
		if next != "" && normalizeComparableValue(next) != normalizeComparableValue(existing.Kana) {
			updates["kana"] = next
			diffs = append(diffs, MachineCSVFieldDiff{Field: "kana", From: existing.Kana, To: next})
		}
	}

	if row.Overview != nil {
		next := strings.TrimSpace(*row.Overview)
		if next != "" && normalizeComparableValue(next) != normalizeComparableValue(existing.Overview) {
			updates["overview"] = next
			diffs = append(diffs, MachineCSVFieldDiff{Field: "overview", From: existing.Overview, To: next})
		}
	}

	if row.Code != nil {
		next := strings.TrimSpace(*row.Code)
		if next != "" && normalizeComparableValue(next) != normalizeComparableValue(existing.Code) {
			updates["code"] = next
			diffs = append(diffs, MachineCSVFieldDiff{Field: "code", From: existing.Code, To: next})
		}
	}

	if row.Abbreviation != nil {
		next := strings.TrimSpace(*row.Abbreviation)
		if next != "" && normalizeComparableValue(next) != normalizeComparableValue(existing.Abbreviation) {
			updates["abbreviation"] = next
			diffs = append(diffs, MachineCSVFieldDiff{Field: "abbreviation", From: existing.Abbreviation, To: next})
		}
	}

	if row.ManufacturerID != nil {
		next := *row.ManufacturerID
		if next > 0 && next != existing.ManufacturerID {
			updates["manufacturer_id"] = next
			diffs = append(diffs, MachineCSVFieldDiff{Field: "manufacturer_id", From: strconv.FormatInt(existing.ManufacturerID, 10), To: strconv.FormatInt(next, 10)})
		}
	}

	if row.MachineType != nil {
		next := strings.TrimSpace(*row.MachineType)
		if next != "" && normalizeComparableValue(next) != normalizeComparableValue(existing.MachineType) {
			updates["machine_type"] = next
			diffs = append(diffs, MachineCSVFieldDiff{Field: "machine_type", From: existing.MachineType, To: next})
		}
	}

	if row.ReleaseDate != nil {
		nextRaw := strings.TrimSpace(*row.ReleaseDate)
		if nextRaw != "" {
			nextDate, err := parseCSVDate(nextRaw)
			if err != nil {
				return nil, nil, fmt.Errorf("release_date の形式が不正です。YYYY-MM-DD を指定してください")
			}
			next := nextDate.Format("2006-01-02")
			current := existing.ReleaseDate.Format("2006-01-02")
			if next != current {
				updates["release_date"] = nextDate
				diffs = append(diffs, MachineCSVFieldDiff{Field: "release_date", From: current, To: next})
			}
		}
	}

	if row.SortOrder != nil {
		next := *row.SortOrder
		if next != existing.SortOrder {
			updates["sort_order"] = next
			diffs = append(diffs, MachineCSVFieldDiff{Field: "sort_order", From: strconv.FormatInt(int64(existing.SortOrder), 10), To: strconv.FormatInt(int64(next), 10)})
		}
	}

	return updates, diffs, nil
}

func parseCSVDate(v string) (time.Time, error) {
	return time.Parse("2006-01-02", strings.TrimSpace(v))
}

func valueInt64OrZero(v *int64) int64 {
	if v == nil {
		return 0
	}
	return *v
}

// -------------------------------------------------------------------------
// D: Delete (削除)
// -------------------------------------------------------------------------

// DeleteMachine 機種IDを指定して、該当する機種情報を削除する
func (u *MachineUseCase) DeleteMachine(ctx context.Context, id int64) error {
	return u.machineRepo.Delete(ctx, id)
}
