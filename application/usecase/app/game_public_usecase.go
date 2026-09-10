package app

import (
	"context"
	"strings"
	"time"
	"yutagame-backend/application/usecase"
	adminUseCase "yutagame-backend/application/usecase/admin"
	"yutagame-backend/domain/model"
	"yutagame-backend/infrastructure/database"

	"gorm.io/gorm"
)

type CatalogItem struct {
	Code      string  `json:"code"`
	Name      string  `json:"name"`
	ImageKey  *string `json:"imageKey,omitempty"`
	GameCount int64   `json:"gameCount"`
}

type KeywordItem struct {
	Code      string `json:"code"`
	Name      string `json:"name"`
	GameCount int64  `json:"gameCount"`
}

type GameListFilter struct {
	SearchWord       string
	MachineCode      string
	GenreCode        string
	ManufacturerCode string
	KeywordCode      string
	Sort             string
}

type TopContents struct {
	ReleaseToday     []model.Game          `json:"releaseToday"`
	RecentlyReleased []model.Game          `json:"recentlyReleased"`
	RecentlyUpdated  []model.Game          `json:"recentlyUpdated"`
	RandomPicks      []model.Game          `json:"randomPicks"`
	RankingTop20     []model.Game          `json:"rankingTop20"`
	FavoriteRanking  []FavoriteRankingItem `json:"favoriteRanking"`
}

type FavoriteRankingItem struct {
	Game  model.Game `json:"game"`
	Count int64      `json:"count"`
}

type PublicRankingResponse struct {
	Data       []model.Game `json:"data"`
	TotalCount int64        `json:"totalCount"`
	TotalPages int          `json:"totalPages"`
	Page       int          `json:"page"`
	Limit      int          `json:"limit"`
}

func (u *GamePublicUseCase) applyActiveRanks(ctx context.Context, games []model.Game) error {
	if u.rankingRepo == nil || len(games) == 0 {
		return nil
	}

	ranksByGameID, err := u.rankingRepo.GetActiveRanksByGameID(ctx)
	if err != nil {
		return err
	}
	for index := range games {
		if rank, ok := ranksByGameID[games[index].ID]; ok {
			games[index].Rank = &rank
		}
	}
	return nil
}

type GamePublicUseCase struct {
	gameRepo         *database.GameRepository
	favoriteRepo     *database.GameFavoriteRepository
	machineRepo      *database.MachineRepository
	genreRepo        *database.GenreRepository
	manufacturerRepo *database.ManufacturerRepository
	keywordRepo      *database.KeywordRepository
	rankingRepo      *database.GameRankingRepository
	gameViewLogRepo  *database.GameViewLogRepository
}

func NewGamePublicUseCase(
	gameRepo *database.GameRepository,
	favoriteRepo *database.GameFavoriteRepository,
	machineRepo *database.MachineRepository,
	genreRepo *database.GenreRepository,
	manufacturerRepo *database.ManufacturerRepository,
	keywordRepo *database.KeywordRepository,
	rankingRepo *database.GameRankingRepository,
	gameViewLogRepo *database.GameViewLogRepository,
) *GamePublicUseCase {
	return &GamePublicUseCase{
		gameRepo:         gameRepo,
		favoriteRepo:     favoriteRepo,
		machineRepo:      machineRepo,
		genreRepo:        genreRepo,
		manufacturerRepo: manufacturerRepo,
		keywordRepo:      keywordRepo,
		rankingRepo:      rankingRepo,
		gameViewLogRepo:  gameViewLogRepo,
	}
}

func (u *GamePublicUseCase) GetMachines(ctx context.Context) ([]CatalogItem, error) {
	items, err := u.machineRepo.FindAllWithGameCount(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]CatalogItem, 0, len(items))
	for _, item := range items {
		result = append(result, CatalogItem{
			Code:      item.Code,
			Name:      item.Name,
			ImageKey:  item.ImageKey,
			GameCount: item.GameCount,
		})
	}
	return result, nil
}

func (u *GamePublicUseCase) GetGenres(ctx context.Context) ([]CatalogItem, error) {
	items, err := u.genreRepo.FindAllWithGameCount(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]CatalogItem, 0, len(items))
	for _, item := range items {
		result = append(result, CatalogItem{
			Code:      item.Code,
			Name:      item.Name,
			ImageKey:  item.ImageKey,
			GameCount: item.GameCount,
		})
	}
	return result, nil
}

func (u *GamePublicUseCase) GetManufacturers(ctx context.Context) ([]CatalogItem, error) {
	items, err := u.manufacturerRepo.FindAllWithGameCount(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]CatalogItem, 0, len(items))
	for _, item := range items {
		result = append(result, CatalogItem{
			Code:      item.Code,
			Name:      item.Name,
			ImageKey:  item.ImageKey,
			GameCount: item.GameCount,
		})
	}
	return result, nil
}

func (u *GamePublicUseCase) GetKeywords(ctx context.Context) ([]KeywordItem, error) {
	items, err := u.keywordRepo.FindAllWithGameCount(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]KeywordItem, 0, len(items))
	for _, item := range items {
		result = append(result, KeywordItem{
			Code:      item.Code,
			Name:      item.Name,
			GameCount: item.GameCount,
		})
	}
	return result, nil
}

func (u *GamePublicUseCase) SearchGames(
	ctx context.Context,
	page, limit int,
	filter GameListFilter,
) ([]model.Game, int64, int, error) {
	filter.SearchWord = strings.TrimSpace(filter.SearchWord)
	filter.MachineCode = strings.TrimSpace(filter.MachineCode)
	filter.GenreCode = strings.TrimSpace(filter.GenreCode)
	filter.ManufacturerCode = strings.TrimSpace(filter.ManufacturerCode)
	filter.KeywordCode = strings.TrimSpace(filter.KeywordCode)
	filter.Sort = strings.TrimSpace(filter.Sort)

	var whereQuery func(*gorm.DB) *gorm.DB
	if filter.SearchWord != "" || filter.MachineCode != "" || filter.GenreCode != "" ||
		filter.ManufacturerCode != "" || filter.KeywordCode != "" {
		whereQuery = func(db *gorm.DB) *gorm.DB {
			if filter.SearchWord != "" {
				likeQuery := "%" + filter.SearchWord + "%"
				db = db.Where("games.name LIKE ? OR games.kana LIKE ?", likeQuery, likeQuery)
			}

			if filter.MachineCode != "" {
				db = db.Where(`
					EXISTS (
						SELECT 1 FROM machines
						WHERE machines.id = games.machine_id
						AND machines.code = ?
					)
				`, filter.MachineCode)
			}

			if filter.GenreCode != "" {
				db = db.Where(`
					EXISTS (
						SELECT 1 FROM genres
						WHERE genres.id = games.genre_id
						AND genres.code = ?
					)
				`, filter.GenreCode)
			}

			if filter.ManufacturerCode != "" {
				db = db.Where(`
					EXISTS (
						SELECT 1 FROM manufacturers
						WHERE manufacturers.id = games.manufacturer_id
						AND manufacturers.code = ?
					)
				`, filter.ManufacturerCode)
			}

			if filter.KeywordCode != "" {
				db = db.Where(`
					EXISTS (
						SELECT 1
						FROM game_keywords gk
						INNER JOIN keywords k ON k.id = gk.keyword_id
						WHERE gk.game_id = games.id
						AND k.code = ?
					)
				`, filter.KeywordCode)
			}

			return db
		}
	}

	order := gameSearchOrder(filter.Sort)
	games, totalCount, totalPages, err := usecase.ExecutePaginatedSearch(
		ctx,
		page,
		limit,
		whereQuery,
		u.gameRepo.CountAll,
		func(ctx context.Context, limit, offset int, whereQueries ...func(*gorm.DB) *gorm.DB) ([]model.Game, error) {
			return u.gameRepo.FindAllWithPaginationByOrder(ctx, limit, offset, order, whereQueries...)
		},
	)
	if err != nil {
		return nil, 0, 0, err
	}
	if err := u.applyActiveRanks(ctx, games); err != nil {
		return nil, 0, 0, err
	}
	return games, totalCount, totalPages, nil
}

func gameSearchOrder(sortKey string) string {
	switch sortKey {
	case "release_desc":
		return "games.release_date desc, games.id desc"
	case "kana_asc":
		return "games.kana asc, games.id asc"
	case "price_asc":
		return "games.list_price asc, games.id asc"
	case "price_desc":
		return "games.list_price desc, games.id asc"
	case "rank_asc":
		return "CASE WHEN game_ranking_active_entries.display_rank IS NULL THEN 1 ELSE 0 END asc, game_ranking_active_entries.display_rank asc, games.id asc"
	case "rank_desc":
		return "CASE WHEN game_ranking_active_entries.display_rank IS NULL THEN 1 ELSE 0 END asc, game_ranking_active_entries.display_rank desc, games.id asc"
	default:
		return "games.release_date asc, games.id asc"
	}
}

func (u *GamePublicUseCase) GetGameByCode(ctx context.Context, code string) (*model.Game, error) {
	game, err := u.gameRepo.FindByCode(ctx, strings.TrimSpace(code))
	if err != nil || game == nil {
		return game, err
	}
	if u.rankingRepo == nil {
		return game, nil
	}
	ranksByGameID, err := u.rankingRepo.GetActiveRanksByGameID(ctx)
	if err != nil {
		return nil, err
	}
	if rank, ok := ranksByGameID[game.ID]; ok {
		game.Rank = &rank
	}
	return game, nil
}

func (u *GamePublicUseCase) GetActiveRanking(ctx context.Context) ([]adminUseCase.GameRankingPublicItem, error) {
	if u.rankingRepo == nil {
		return []adminUseCase.GameRankingPublicItem{}, nil
	}
	return adminUseCase.NewGameRankingUseCase(u.gameRepo, u.rankingRepo).GetPublicRanking(ctx)
}

func (u *GamePublicUseCase) GetPublicRankingPage(
	ctx context.Context,
	rankingType, period string,
	page, limit int,
) (*PublicRankingResponse, error) {
	if limit < 1 || limit > 100 {
		limit = 20
	}

	var (
		games []model.Game
		err   error
	)
	switch rankingType {
	case "views":
		games, err = u.getViewRankingGames(ctx, period)
	case "favorites":
		games, err = u.getFavoriteRankingGames(ctx, period)
	default:
		games, err = u.getCuratedRankingGames(ctx)
	}
	if err != nil {
		return nil, err
	}

	totalCount := int64(len(games))
	pagination := usecase.CalculatePagination(totalCount, page, limit)
	start := pagination.Offset
	end := start + limit
	if end > len(games) {
		end = len(games)
	}
	if start > len(games) {
		start = len(games)
	}
	return &PublicRankingResponse{
		Data:       games[start:end],
		TotalCount: totalCount,
		TotalPages: pagination.TotalPages,
		Page:       pagination.ActivePage,
		Limit:      limit,
	}, nil
}

func (u *GamePublicUseCase) getCuratedRankingGames(ctx context.Context) ([]model.Game, error) {
	if u.rankingRepo == nil {
		return []model.Game{}, nil
	}
	entries, err := u.rankingRepo.GetActive(ctx)
	if err != nil {
		return nil, err
	}
	previousRanksByGameID, err := u.rankingRepo.GetPreviousRanksByGameID(ctx)
	if err != nil {
		return nil, err
	}
	games := make([]model.Game, 0, len(entries))
	for _, entry := range entries {
		if entry.Game == nil {
			continue
		}
		game := *entry.Game
		rank := entry.Rank
		game.Rank = &rank
		if previousRank, ok := previousRanksByGameID[game.ID]; ok {
			game.PreviousRank = &previousRank
		}
		games = append(games, game)
	}
	return games, nil
}

func (u *GamePublicUseCase) getFavoriteRankingGames(ctx context.Context, period string) ([]model.Game, error) {
	period = strings.TrimSpace(period)
	if period == "" {
		period = "monthly"
	}

	var rows []database.FavoriteRankingRow
	var err error
	if period == "total" {
		rows, err = u.favoriteRepo.FindTopGameIDsByCount(ctx, 100000)
	} else {
		now := time.Now()
		from := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		to := from.AddDate(0, 1, 0)
		if period == "yearly" {
			from = time.Date(now.Year(), time.January, 1, 0, 0, 0, 0, now.Location())
			to = from.AddDate(1, 0, 0)
		}
		rows, err = u.favoriteRepo.FindTopGameIDsByCountInRange(ctx, from, to, 100000)
	}
	if err != nil {
		return nil, err
	}
	ids := make([]int64, 0, len(rows))
	countsByID := make(map[int64]int64, len(rows))
	for _, row := range rows {
		ids = append(ids, row.GameID)
		countsByID[row.GameID] = row.Count
	}
	return u.rankedGamesByIDs(ctx, ids, countsByID)
}

func (u *GamePublicUseCase) getViewRankingGames(ctx context.Context, period string) ([]model.Game, error) {
	period = strings.TrimSpace(period)
	if period == "" {
		period = "monthly"
	}
	now := time.Now()
	var rows []database.AccessLogTopItem
	var err error
	if period == "total" {
		rows, err = u.gameViewLogRepo.TopByGameCodeAllTime(ctx, 100000)
	} else {
		from := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		to := from.AddDate(0, 1, 0)
		if period == "yearly" {
			from = time.Date(now.Year(), time.January, 1, 0, 0, 0, 0, now.Location())
			to = from.AddDate(1, 0, 0)
		}
		rows, err = u.gameViewLogRepo.TopByGameCode(ctx, from, to, 100000)
	}
	if err != nil {
		return nil, err
	}
	allGames, err := u.gameRepo.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	gamesByCode := make(map[string]model.Game, len(allGames))
	for _, game := range allGames {
		gamesByCode[game.Code] = game
	}
	games := make([]model.Game, 0, len(rows))
	for index, row := range rows {
		game, ok := gamesByCode[row.Value]
		if !ok {
			continue
		}
		rank := index + 1
		count := row.Count
		game.Rank = &rank
		game.RankingCount = &count
		games = append(games, game)
	}
	return games, nil
}

func (u *GamePublicUseCase) rankedGamesByIDs(ctx context.Context, ids []int64, countsByID map[int64]int64) ([]model.Game, error) {
	games, err := u.gameRepo.FindByIDsWithRelations(ctx, ids)
	if err != nil {
		return nil, err
	}
	gamesByID := make(map[int64]model.Game, len(games))
	for _, game := range games {
		gamesByID[game.ID] = game
	}
	result := make([]model.Game, 0, len(ids))
	for index, id := range ids {
		game, ok := gamesByID[id]
		if !ok {
			continue
		}
		rank := index + 1
		count := countsByID[id]
		game.Rank = &rank
		game.RankingCount = &count
		result = append(result, game)
	}
	return result, nil
}

func (u *GamePublicUseCase) GetTopContents(ctx context.Context, releaseLimit, recentLimit, randomLimit int) (*TopContents, error) {
	now := time.Now()

	releaseToday, err := u.gameRepo.FindReleasedOnMonthDay(ctx, int(now.Month()), now.Day(), releaseLimit)
	if err != nil {
		return nil, err
	}

	recentlyReleased, err := u.gameRepo.FindRecentlyReleased(ctx, recentLimit)
	if err != nil {
		return nil, err
	}

	recentlyUpdated, err := u.gameRepo.FindRecentlyUpdated(ctx, recentLimit)
	if err != nil {
		return nil, err
	}

	randomPicks, err := u.gameRepo.FindRandom(ctx, randomLimit)
	if err != nil {
		return nil, err
	}

	rankingTop20 := make([]model.Game, 0, 20)
	if u.rankingRepo != nil {
		entries, err := u.rankingRepo.GetActive(ctx)
		if err != nil {
			return nil, err
		}
		for _, entry := range entries {
			if entry.Game == nil || len(rankingTop20) == 20 {
				continue
			}
			game := *entry.Game
			rank := entry.Rank
			game.Rank = &rank
			rankingTop20 = append(rankingTop20, game)
		}
	}
	for _, games := range [][]model.Game{releaseToday, recentlyReleased, recentlyUpdated, randomPicks} {
		if err := u.applyActiveRanks(ctx, games); err != nil {
			return nil, err
		}
	}

	countsByID := map[int64]int64{}
	rankingRows, err := u.favoriteRepo.FindTopGameIDsByCount(ctx, randomLimit)
	if err != nil {
		return nil, err
	}

	rankedGames := make([]model.Game, 0, len(rankingRows))
	if len(rankingRows) > 0 {
		ids := make([]int64, 0, len(rankingRows))
		for _, row := range rankingRows {
			ids = append(ids, row.GameID)
			countsByID[row.GameID] = row.Count
		}

		games, err := u.gameRepo.FindByIDsWithRelations(ctx, ids)
		if err != nil {
			return nil, err
		}
		gamesByID := make(map[int64]model.Game, len(games))
		for _, game := range games {
			gamesByID[game.ID] = game
		}
		for _, row := range rankingRows {
			if game, ok := gamesByID[row.GameID]; ok {
				rankedGames = append(rankedGames, game)
			}
		}
	}

	favoriteRanking := make([]FavoriteRankingItem, 0, len(rankedGames))
	for _, game := range rankedGames {
		favoriteRanking = append(favoriteRanking, FavoriteRankingItem{
			Game:  game,
			Count: countsByID[game.ID],
		})
	}

	return &TopContents{
		ReleaseToday:     releaseToday,
		RecentlyReleased: recentlyReleased,
		RecentlyUpdated:  recentlyUpdated,
		RandomPicks:      randomPicks,
		RankingTop20:     rankingTop20,
		FavoriteRanking:  favoriteRanking,
	}, nil
}
