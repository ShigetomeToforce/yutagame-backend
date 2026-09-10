package database

import (
	"context"
	"errors"
	"time"
	"yutagame-backend/domain/model"

	"gorm.io/gorm"
)

type GameFavoriteRepository struct {
	db *gorm.DB
}

type FavoriteRankingRow struct {
	GameID int64 `gorm:"column:game_id"`
	Count  int64 `gorm:"column:count"`
}

func NewGameFavoriteRepository(db *gorm.DB) *GameFavoriteRepository {
	return &GameFavoriteRepository{db: db}
}

func (r *GameFavoriteRepository) Create(ctx context.Context, favorite *model.GameFavorite) error {
	return r.db.WithContext(ctx).Create(favorite).Error
}

func (r *GameFavoriteRepository) CountByGameID(ctx context.Context, gameID int64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.GameFavorite{}).
		Where("game_id = ?", gameID).
		Count(&count).Error
	return count, err
}

func (r *GameFavoriteRepository) ExistsByGameIDVisitorID(
	ctx context.Context,
	gameID int64,
	visitorID, favoriteDate string,
) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.GameFavorite{}).
		Where("game_id = ? AND visitor_id = ? AND favorite_date = ?", gameID, visitorID, favoriteDate).
		Count(&count).Error
	return count > 0, err
}

func (r *GameFavoriteRepository) CountRecentByGameIDs(ctx context.Context, since time.Time, gameIDs []int64) (map[int64]int64, error) {
	result := make(map[int64]int64)
	if len(gameIDs) == 0 {
		return result, nil
	}

	var rows []struct {
		GameID int64 `gorm:"column:game_id"`
		Count  int64 `gorm:"column:count"`
	}
	err := r.db.WithContext(ctx).
		Model(&model.GameFavorite{}).
		Select("game_id, COUNT(*) as count").
		Where("created_at >= ? AND game_id IN ?", since, gameIDs).
		Group("game_id").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	for _, row := range rows {
		result[row.GameID] = row.Count
	}
	return result, nil
}

func (r *GameFavoriteRepository) FindTopGameIDsByCount(ctx context.Context, limit int) ([]FavoriteRankingRow, error) {
	if limit < 1 {
		return []FavoriteRankingRow{}, nil
	}

	var rows []FavoriteRankingRow
	err := r.db.WithContext(ctx).
		Model(&model.GameFavorite{}).
		Select("game_id, COUNT(*) as count").
		Group("game_id").
		Order("count desc, game_id desc").
		Limit(limit).
		Scan(&rows).Error
	return rows, err
}

func (r *GameFavoriteRepository) FindTopGameIDsByCountInRange(
	ctx context.Context,
	from, to time.Time,
	limit int,
) ([]FavoriteRankingRow, error) {
	if limit < 1 {
		return []FavoriteRankingRow{}, nil
	}

	var rows []FavoriteRankingRow
	err := r.db.WithContext(ctx).
		Model(&model.GameFavorite{}).
		Select("game_id, COUNT(*) as count").
		Where("created_at >= ? AND created_at < ?", from, to).
		Group("game_id").
		Order("count desc, game_id desc").
		Limit(limit).
		Scan(&rows).Error
	return rows, err
}

func (r *GameFavoriteRepository) FindByID(ctx context.Context, id int64) (*model.GameFavorite, error) {
	var item model.GameFavorite
	err := r.db.WithContext(ctx).First(&item, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &item, err
}
