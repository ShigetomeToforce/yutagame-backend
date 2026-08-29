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

// UpdateMachine 既存のゲーム情報を更新する
func (u *GameUseCase) UpdateGame(ctx context.Context, g *model.Game, keywordIDs []int64) error {
	if err := u.gameRepo.Update(ctx, g); err != nil {
		return err
	}

	return u.gameRepo.ReplaceKeywords(ctx, g.ID, keywordIDs)
}

// -------------------------------------------------------------------------
// D: Delete (削除)
// -------------------------------------------------------------------------

// DeleteMachine 機種IDを指定して、該当する機種情報を削除する
func (u *GameUseCase) DeleteGame(ctx context.Context, id int64) error {
	return u.gameRepo.Delete(ctx, id)
}
