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

// GenreListFilter ジャンル検索用構造体
type GenreListFilter struct {
	SearchWord string
}

// GenreUseCase ジャンル管理のビジネスロジックを担当するユースケース
type GenreUseCase struct {
	genreRepo *database.GenreRepository
}

type GenreCSVRowInput struct {
	RowNumber int64   `json:"rowNumber"`
	ID        *int64  `json:"id"`
	Name      *string `json:"name"`
	Kana      *string `json:"kana"`
	Overview  *string `json:"overview"`
	Code      *string `json:"code"`
}

type GenreCSVFieldDiff struct {
	Field string `json:"field"`
	From  string `json:"from"`
	To    string `json:"to"`
}

type GenreCSVOperation struct {
	RowNumber  int64               `json:"rowNumber"`
	Action     string              `json:"action"` // create, update, skip
	ID         *int64              `json:"id"`
	Selectable bool                `json:"selectable"`
	Reason     string              `json:"reason"`
	Diffs      []GenreCSVFieldDiff `json:"diffs"`
	Payload    GenreCSVRowInput    `json:"payload"`
}

type GenreCSVPreview struct {
	Operations    []GenreCSVOperation `json:"operations"`
	Creatable     int                 `json:"creatable"`
	Updatable     int                 `json:"updatable"`
	Skipped       int                 `json:"skipped"`
	SelectableAll int                 `json:"selectableAll"`
}

// NewGenreUseCase GenreUseCaseの新しいインスタンスを生成するコンストラクタ
func NewGenreUseCase(genreRepo *database.GenreRepository) *GenreUseCase {
	return &GenreUseCase{genreRepo: genreRepo}
}

// =========================================================================
// Genre Management CRUD (ジャンル管理ロジック) - ルーティングのガード内側で利用
// =========================================================================

// -------------------------------------------------------------------------
// C: Create (作成)
// -------------------------------------------------------------------------

// CreateGenre 新しいジャンルを作成する
func (u *GenreUseCase) CreateGenre(ctx context.Context, g *model.Genre) error {
	if err := validateDuplicateCode(ctx, g.Code, 0, func(ctx context.Context, code string) (*model.Genre, error) {
		return u.genreRepo.FindByCode(ctx, code)
	}); err != nil {
		return err
	}

	return u.genreRepo.Create(ctx, g)
}

// -------------------------------------------------------------------------
// R: Read (取得)
// -------------------------------------------------------------------------

// GetGenreByID ジャンルIDを指定して、該当するジャンル情報を1件取得する
func (u *GenreUseCase) GetGenreByID(ctx context.Context, id int64) (*model.Genre, error) {
	return u.genreRepo.FindByID(ctx, id)
}

// GetGenreByCode ジャンルコードを指定して、該当するジャンル情報を1件取得する
func (u *GenreUseCase) GetGenreByCode(ctx context.Context, code string) (*model.Genre, error) {
	return u.genreRepo.FindByCode(ctx, code)
}

// GetAllGenres 登録されているすべてのジャンル情報を取得する（ページングなしの全件マスターデータ用）
func (u *GenreUseCase) GetAllGenres(ctx context.Context) ([]model.Genre, error) {
	return u.genreRepo.FindAll(ctx)
}

// GetGenresWithPagination 指定されたページ、件数、検索キーワードに基づいて、ページング・検索適用済みのジャンル情報を取得する
func (u *GenreUseCase) GetGenresWithPagination(
	ctx context.Context,
	page, limit int,
	filter GenreListFilter,
) ([]model.Genre, int64, int, error) {
	var whereQuery func(*gorm.DB) *gorm.DB
	if filter.SearchWord != "" {
		whereQuery = func(db *gorm.DB) *gorm.DB {
			likeQuery := "%" + filter.SearchWord + "%"
			return db.Where("name LIKE ? OR kana LIKE ?", likeQuery, likeQuery)
		}
	}

	return usecase.ExecutePaginatedSearch(
		ctx, page, limit, whereQuery,
		u.genreRepo.CountAll,
		u.genreRepo.FindAllWithPagination,
	)
}

// -------------------------------------------------------------------------
// U: Update (更新)
// -------------------------------------------------------------------------

// UpdateGenre 既存のジャンル情報を更新する
func (u *GenreUseCase) UpdateGenre(ctx context.Context, g *model.Genre) error {
	if err := validateDuplicateCode(ctx, g.Code, g.ID, func(ctx context.Context, code string) (*model.Genre, error) {
		return u.genreRepo.FindByCode(ctx, code)
	}); err != nil {
		return err
	}

	return u.genreRepo.Update(ctx, g)
}

// SetGenreImageKey はジャンル画像キーを更新する
func (u *GenreUseCase) SetGenreImageKey(ctx context.Context, id int64, imageKey *string) error {
	return u.genreRepo.UpdateImageKey(ctx, id, imageKey)
}

// BuildCSVPreview CSV入力行から、実行可能な更新・登録内容の差分プレビューを作成する
func (u *GenreUseCase) BuildCSVPreview(ctx context.Context, rows []GenreCSVRowInput) (*GenreCSVPreview, error) {
	ids := make([]int64, 0, len(rows))
	for _, row := range rows {
		if row.ID != nil && *row.ID > 0 {
			ids = append(ids, *row.ID)
		}
	}

	existingList, err := u.genreRepo.FindByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}

	existingByID := map[int64]model.Genre{}
	for _, g := range existingList {
		existingByID[g.ID] = g
	}

	preview := &GenreCSVPreview{Operations: make([]GenreCSVOperation, 0, len(rows))}

	for _, row := range rows {
		op := GenreCSVOperation{
			RowNumber:  row.RowNumber,
			ID:         row.ID,
			Action:     "skip",
			Selectable: false,
			Diffs:      make([]GenreCSVFieldDiff, 0),
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
			missing := missingRequiredForCreate(row)
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

		updates, diffs := buildUpdateMapAndDiff(existing, row)
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
func (u *GenreUseCase) ApplyCSVOperations(ctx context.Context, operations []GenreCSVOperation) (int, int, error) {
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
			missing := missingRequiredForCreate(op.Payload)
			if len(missing) > 0 {
				continue
			}

			code := strings.TrimSpace(valueOrEmpty(op.Payload.Code))
			if err := validateDuplicateCode(ctx, code, 0, func(ctx context.Context, code string) (*model.Genre, error) {
				return u.genreRepo.FindByCode(ctx, code)
			}); err != nil {
				return created, updated, fmt.Errorf("row %d: %w", op.RowNumber, err)
			}

			genre := &model.Genre{
				ID:       *op.ID,
				Name:     strings.TrimSpace(valueOrEmpty(op.Payload.Name)),
				Kana:     strings.TrimSpace(valueOrEmpty(op.Payload.Kana)),
				Overview: strings.TrimSpace(valueOrEmpty(op.Payload.Overview)),
				Code:     code,
			}

			if err := u.genreRepo.Create(ctx, genre); err != nil {
				return created, updated, fmt.Errorf("row %d create failed: %w", op.RowNumber, err)
			}
			created++

		case "update":
			existing, err := u.genreRepo.FindByID(ctx, *op.ID)
			if err != nil {
				return created, updated, fmt.Errorf("row %d: %w", op.RowNumber, err)
			}
			if existing == nil {
				continue
			}

			updates, _ := buildUpdateMapAndDiff(*existing, op.Payload)
			if len(updates) == 0 {
				continue
			}

			if codeRaw, ok := updates["code"]; ok {
				code, _ := codeRaw.(string)
				if err := validateDuplicateCode(ctx, code, *op.ID, func(ctx context.Context, code string) (*model.Genre, error) {
					return u.genreRepo.FindByCode(ctx, code)
				}); err != nil {
					return created, updated, fmt.Errorf("row %d: %w", op.RowNumber, err)
				}
			}

			if err := u.genreRepo.UpdateFieldsByID(ctx, *op.ID, updates); err != nil {
				return created, updated, fmt.Errorf("row %d update failed: %w", op.RowNumber, err)
			}
			updated++
		}
	}

	return created, updated, nil
}

func valueOrEmpty(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}

func missingRequiredForCreate(row GenreCSVRowInput) []string {
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

func buildUpdateMapAndDiff(existing model.Genre, row GenreCSVRowInput) (map[string]any, []GenreCSVFieldDiff) {
	updates := map[string]any{}
	diffs := make([]GenreCSVFieldDiff, 0, 4)

	if row.Name != nil {
		next := strings.TrimSpace(*row.Name)
		if next != "" && normalizeComparableValue(next) != normalizeComparableValue(existing.Name) {
			updates["name"] = next
			diffs = append(diffs, GenreCSVFieldDiff{Field: "name", From: existing.Name, To: next})
		}
	}

	if row.Kana != nil {
		next := strings.TrimSpace(*row.Kana)
		if next != "" && normalizeComparableValue(next) != normalizeComparableValue(existing.Kana) {
			updates["kana"] = next
			diffs = append(diffs, GenreCSVFieldDiff{Field: "kana", From: existing.Kana, To: next})
		}
	}

	if row.Overview != nil {
		next := strings.TrimSpace(*row.Overview)
		if next != "" && normalizeComparableValue(next) != normalizeComparableValue(existing.Overview) {
			updates["overview"] = next
			diffs = append(diffs, GenreCSVFieldDiff{Field: "overview", From: existing.Overview, To: next})
		}
	}

	if row.Code != nil {
		next := strings.TrimSpace(*row.Code)
		if next != "" && normalizeComparableValue(next) != normalizeComparableValue(existing.Code) {
			updates["code"] = next
			diffs = append(diffs, GenreCSVFieldDiff{Field: "code", From: existing.Code, To: next})
		}
	}

	return updates, diffs
}

func normalizeComparableValue(v string) string {
	trimmed := strings.TrimSpace(v)
	trimmed = strings.ReplaceAll(trimmed, "\r\n", "\n")
	trimmed = strings.ReplaceAll(trimmed, "\r", "\n")
	return trimmed
}

// -------------------------------------------------------------------------
// D: Delete (削除)
// -------------------------------------------------------------------------

// DeleteGenre ジャンルIDを指定して、該当するジャンル情報を削除する
func (u *GenreUseCase) DeleteGenre(ctx context.Context, id int64) error {
	return u.genreRepo.Delete(ctx, id)
}
