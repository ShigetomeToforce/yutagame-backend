package admin

import (
	"context"
	"fmt"
	"yutagame-backend/domain/model"
	"yutagame-backend/infrastructure/database"
)

type GameRankingListItem struct {
	GameID       int64   `json:"gameId"`
	Rank         int     `json:"rank"`
	Code         string  `json:"code"`
	Name         string  `json:"name"`
	Kana         string  `json:"kana"`
	ImageKey     *string `json:"imageKey,omitempty"`
	Price        int32   `json:"price"`
	ReleaseAt    string  `json:"releaseDate"`
	Manufacturer string  `json:"manufacturerName"`
	Machine      string  `json:"machineName"`
	Genre        string  `json:"genreName"`
}

type GameRankingUseCase struct {
	gameRepo    *database.GameRepository
	rankingRepo *database.GameRankingRepository
}

func NewGameRankingUseCase(gameRepo *database.GameRepository, rankingRepo *database.GameRankingRepository) *GameRankingUseCase {
	return &GameRankingUseCase{gameRepo: gameRepo, rankingRepo: rankingRepo}
}

func (u *GameRankingUseCase) GetCurrent(ctx context.Context) ([]GameRankingListItem, string, error) {
	draft, err := u.rankingRepo.GetDraft(ctx)
	if err != nil {
		return nil, "draft", err
	}
	if len(draft) > 0 {
		items, err := u.mapDraftEntries(draft)
		return items, "draft", err
	}

	active, err := u.rankingRepo.GetActive(ctx)
	if err != nil {
		return nil, "active", err
	}
	if len(active) > 0 {
		items, err := u.mapActiveEntries(active)
		return items, "active", err
	}

	defaultItems, err := u.GetDefaultOrder(ctx)
	if err != nil {
		return nil, "active", err
	}
	return defaultItems, "active", nil
}

func (u *GameRankingUseCase) GetDraft(ctx context.Context) ([]GameRankingListItem, error) {
	entries, err := u.rankingRepo.GetDraft(ctx)
	if err != nil {
		return nil, err
	}
	if len(entries) == 0 {
		return u.GetDefaultOrder(ctx)
	}
	return u.mapDraftEntries(entries)
}

func (u *GameRankingUseCase) GetActive(ctx context.Context) ([]GameRankingListItem, error) {
	entries, err := u.rankingRepo.GetActive(ctx)
	if err != nil {
		return nil, err
	}
	if len(entries) == 0 {
		return u.GetDefaultOrder(ctx)
	}
	return u.mapActiveEntries(entries)
}

func (u *GameRankingUseCase) SaveDraft(ctx context.Context, gameIDs []int64) ([]GameRankingListItem, error) {
	allGames, err := u.gameRepo.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	order := make([]int64, 0, len(allGames))
	seen := make(map[int64]struct{}, len(allGames))
	for _, id := range gameIDs {
		if _, exists := seen[id]; exists || id <= 0 {
			continue
		}
		order = append(order, id)
		seen[id] = struct{}{}
	}
	database.SortGamesByReleaseDateAsc(allGames)
	for _, game := range allGames {
		if _, exists := seen[game.ID]; exists {
			continue
		}
		order = append(order, game.ID)
	}
	if err := u.rankingRepo.ReplaceDraft(ctx, order); err != nil {
		return nil, err
	}
	return u.GetDraft(ctx)
}

func (u *GameRankingUseCase) PublishDraft(ctx context.Context) ([]GameRankingListItem, error) {
	if err := u.rankingRepo.CopyDraftToActive(ctx); err != nil {
		return nil, err
	}
	return u.GetActive(ctx)
}

func (u *GameRankingUseCase) GetDefaultOrder(ctx context.Context) ([]GameRankingListItem, error) {
	games, err := u.gameRepo.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	database.SortGamesByReleaseDateAsc(games)
	result := make([]GameRankingListItem, 0, len(games))
	for idx, game := range games {
		result = append(result, u.buildListItem(game, idx+1))
	}
	return result, nil
}

func (u *GameRankingUseCase) buildListItem(game model.Game, rank int) GameRankingListItem {
	manufacturerName := ""
	if game.Manufacturer != nil {
		manufacturerName = game.Manufacturer.Name
	}
	machineName := ""
	if game.Machine != nil {
		machineName = game.Machine.Name
	}
	genreName := ""
	if game.Genre != nil {
		genreName = game.Genre.Name
	}
	return GameRankingListItem{
		GameID:       game.ID,
		Rank:         rank,
		Code:         game.Code,
		Name:         game.Name,
		Kana:         game.Kana,
		ImageKey:     game.ImageKey,
		Price:        game.ListPrice,
		ReleaseAt:    game.ReleaseDate.Format("2006-01-02"),
		Manufacturer: manufacturerName,
		Machine:      machineName,
		Genre:        genreName,
	}
}

func (u *GameRankingUseCase) mapDraftEntries(entries []model.GameRankingDraftEntry) ([]GameRankingListItem, error) {
	items := make([]GameRankingListItem, 0, len(entries))
	byGameID := make(map[int64]model.Game, len(entries))
	ids := make([]int64, 0, len(entries))
	for _, entry := range entries {
		ids = append(ids, entry.GameID)
		if entry.Game != nil {
			byGameID[entry.GameID] = *entry.Game
		}
	}
	if len(ids) > 0 {
		games, err := u.gameRepo.FindByIDsWithRelations(context.Background(), ids)
		if err != nil {
			return nil, err
		}
		for _, game := range games {
			byGameID[game.ID] = game
		}
	}
	for _, entry := range entries {
		if game, ok := byGameID[entry.GameID]; ok {
			items = append(items, u.buildListItem(game, entry.Rank))
		} else {
			items = append(items, GameRankingListItem{
				GameID: entry.GameID,
				Rank:   entry.Rank,
			})
		}
	}
	return items, nil
}

func (u *GameRankingUseCase) mapActiveEntries(entries []model.GameRankingActiveEntry) ([]GameRankingListItem, error) {
	items := make([]GameRankingListItem, 0, len(entries))
	byGameID := make(map[int64]model.Game, len(entries))
	ids := make([]int64, 0, len(entries))
	for _, entry := range entries {
		ids = append(ids, entry.GameID)
		if entry.Game != nil {
			byGameID[entry.GameID] = *entry.Game
		}
	}
	if len(ids) > 0 {
		games, err := u.gameRepo.FindByIDsWithRelations(context.Background(), ids)
		if err != nil {
			return nil, err
		}
		for _, game := range games {
			byGameID[game.ID] = game
		}
	}
	for _, entry := range entries {
		if game, ok := byGameID[entry.GameID]; ok {
			items = append(items, u.buildListItem(game, entry.Rank))
		} else {
			items = append(items, GameRankingListItem{
				GameID: entry.GameID,
				Rank:   entry.Rank,
			})
		}
	}
	return items, nil
}

func (u *GameRankingUseCase) GetPublicRanking(ctx context.Context) ([]GameRankingPublicItem, error) {
	items, err := u.GetActive(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]GameRankingPublicItem, 0, len(items))
	for _, item := range items {
		result = append(result, GameRankingPublicItem{
			Rank:             item.Rank,
			GameID:           item.GameID,
			Code:             item.Code,
			Name:             item.Name,
			ImageKey:         item.ImageKey,
			ManufacturerName: item.Manufacturer,
			MachineName:      item.Machine,
			GenreName:        item.Genre,
			ReleaseDate:      item.ReleaseAt,
			Price:            item.Price,
		})
	}
	return result, nil
}

type GameRankingPublicItem struct {
	Rank             int     `json:"rank"`
	GameID           int64   `json:"gameId"`
	Code             string  `json:"code"`
	Name             string  `json:"name"`
	ImageKey         *string `json:"imageKey,omitempty"`
	ManufacturerName string  `json:"manufacturerName"`
	MachineName      string  `json:"machineName"`
	GenreName        string  `json:"genreName"`
	ReleaseDate      string  `json:"releaseDate"`
	Price            int32   `json:"price"`
}

func (u *GameRankingUseCase) ValidateGameIDs(gameIDs []int64) error {
	if len(gameIDs) == 0 {
		return fmt.Errorf("ランキング対象のゲームがありません")
	}
	all, err := u.gameRepo.FindAll(context.Background())
	if err != nil {
		return err
	}
	allowed := map[int64]struct{}{}
	for _, game := range all {
		allowed[game.ID] = struct{}{}
	}
	for _, id := range gameIDs {
		if _, ok := allowed[id]; !ok {
			return fmt.Errorf("存在しないゲームIDが含まれています: %d", id)
		}
	}
	return nil
}
