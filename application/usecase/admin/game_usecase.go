package admin

import (
	"context"
	"fmt"
	"net/url"
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

// GameListFilter ゲーム検索用構造体
// 各 boolean フィールドは nil = 未指定, true = 条件付き, false = 条件付き(false) として扱う。
type GameListFilter struct {
	SearchWord      string
	ManufacturerIDs []int64
	MachineIDs      []int64
	GenreIDs        []int64
	KeywordIDs      []int64
	IsPlay          *bool
	IsClear         *bool
	IsFavourite     *bool
}

// GameUseCase ゲーム管理のビジネスロジックを担当するユースケース
type GameUseCase struct {
	gameRepo *database.GameRepository
}

var validAffiliateCategories = map[string]struct{}{
	"AMAZON":            {},
	"RAKUTEN":           {},
	"YAHOO":             {},
	"SURUGAYA":          {},
	"PLAYSTATION_STORE": {},
	"NINTENDO_STORE":    {},
	"STEAM":             {},
}

// NewGameUseCase GameUseCaseの新しいインスタンスを生成するコンストラクタ
func NewGameUseCase(gameRepo *database.GameRepository) *GameUseCase {
	return &GameUseCase{gameRepo: gameRepo}
}

// =========================================================================
// Game Management CRUD (ゲーム管理ロジック) - ルーティングのガード内側で利用
// =========================================================================

// -------------------------------------------------------------------------
// C: Create (作成)
// -------------------------------------------------------------------------

// CreateGame 新しいゲームを作成する
func (u *GameUseCase) CreateGame(ctx context.Context, g *model.Game, keywordIDs []int64) error {
	if err := validateDuplicateCode(ctx, g.Code, 0, func(ctx context.Context, code string) (*model.Game, error) {
		return u.gameRepo.FindByCode(ctx, code)
	}); err != nil {
		return err
	}

	if err := u.gameRepo.Create(ctx, g); err != nil {
		return err
	}

	if len(keywordIDs) > 0 {
		if err := u.gameRepo.ReplaceKeywords(ctx, g.ID, keywordIDs); err != nil {
			return err
		}
	}

	return nil
}

// -------------------------------------------------------------------------
// R: Read (取得)
// -------------------------------------------------------------------------

// GetGameByID ゲームIDを指定して、該当するゲーム情報を1件取得する
func (u *GameUseCase) GetGameByID(ctx context.Context, id int64) (*model.Game, error) {
	return u.gameRepo.FindByID(ctx, id)
}

// GetGameByCode ゲームコードを指定して、該当するゲーム情報を1件取得する
func (u *GameUseCase) GetGameByCode(ctx context.Context, code string) (*model.Game, error) {
	return u.gameRepo.FindByCode(ctx, code)
}

// GetAllGames 登録されているすべてのゲーム情報を取得する（ページングなしの全件マスターデータ用）
func (u *GameUseCase) GetAllGames(ctx context.Context) ([]model.Game, error) {
	return u.gameRepo.FindAll(ctx)
}

// GetGamesWithPagination 指定されたページ、件数、検索キーワードに基づいて、ページング・検索適用済みのゲーム情報を取得する
func (u *GameUseCase) GetGamesWithPagination(
	ctx context.Context,
	page, limit int,
	filter GameListFilter,
) ([]model.Game, int64, int, error) {
	var whereQuery func(*gorm.DB) *gorm.DB

	if filter.SearchWord != "" || len(filter.ManufacturerIDs) > 0 ||
		len(filter.MachineIDs) > 0 || len(filter.GenreIDs) > 0 ||
		len(filter.KeywordIDs) > 0 || filter.IsPlay != nil ||
		filter.IsClear != nil || filter.IsFavourite != nil {

		whereQuery = func(db *gorm.DB) *gorm.DB {
			if filter.SearchWord != "" {
				likeQuery := "%" + filter.SearchWord + "%"
				db = db.Where("name LIKE ? OR kana LIKE ?", likeQuery, likeQuery)
			}

			if len(filter.ManufacturerIDs) > 0 {
				db = db.Where("manufacturer_id IN ?", filter.ManufacturerIDs)
			}

			if len(filter.MachineIDs) > 0 {
				db = db.Where("machine_id IN ?", filter.MachineIDs)
			}

			if len(filter.GenreIDs) > 0 {
				db = db.Where("genre_id IN ?", filter.GenreIDs)
			}

			if len(filter.KeywordIDs) > 0 {
				db = db.Where(`
                    EXISTS (
                        SELECT 1
                        FROM game_keywords gk
                        WHERE gk.game_id = games.id
                        AND gk.keyword_id IN ?
                    )
                `, filter.KeywordIDs)
			}

			if filter.IsPlay != nil {
				db = db.Where("is_play = ?", *filter.IsPlay)
			}

			if filter.IsClear != nil {
				db = db.Where("is_clear = ?", *filter.IsClear)
			}

			if filter.IsFavourite != nil {
				db = db.Where("is_favourite = ?", *filter.IsFavourite)
			}

			return db
		}
	}

	return usecase.ExecutePaginatedSearch(
		ctx, page, limit, whereQuery,
		u.gameRepo.CountAll,
		u.gameRepo.FindAllWithPagination,
	)
}

// -------------------------------------------------------------------------
// U: Update (更新)
// -------------------------------------------------------------------------

// UpdateGame 既存のゲーム情報を更新する
func (u *GameUseCase) UpdateGame(ctx context.Context, g *model.Game, keywordIDs []int64) error {
	if err := validateDuplicateCode(ctx, g.Code, g.ID, func(ctx context.Context, code string) (*model.Game, error) {
		return u.gameRepo.FindByCode(ctx, code)
	}); err != nil {
		return err
	}

	if err := u.gameRepo.Update(ctx, g); err != nil {
		return err
	}

	return u.gameRepo.ReplaceKeywords(ctx, g.ID, keywordIDs)
}

// SetGameImageKey ゲーム画像キーを更新する
func (u *GameUseCase) SetGameImageKey(ctx context.Context, id int64, imageKey *string) error {
	return u.gameRepo.UpdateImageKey(ctx, id, imageKey)
}

func (u *GameUseCase) ListAffiliatesByGameID(ctx context.Context, gameID int64) ([]model.GameAffiliate, error) {
	return u.gameRepo.ListAffiliatesByGameID(ctx, gameID)
}

func (u *GameUseCase) CreateAffiliate(ctx context.Context, gameID int64, category, rawURL string) (*model.GameAffiliate, error) {
	if err := u.validateAffiliateInput(category, rawURL); err != nil {
		return nil, err
	}

	game, err := u.gameRepo.FindByID(ctx, gameID)
	if err != nil {
		return nil, err
	}
	if game == nil {
		return nil, fmt.Errorf("指定されたゲームが見つかりませんでした")
	}

	category = normalizeAffiliateCategory(category)
	urlValue := strings.TrimSpace(rawURL)

	existing, err := u.gameRepo.FindAffiliateByGameIDAndCategory(ctx, gameID, category)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, fmt.Errorf("同じカテゴリの購入リンクは既に登録されています")
	}

	item := &model.GameAffiliate{
		GameID:   gameID,
		Category: category,
		URL:      urlValue,
	}
	if err := u.gameRepo.CreateAffiliate(ctx, item); err != nil {
		return nil, err
	}
	return item, nil
}

func (u *GameUseCase) UpdateAffiliate(ctx context.Context, gameID, affiliateID int64, category, rawURL string) (*model.GameAffiliate, error) {
	if err := u.validateAffiliateInput(category, rawURL); err != nil {
		return nil, err
	}

	item, err := u.gameRepo.FindAffiliateByID(ctx, affiliateID)
	if err != nil {
		return nil, err
	}
	if item == nil || item.GameID != gameID {
		return nil, fmt.Errorf("指定された購入リンクが見つかりませんでした")
	}

	category = normalizeAffiliateCategory(category)
	urlValue := strings.TrimSpace(rawURL)

	duplicate, err := u.gameRepo.FindAffiliateByGameIDAndCategory(ctx, gameID, category)
	if err != nil {
		return nil, err
	}
	if duplicate != nil && duplicate.ID != item.ID {
		return nil, fmt.Errorf("同じカテゴリの購入リンクは既に登録されています")
	}

	item.Category = category
	item.URL = urlValue
	if err := u.gameRepo.UpdateAffiliate(ctx, item); err != nil {
		return nil, err
	}
	return item, nil
}

func (u *GameUseCase) DeleteAffiliate(ctx context.Context, gameID, affiliateID int64) error {
	item, err := u.gameRepo.FindAffiliateByID(ctx, affiliateID)
	if err != nil {
		return err
	}
	if item == nil || item.GameID != gameID {
		return fmt.Errorf("指定された購入リンクが見つかりませんでした")
	}
	return u.gameRepo.DeleteAffiliate(ctx, affiliateID)
}

func normalizeAffiliateCategory(value string) string {
	return strings.ToUpper(strings.TrimSpace(value))
}

func (u *GameUseCase) validateAffiliateInput(category, rawURL string) error {
	normalizedCategory := normalizeAffiliateCategory(category)
	if _, ok := validAffiliateCategories[normalizedCategory]; !ok {
		return fmt.Errorf("カテゴリが不正です")
	}

	urlValue := strings.TrimSpace(rawURL)
	if urlValue == "" {
		return fmt.Errorf("URLは必須です")
	}
	parsed, err := url.ParseRequestURI(urlValue)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return fmt.Errorf("URLの形式が不正です")
	}
	return nil
}

type GameCSVRowInput struct {
	RowNumber       int64   `json:"rowNumber"`
	ID              *int64  `json:"id"`
	Name            *string `json:"name"`
	Kana            *string `json:"kana"`
	Overview        *string `json:"overview"`
	Code            *string `json:"code"`
	ManufacturerID  *int64  `json:"manufacturerId"`
	MachineID       *int64  `json:"machineId"`
	GenreID         *int64  `json:"genreId"`
	SubGenre        *string `json:"subGenre"`
	CatchCopy       *string `json:"catchCopy"`
	SubCatch        *string `json:"subCatch"`
	ListPrice       *int32  `json:"listPrice"`
	ReleaseDate     *string `json:"releaseDate"`
	OfficialSiteURL *string `json:"officialSiteUrl"`
	YouTubeURL      *string `json:"youtubeUrl"`
	IsPlay          *bool   `json:"isPlay"`
	IsClear         *bool   `json:"isClear"`
	IsFavourite     *bool   `json:"isFavourite"`
}

type GameCSVFieldDiff struct {
	Field string `json:"field"`
	From  string `json:"from"`
	To    string `json:"to"`
}

type GameCSVOperation struct {
	RowNumber  int64              `json:"rowNumber"`
	Action     string             `json:"action"`
	ID         *int64             `json:"id"`
	Selectable bool               `json:"selectable"`
	Reason     string             `json:"reason"`
	Diffs      []GameCSVFieldDiff `json:"diffs"`
	Payload    GameCSVRowInput    `json:"payload"`
}

type GameCSVPreview struct {
	Operations    []GameCSVOperation `json:"operations"`
	Creatable     int                `json:"creatable"`
	Updatable     int                `json:"updatable"`
	Skipped       int                `json:"skipped"`
	SelectableAll int                `json:"selectableAll"`
}

// BuildCSVPreview CSV入力行から、実行可能な更新・登録内容の差分プレビューを作成する
func (u *GameUseCase) BuildCSVPreview(ctx context.Context, rows []GameCSVRowInput) (*GameCSVPreview, error) {
	ids := make([]int64, 0, len(rows))
	for _, row := range rows {
		if row.ID != nil && *row.ID > 0 {
			ids = append(ids, *row.ID)
		}
	}

	existingList, err := u.gameRepo.FindByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}

	existingByID := map[int64]model.Game{}
	for _, item := range existingList {
		existingByID[item.ID] = item
	}

	preview := &GameCSVPreview{Operations: make([]GameCSVOperation, 0, len(rows))}

	for _, row := range rows {
		op := GameCSVOperation{
			RowNumber:  row.RowNumber,
			ID:         row.ID,
			Action:     "skip",
			Selectable: false,
			Diffs:      make([]GameCSVFieldDiff, 0),
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
			missing := missingRequiredForGameCreate(row)
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

		updates, diffs, buildErr := buildGameUpdateMapAndDiff(existing, row)
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
func (u *GameUseCase) ApplyCSVOperations(ctx context.Context, operations []GameCSVOperation) (int, int, error) {
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
			missing := missingRequiredForGameCreate(op.Payload)
			if len(missing) > 0 {
				continue
			}

			code := strings.TrimSpace(valueOrEmpty(op.Payload.Code))
			if err := validateDuplicateCode(ctx, code, 0, func(ctx context.Context, code string) (*model.Game, error) {
				return u.gameRepo.FindByCode(ctx, code)
			}); err != nil {
				return created, updated, fmt.Errorf("row %d: %w", op.RowNumber, err)
			}

			releaseDate, err := parseCSVDate(valueOrEmpty(op.Payload.ReleaseDate))
			if err != nil {
				return created, updated, fmt.Errorf("row %d: release_date: %w", op.RowNumber, err)
			}

			listPrice := int32(0)
			if op.Payload.ListPrice != nil {
				listPrice = *op.Payload.ListPrice
			}

			item := &model.Game{
				ID:              *op.ID,
				Name:            strings.TrimSpace(valueOrEmpty(op.Payload.Name)),
				Kana:            strings.TrimSpace(valueOrEmpty(op.Payload.Kana)),
				Overview:        strings.TrimSpace(valueOrEmpty(op.Payload.Overview)),
				Code:            code,
				ManufacturerID:  valueInt64OrZero(op.Payload.ManufacturerID),
				MachineID:       valueInt64OrZero(op.Payload.MachineID),
				GenreID:         valueInt64OrZero(op.Payload.GenreID),
				SubGenre:        strings.TrimSpace(valueOrEmpty(op.Payload.SubGenre)),
				CatchCopy:       strings.TrimSpace(valueOrEmpty(op.Payload.CatchCopy)),
				SubCatch:        strings.TrimSpace(valueOrEmpty(op.Payload.SubCatch)),
				ListPrice:       listPrice,
				ReleaseDate:     releaseDate,
				OfficialSiteURL: strings.TrimSpace(valueOrEmpty(op.Payload.OfficialSiteURL)),
				YouTubeURL:      strings.TrimSpace(valueOrEmpty(op.Payload.YouTubeURL)),
				IsPlay:          valueBoolOrFalse(op.Payload.IsPlay),
				IsClear:         valueBoolOrFalse(op.Payload.IsClear),
				IsFavourite:     valueBoolOrFalse(op.Payload.IsFavourite),
			}

			if err := u.gameRepo.Create(ctx, item); err != nil {
				return created, updated, fmt.Errorf("row %d create failed: %w", op.RowNumber, err)
			}
			created++

		case "update":
			existing, err := u.gameRepo.FindByID(ctx, *op.ID)
			if err != nil {
				return created, updated, fmt.Errorf("row %d: %w", op.RowNumber, err)
			}
			if existing == nil {
				continue
			}

			updates, _, buildErr := buildGameUpdateMapAndDiff(*existing, op.Payload)
			if buildErr != nil {
				return created, updated, fmt.Errorf("row %d: %w", op.RowNumber, buildErr)
			}
			if len(updates) == 0 {
				continue
			}

			if codeRaw, ok := updates["code"]; ok {
				code, _ := codeRaw.(string)
				if err := validateDuplicateCode(ctx, code, *op.ID, func(ctx context.Context, code string) (*model.Game, error) {
					return u.gameRepo.FindByCode(ctx, code)
				}); err != nil {
					return created, updated, fmt.Errorf("row %d: %w", op.RowNumber, err)
				}
			}

			if err := u.gameRepo.UpdateFieldsByID(ctx, *op.ID, updates); err != nil {
				return created, updated, fmt.Errorf("row %d update failed: %w", op.RowNumber, err)
			}
			updated++
		}
	}

	return created, updated, nil
}

func missingRequiredForGameCreate(row GameCSVRowInput) []string {
	missing := make([]string, 0, 18)
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
	if row.ManufacturerID == nil || *row.ManufacturerID < 1 {
		missing = append(missing, "manufacturer_id")
	}
	if row.MachineID == nil || *row.MachineID < 1 {
		missing = append(missing, "machine_id")
	}
	if row.GenreID == nil || *row.GenreID < 1 {
		missing = append(missing, "genre_id")
	}
	if strings.TrimSpace(valueOrEmpty(row.SubGenre)) == "" {
		missing = append(missing, "sub_genre")
	}
	if strings.TrimSpace(valueOrEmpty(row.CatchCopy)) == "" {
		missing = append(missing, "catch_copy")
	}
	if strings.TrimSpace(valueOrEmpty(row.SubCatch)) == "" {
		missing = append(missing, "sub_catch")
	}
	if row.ListPrice == nil {
		missing = append(missing, "list_price")
	}
	if strings.TrimSpace(valueOrEmpty(row.ReleaseDate)) == "" {
		missing = append(missing, "release_date")
	}
	if strings.TrimSpace(valueOrEmpty(row.OfficialSiteURL)) == "" {
		missing = append(missing, "official_site_url")
	}
	if strings.TrimSpace(valueOrEmpty(row.YouTubeURL)) == "" {
		missing = append(missing, "youtube_url")
	}
	if row.IsPlay == nil {
		missing = append(missing, "is_play")
	}
	if row.IsClear == nil {
		missing = append(missing, "is_clear")
	}
	if row.IsFavourite == nil {
		missing = append(missing, "is_favourite")
	}
	return missing
}

func buildGameUpdateMapAndDiff(existing model.Game, row GameCSVRowInput) (map[string]any, []GameCSVFieldDiff, error) {
	updates := map[string]any{}
	diffs := make([]GameCSVFieldDiff, 0, 18)

	if row.Name != nil {
		next := strings.TrimSpace(*row.Name)
		if next != "" && normalizeComparableValue(next) != normalizeComparableValue(existing.Name) {
			updates["name"] = next
			diffs = append(diffs, GameCSVFieldDiff{Field: "name", From: existing.Name, To: next})
		}
	}

	if row.Kana != nil {
		next := strings.TrimSpace(*row.Kana)
		if next != "" && normalizeComparableValue(next) != normalizeComparableValue(existing.Kana) {
			updates["kana"] = next
			diffs = append(diffs, GameCSVFieldDiff{Field: "kana", From: existing.Kana, To: next})
		}
	}

	if row.Overview != nil {
		next := strings.TrimSpace(*row.Overview)
		if next != "" && normalizeComparableValue(next) != normalizeComparableValue(existing.Overview) {
			updates["overview"] = next
			diffs = append(diffs, GameCSVFieldDiff{Field: "overview", From: existing.Overview, To: next})
		}
	}

	if row.Code != nil {
		next := strings.TrimSpace(*row.Code)
		if next != "" && normalizeComparableValue(next) != normalizeComparableValue(existing.Code) {
			updates["code"] = next
			diffs = append(diffs, GameCSVFieldDiff{Field: "code", From: existing.Code, To: next})
		}
	}

	if row.ManufacturerID != nil {
		next := *row.ManufacturerID
		if next > 0 && next != existing.ManufacturerID {
			updates["manufacturer_id"] = next
			diffs = append(diffs, GameCSVFieldDiff{Field: "manufacturer_id", From: strconv.FormatInt(existing.ManufacturerID, 10), To: strconv.FormatInt(next, 10)})
		}
	}

	if row.MachineID != nil {
		next := *row.MachineID
		if next > 0 && next != existing.MachineID {
			updates["machine_id"] = next
			diffs = append(diffs, GameCSVFieldDiff{Field: "machine_id", From: strconv.FormatInt(existing.MachineID, 10), To: strconv.FormatInt(next, 10)})
		}
	}

	if row.GenreID != nil {
		next := *row.GenreID
		if next > 0 && next != existing.GenreID {
			updates["genre_id"] = next
			diffs = append(diffs, GameCSVFieldDiff{Field: "genre_id", From: strconv.FormatInt(existing.GenreID, 10), To: strconv.FormatInt(next, 10)})
		}
	}

	if row.SubGenre != nil {
		next := strings.TrimSpace(*row.SubGenre)
		if next != "" && normalizeComparableValue(next) != normalizeComparableValue(existing.SubGenre) {
			updates["sub_genre"] = next
			diffs = append(diffs, GameCSVFieldDiff{Field: "sub_genre", From: existing.SubGenre, To: next})
		}
	}

	if row.CatchCopy != nil {
		next := strings.TrimSpace(*row.CatchCopy)
		if next != "" && normalizeComparableValue(next) != normalizeComparableValue(existing.CatchCopy) {
			updates["catch_copy"] = next
			diffs = append(diffs, GameCSVFieldDiff{Field: "catch_copy", From: existing.CatchCopy, To: next})
		}
	}

	if row.SubCatch != nil {
		next := strings.TrimSpace(*row.SubCatch)
		if next != "" && normalizeComparableValue(next) != normalizeComparableValue(existing.SubCatch) {
			updates["sub_catch"] = next
			diffs = append(diffs, GameCSVFieldDiff{Field: "sub_catch", From: existing.SubCatch, To: next})
		}
	}

	if row.ListPrice != nil {
		next := *row.ListPrice
		if next != existing.ListPrice {
			updates["list_price"] = next
			diffs = append(diffs, GameCSVFieldDiff{Field: "list_price", From: strconv.FormatInt(int64(existing.ListPrice), 10), To: strconv.FormatInt(int64(next), 10)})
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
				diffs = append(diffs, GameCSVFieldDiff{Field: "release_date", From: current, To: next})
			}
		}
	}

	if row.OfficialSiteURL != nil {
		next := strings.TrimSpace(*row.OfficialSiteURL)
		if next != "" && normalizeComparableValue(next) != normalizeComparableValue(existing.OfficialSiteURL) {
			updates["official_site_url"] = next
			diffs = append(diffs, GameCSVFieldDiff{Field: "official_site_url", From: existing.OfficialSiteURL, To: next})
		}
	}

	if row.YouTubeURL != nil {
		next := strings.TrimSpace(*row.YouTubeURL)
		if next != "" && normalizeComparableValue(next) != normalizeComparableValue(existing.YouTubeURL) {
			updates["youtube_url"] = next
			diffs = append(diffs, GameCSVFieldDiff{Field: "youtube_url", From: existing.YouTubeURL, To: next})
		}
	}

	if row.IsPlay != nil {
		next := *row.IsPlay
		if next != existing.IsPlay {
			updates["is_play"] = next
			diffs = append(diffs, GameCSVFieldDiff{Field: "is_play", From: strconv.FormatBool(existing.IsPlay), To: strconv.FormatBool(next)})
		}
	}

	if row.IsClear != nil {
		next := *row.IsClear
		if next != existing.IsClear {
			updates["is_clear"] = next
			diffs = append(diffs, GameCSVFieldDiff{Field: "is_clear", From: strconv.FormatBool(existing.IsClear), To: strconv.FormatBool(next)})
		}
	}

	if row.IsFavourite != nil {
		next := *row.IsFavourite
		if next != existing.IsFavourite {
			updates["is_favourite"] = next
			diffs = append(diffs, GameCSVFieldDiff{Field: "is_favourite", From: strconv.FormatBool(existing.IsFavourite), To: strconv.FormatBool(next)})
		}
	}

	return updates, diffs, nil
}

func valueBoolOrFalse(v *bool) bool {
	if v == nil {
		return false
	}
	return *v
}

// -------------------------------------------------------------------------
// D: Delete (削除)
// -------------------------------------------------------------------------

// DeleteMachine 機種IDを指定して、該当する機種情報を削除する
func (u *GameUseCase) DeleteGame(ctx context.Context, id int64) error {
	return u.gameRepo.Delete(ctx, id)
}
