package app

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	"yutagame-backend/domain/model"
	"yutagame-backend/infrastructure/database"
)

type GameFavoriteUseCase struct {
	gameRepo     *database.GameRepository
	favoriteRepo *database.GameFavoriteRepository
}

func NewGameFavoriteUseCase(
	gameRepo *database.GameRepository,
	favoriteRepo *database.GameFavoriteRepository,
) *GameFavoriteUseCase {
	return &GameFavoriteUseCase{gameRepo: gameRepo, favoriteRepo: favoriteRepo}
}

type GameFavoriteResult struct {
	GameID       int64  `json:"gameId"`
	Count        int64  `json:"count"`
	AlreadyVoted bool   `json:"alreadyVoted"`
	FavoriteDate string `json:"favoriteDate"`
}

func (u *GameFavoriteUseCase) GetStatus(
	ctx context.Context,
	gameCode, visitorID string,
) (*GameFavoriteResult, error) {
	gameCode = strings.TrimSpace(gameCode)
	visitorID = strings.TrimSpace(visitorID)
	if gameCode == "" {
		return nil, errors.New("game code is required")
	}

	game, err := u.gameRepo.FindByCode(ctx, gameCode)
	if err != nil {
		return nil, err
	}
	if game == nil {
		return nil, fmt.Errorf("指定されたゲームが見つかりませんでした")
	}

	favoriteDate := time.Now().Format("2006-01-02")

	count, err := u.favoriteRepo.CountByGameID(ctx, game.ID)
	if err != nil {
		return nil, err
	}

	alreadyVoted := false
	if visitorID != "" {
		alreadyVoted, err = u.favoriteRepo.ExistsByGameIDVisitorID(ctx, game.ID, visitorID, favoriteDate)
		if err != nil {
			return nil, err
		}
	}

	return &GameFavoriteResult{
		GameID:       game.ID,
		Count:        count,
		AlreadyVoted: alreadyVoted,
		FavoriteDate: favoriteDate,
	}, nil
}

func (u *GameFavoriteUseCase) PushToday(
	ctx context.Context,
	gameCode, visitorID string,
) (*GameFavoriteResult, error) {
	gameCode = strings.TrimSpace(gameCode)
	visitorID = strings.TrimSpace(visitorID)
	if gameCode == "" {
		return nil, errors.New("game code is required")
	}
	if visitorID == "" {
		return nil, errors.New("visitor id is required")
	}

	game, err := u.gameRepo.FindByCode(ctx, gameCode)
	if err != nil {
		return nil, err
	}
	if game == nil {
		return nil, fmt.Errorf("指定されたゲームが見つかりませんでした")
	}

	favoriteDate := time.Now().Format("2006-01-02")

	exists, err := u.favoriteRepo.ExistsByGameIDVisitorID(ctx, game.ID, visitorID, favoriteDate)
	if err != nil {
		return nil, err
	}

	count, err := u.favoriteRepo.CountByGameID(ctx, game.ID)
	if err != nil {
		return nil, err
	}

	if exists {
		return &GameFavoriteResult{
			GameID:       game.ID,
			Count:        count,
			AlreadyVoted: true,
			FavoriteDate: favoriteDate,
		}, nil
	}

	if err := u.favoriteRepo.Create(ctx, &model.GameFavorite{
		GameID:       game.ID,
		VisitorID:    visitorID,
		FavoriteDate: favoriteDate,
	}); err != nil {
		return nil, err
	}

	count++
	return &GameFavoriteResult{
		GameID:       game.ID,
		Count:        count,
		AlreadyVoted: false,
		FavoriteDate: favoriteDate,
	}, nil
}
