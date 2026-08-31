package admin

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"yutagame-backend/application/usecase"
	"yutagame-backend/domain/model"
	"yutagame-backend/infrastructure/database"

	"gorm.io/gorm"
)

// =========================================================================
// 構造体＆コンストラクタ
// =========================================================================

// KeywordListFilter キーワード検索用構造体
type KeywordListFilter struct {
	SearchWord string
}

// KeywordUseCase キーワード管理のビジネスロジックを担当するユースケース
type KeywordUseCase struct {
	keywordRepo *database.KeywordRepository
}

// NewKeywordUseCase KeywordUseCaseの新しいインスタンスを生成するコンストラクタ
func NewKeywordUseCase(keywordRepo *database.KeywordRepository) *KeywordUseCase {
	return &KeywordUseCase{keywordRepo: keywordRepo}
}

// =========================================================================
// Keyword Management CRUD (キーワード管理ロジック) - ルーティングのガード内側で利用
// =========================================================================

// -------------------------------------------------------------------------
// C: Create (作成)
// -------------------------------------------------------------------------

// CreateKeyword 新しいキーワードを作成する
func (u *KeywordUseCase) CreateKeyword(ctx context.Context, g *model.Keyword) error {
	if err := validateDuplicateCode(ctx, g.Code, 0, func(ctx context.Context, code string) (*model.Keyword, error) {
		return u.keywordRepo.FindByCode(ctx, code)
	}); err != nil {
		return err
	}

	return u.keywordRepo.Create(ctx, g)
}

// -------------------------------------------------------------------------
// R: Read (取得)
// -------------------------------------------------------------------------

// GetKeywordByID キーワードIDを指定して、該当するキーワード情報を1件取得する
func (u *KeywordUseCase) GetKeywordByID(ctx context.Context, id int64) (*model.Keyword, error) {
	return u.keywordRepo.FindByID(ctx, id)
}

// GetKeywordByCode キーワードコードを指定して、該当するキーワード情報を1件取得する
func (u *KeywordUseCase) GetKeywordByCode(ctx context.Context, code string) (*model.Keyword, error) {
	return u.keywordRepo.FindByCode(ctx, code)
}

// GetAllKeywords 登録されているすべてのキーワード情報を取得する（ページングなしの全件マスターデータ用）
func (u *KeywordUseCase) GetAllKeywords(ctx context.Context) ([]model.Keyword, error) {
	return u.keywordRepo.FindAll(ctx)
}

// GetKeywordsWithPagination 指定されたページ、件数、検索キーワードに基づいて、ページング・検索適用済みのキーワード情報を取得する
func (u *KeywordUseCase) GetKeywordsWithPagination(
	ctx context.Context,
	page, limit int,
	filter KeywordListFilter,
) ([]model.Keyword, int64, int, error) {
	var whereQuery func(*gorm.DB) *gorm.DB
	if filter.SearchWord != "" {
		whereQuery = func(db *gorm.DB) *gorm.DB {
			likeQuery := "%" + filter.SearchWord + "%"
			return db.Where("name LIKE ? OR kana LIKE ?", likeQuery, likeQuery)
		}
	}

	return usecase.ExecutePaginatedSearch(
		ctx, page, limit, whereQuery,
		u.keywordRepo.CountAll,
		u.keywordRepo.FindAllWithPagination,
	)
}

// -------------------------------------------------------------------------
// U: Update (更新)
// -------------------------------------------------------------------------

// UpdateKeyword 既存のキーワード情報を更新する
func (u *KeywordUseCase) UpdateKeyword(ctx context.Context, g *model.Keyword) error {
	if err := validateDuplicateCode(ctx, g.Code, g.ID, func(ctx context.Context, code string) (*model.Keyword, error) {
		return u.keywordRepo.FindByCode(ctx, code)
	}); err != nil {
		return err
	}

	return u.keywordRepo.Update(ctx, g)
}

type KeywordCSVRowInput struct {
	RowNumber   int64   `json:"rowNumber"`
	ID          *int64  `json:"id"`
	Name        *string `json:"name"`
	Kana        *string `json:"kana"`
	Overview    *string `json:"overview"`
	Code        *string `json:"code"`
	KeywordType *string `json:"keywordType"`
	SortOrder   *int32  `json:"sortOrder"`
}

type KeywordCSVFieldDiff struct {
	Field string `json:"field"`
	From  string `json:"from"`
	To    string `json:"to"`
}

type KeywordCSVOperation struct {
	RowNumber  int64                 `json:"rowNumber"`
	Action     string                `json:"action"`
	ID         *int64                `json:"id"`
	Selectable bool                  `json:"selectable"`
	Reason     string                `json:"reason"`
	Diffs      []KeywordCSVFieldDiff `json:"diffs"`
	Payload    KeywordCSVRowInput    `json:"payload"`
}

type KeywordCSVPreview struct {
	Operations    []KeywordCSVOperation `json:"operations"`
	Creatable     int                   `json:"creatable"`
	Updatable     int                   `json:"updatable"`
	Skipped       int                   `json:"skipped"`
	SelectableAll int                   `json:"selectableAll"`
}

// BuildCSVPreview CSV入力行から、実行可能な更新・登録内容の差分プレビューを作成する
func (u *KeywordUseCase) BuildCSVPreview(ctx context.Context, rows []KeywordCSVRowInput) (*KeywordCSVPreview, error) {
	ids := make([]int64, 0, len(rows))
	for _, row := range rows {
		if row.ID != nil && *row.ID > 0 {
			ids = append(ids, *row.ID)
		}
	}

	existingList, err := u.keywordRepo.FindByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}

	existingByID := map[int64]model.Keyword{}
	for _, item := range existingList {
		existingByID[item.ID] = item
	}

	preview := &KeywordCSVPreview{Operations: make([]KeywordCSVOperation, 0, len(rows))}

	for _, row := range rows {
		op := KeywordCSVOperation{
			RowNumber:  row.RowNumber,
			ID:         row.ID,
			Action:     "skip",
			Selectable: false,
			Diffs:      make([]KeywordCSVFieldDiff, 0),
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
			missing := missingRequiredForKeywordCreate(row)
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

		updates, diffs := buildKeywordUpdateMapAndDiff(existing, row)
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
func (u *KeywordUseCase) ApplyCSVOperations(ctx context.Context, operations []KeywordCSVOperation) (int, int, error) {
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
			missing := missingRequiredForKeywordCreate(op.Payload)
			if len(missing) > 0 {
				continue
			}

			code := strings.TrimSpace(valueOrEmpty(op.Payload.Code))
			if err := validateDuplicateCode(ctx, code, 0, func(ctx context.Context, code string) (*model.Keyword, error) {
				return u.keywordRepo.FindByCode(ctx, code)
			}); err != nil {
				return created, updated, fmt.Errorf("row %d: %w", op.RowNumber, err)
			}

			sortOrder := int32(0)
			if op.Payload.SortOrder != nil {
				sortOrder = *op.Payload.SortOrder
			}

			item := &model.Keyword{
				ID:          *op.ID,
				Name:        strings.TrimSpace(valueOrEmpty(op.Payload.Name)),
				Kana:        strings.TrimSpace(valueOrEmpty(op.Payload.Kana)),
				Overview:    strings.TrimSpace(valueOrEmpty(op.Payload.Overview)),
				Code:        code,
				KeywordType: strings.TrimSpace(valueOrEmpty(op.Payload.KeywordType)),
				SortOrder:   sortOrder,
			}

			if err := u.keywordRepo.Create(ctx, item); err != nil {
				return created, updated, fmt.Errorf("row %d create failed: %w", op.RowNumber, err)
			}
			created++

		case "update":
			existing, err := u.keywordRepo.FindByID(ctx, *op.ID)
			if err != nil {
				return created, updated, fmt.Errorf("row %d: %w", op.RowNumber, err)
			}
			if existing == nil {
				continue
			}

			updates, _ := buildKeywordUpdateMapAndDiff(*existing, op.Payload)
			if len(updates) == 0 {
				continue
			}

			if codeRaw, ok := updates["code"]; ok {
				code, _ := codeRaw.(string)
				if err := validateDuplicateCode(ctx, code, *op.ID, func(ctx context.Context, code string) (*model.Keyword, error) {
					return u.keywordRepo.FindByCode(ctx, code)
				}); err != nil {
					return created, updated, fmt.Errorf("row %d: %w", op.RowNumber, err)
				}
			}

			if err := u.keywordRepo.UpdateFieldsByID(ctx, *op.ID, updates); err != nil {
				return created, updated, fmt.Errorf("row %d update failed: %w", op.RowNumber, err)
			}
			updated++
		}
	}

	return created, updated, nil
}

func missingRequiredForKeywordCreate(row KeywordCSVRowInput) []string {
	missing := make([]string, 0, 6)
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
	if strings.TrimSpace(valueOrEmpty(row.KeywordType)) == "" {
		missing = append(missing, "keyword_type")
	}
	if row.SortOrder == nil {
		missing = append(missing, "sort_order")
	}
	return missing
}

func buildKeywordUpdateMapAndDiff(existing model.Keyword, row KeywordCSVRowInput) (map[string]any, []KeywordCSVFieldDiff) {
	updates := map[string]any{}
	diffs := make([]KeywordCSVFieldDiff, 0, 6)

	if row.Name != nil {
		next := strings.TrimSpace(*row.Name)
		if next != "" && normalizeComparableValue(next) != normalizeComparableValue(existing.Name) {
			updates["name"] = next
			diffs = append(diffs, KeywordCSVFieldDiff{Field: "name", From: existing.Name, To: next})
		}
	}

	if row.Kana != nil {
		next := strings.TrimSpace(*row.Kana)
		if next != "" && normalizeComparableValue(next) != normalizeComparableValue(existing.Kana) {
			updates["kana"] = next
			diffs = append(diffs, KeywordCSVFieldDiff{Field: "kana", From: existing.Kana, To: next})
		}
	}

	if row.Overview != nil {
		next := strings.TrimSpace(*row.Overview)
		if next != "" && normalizeComparableValue(next) != normalizeComparableValue(existing.Overview) {
			updates["overview"] = next
			diffs = append(diffs, KeywordCSVFieldDiff{Field: "overview", From: existing.Overview, To: next})
		}
	}

	if row.Code != nil {
		next := strings.TrimSpace(*row.Code)
		if next != "" && normalizeComparableValue(next) != normalizeComparableValue(existing.Code) {
			updates["code"] = next
			diffs = append(diffs, KeywordCSVFieldDiff{Field: "code", From: existing.Code, To: next})
		}
	}

	if row.KeywordType != nil {
		next := strings.TrimSpace(*row.KeywordType)
		if next != "" && normalizeComparableValue(next) != normalizeComparableValue(existing.KeywordType) {
			updates["keyword_type"] = next
			diffs = append(diffs, KeywordCSVFieldDiff{Field: "keyword_type", From: existing.KeywordType, To: next})
		}
	}

	if row.SortOrder != nil {
		next := *row.SortOrder
		if next != existing.SortOrder {
			updates["sort_order"] = next
			diffs = append(diffs, KeywordCSVFieldDiff{
				Field: "sort_order",
				From:  strconv.FormatInt(int64(existing.SortOrder), 10),
				To:    strconv.FormatInt(int64(next), 10),
			})
		}
	}

	return updates, diffs
}

// -------------------------------------------------------------------------
// D: Delete (削除)
// -------------------------------------------------------------------------

// DeleteKeyword キーワードIDを指定して、該当するキーワード情報を削除する
func (u *KeywordUseCase) DeleteKeyword(ctx context.Context, id int64) error {
	return u.keywordRepo.Delete(ctx, id)
}
