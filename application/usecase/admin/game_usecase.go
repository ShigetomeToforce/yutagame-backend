package admin

import (
	"context"
	"yutagame-backend/domain/model"
	"yutagame-backend/infrastructure/database"
)

type GameUseCase struct {
	gameRepo *database.GameRepository
}

func NewGameUseCase(gameRepo *database.GameRepository) *GameUseCase {
	return &GameUseCase{gameRepo: gameRepo}
}

func (u *GameUseCase) SearchGames(ctx context.Context, machineID, manufacturerID, keywordID int64) ([]model.Game, error) {
	return u.gameRepo.FindAll(ctx, machineID, manufacturerID, keywordID)
}

func (u *GameUseCase) GetGameByID(ctx context.Context, id int64) (*model.Game, error) {
	return u.gameRepo.FindByID(ctx, id)
}

func (u *GameUseCase) CreateGame(ctx context.Context, g *model.Game) error {
	return u.gameRepo.Create(ctx, g)
}

func (u *GameUseCase) UpdateGame(ctx context.Context, g *model.Game) error {
	return u.gameRepo.Update(ctx, g)
}

func (u *GameUseCase) DeleteGame(ctx context.Context, id int64) error {
	return u.gameRepo.Delete(ctx, id)
}
