package app

import (
	"context"
	"strings"
	"time"
	"yutagame-backend/application/usecase"
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
}

type TopContents struct {
	ReleaseToday     []model.Game `json:"releaseToday"`
	RecentlyReleased []model.Game `json:"recentlyReleased"`
	RecentlyUpdated  []model.Game `json:"recentlyUpdated"`
	RandomPicks      []model.Game `json:"randomPicks"`
}

type GamePublicUseCase struct {
	gameRepo         *database.GameRepository
	machineRepo      *database.MachineRepository
	genreRepo        *database.GenreRepository
	manufacturerRepo *database.ManufacturerRepository
	keywordRepo      *database.KeywordRepository
}

func NewGamePublicUseCase(
	gameRepo *database.GameRepository,
	machineRepo *database.MachineRepository,
	genreRepo *database.GenreRepository,
	manufacturerRepo *database.ManufacturerRepository,
	keywordRepo *database.KeywordRepository,
) *GamePublicUseCase {
	return &GamePublicUseCase{
		gameRepo:         gameRepo,
		machineRepo:      machineRepo,
		genreRepo:        genreRepo,
		manufacturerRepo: manufacturerRepo,
		keywordRepo:      keywordRepo,
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

	return usecase.ExecutePaginatedSearch(
		ctx,
		page,
		limit,
		whereQuery,
		u.gameRepo.CountAll,
		u.gameRepo.FindAllWithPagination,
	)
}

func (u *GamePublicUseCase) GetGameByCode(ctx context.Context, code string) (*model.Game, error) {
	return u.gameRepo.FindByCode(ctx, strings.TrimSpace(code))
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

	return &TopContents{
		ReleaseToday:     releaseToday,
		RecentlyReleased: recentlyReleased,
		RecentlyUpdated:  recentlyUpdated,
		RandomPicks:      randomPicks,
	}, nil
}
