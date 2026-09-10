package database

import (
	"context"
	"sort"
	"time"
	"yutagame-backend/domain/model"

	"gorm.io/gorm"
)

type GameRankingRepository struct {
	db *gorm.DB
}

func NewGameRankingRepository(db *gorm.DB) *GameRankingRepository {
	return &GameRankingRepository{db: db}
}

func (r *GameRankingRepository) ReplaceActive(ctx context.Context, orderedGameIDs []int64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("1 = 1").Delete(&model.GameRankingActiveEntry{}).Error; err != nil {
			return err
		}

		for idx, gameID := range uniqueOrderedIDs(orderedGameIDs) {
			if err := tx.Create(&model.GameRankingActiveEntry{
				GameID:    gameID,
				Rank:      idx + 1,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *GameRankingRepository) ReplaceDraft(ctx context.Context, orderedGameIDs []int64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("1 = 1").Delete(&model.GameRankingDraftEntry{}).Error; err != nil {
			return err
		}

		for idx, gameID := range uniqueOrderedIDs(orderedGameIDs) {
			if err := tx.Create(&model.GameRankingDraftEntry{
				GameID:    gameID,
				Rank:      idx + 1,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *GameRankingRepository) CopyDraftToActive(ctx context.Context) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var draftEntries []model.GameRankingDraftEntry
		if err := tx.Order("display_rank asc, id asc").Find(&draftEntries).Error; err != nil {
			return err
		}
		if err := tx.Where("1 = 1").Delete(&model.GameRankingActiveEntry{}).Error; err != nil {
			return err
		}
		for idx, entry := range draftEntries {
			if err := tx.Create(&model.GameRankingActiveEntry{
				GameID:    entry.GameID,
				Rank:      idx + 1,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *GameRankingRepository) GetActive(ctx context.Context) ([]model.GameRankingActiveEntry, error) {
	var entries []model.GameRankingActiveEntry
	if err := r.db.WithContext(ctx).
		Preload("Game").
		Preload("Game.Manufacturer").
		Preload("Game.Machine").
		Preload("Game.Genre").
		Order("display_rank asc, id asc").
		Find(&entries).Error; err != nil {
		return nil, err
	}
	return entries, nil
}

func (r *GameRankingRepository) GetDraft(ctx context.Context) ([]model.GameRankingDraftEntry, error) {
	var entries []model.GameRankingDraftEntry
	if err := r.db.WithContext(ctx).
		Preload("Game").
		Preload("Game.Manufacturer").
		Preload("Game.Machine").
		Preload("Game.Genre").
		Order("display_rank asc, id asc").
		Find(&entries).Error; err != nil {
		return nil, err
	}
	return entries, nil
}

func uniqueOrderedIDs(ids []int64) []int64 {
	seen := make(map[int64]struct{}, len(ids))
	result := make([]int64, 0, len(ids))
	for _, id := range ids {
		if id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	return result
}

func SortGamesByReleaseDateAsc(games []model.Game) {
	sort.Slice(games, func(i, j int) bool {
		if games[i].ReleaseDate.Equal(games[j].ReleaseDate) {
			return games[i].ID < games[j].ID
		}
		return games[i].ReleaseDate.Before(games[j].ReleaseDate)
	})
}
