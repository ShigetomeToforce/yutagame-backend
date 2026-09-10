package database

import (
	"context"
	"errors"
	"yutagame-backend/domain/model"

	"gorm.io/gorm"
)

const gameRecommendationOrder = "CASE status WHEN 'NEW' THEN 1 WHEN 'IN_PROGRESS' THEN 2 WHEN 'DONE' THEN 3 ELSE 4 END asc, updated_at desc, id desc"

type GameRecommendationRepository struct{ db *gorm.DB }

func NewGameRecommendationRepository(db *gorm.DB) *GameRecommendationRepository {
	return &GameRecommendationRepository{db: db}
}

func (r *GameRecommendationRepository) Create(ctx context.Context, item *model.GameRecommendation) error {
	return r.db.WithContext(ctx).Create(item).Error
}

func (r *GameRecommendationRepository) FindByID(ctx context.Context, id int64) (*model.GameRecommendation, error) {
	var item model.GameRecommendation
	err := r.db.WithContext(ctx).First(&item, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &item, err
}

func (r *GameRecommendationRepository) FindAll(ctx context.Context) ([]model.GameRecommendation, error) {
	var items []model.GameRecommendation
	err := r.db.WithContext(ctx).Order(gameRecommendationOrder).Find(&items).Error
	return items, err
}

func (r *GameRecommendationRepository) FindAllWithPagination(ctx context.Context, limit, offset int, whereQueries ...func(*gorm.DB) *gorm.DB) ([]model.GameRecommendation, error) {
	return ExecuteFindWithPagination[model.GameRecommendation](ctx, r.db, limit, offset, gameRecommendationOrder, nil, whereQueries...)
}

func (r *GameRecommendationRepository) CountAll(ctx context.Context, whereQueries ...func(*gorm.DB) *gorm.DB) (int64, error) {
	return ExecuteCount[model.GameRecommendation](ctx, r.db, whereQueries...)
}

func (r *GameRecommendationRepository) Update(ctx context.Context, item *model.GameRecommendation) error {
	return r.db.WithContext(ctx).Save(item).Error
}

func (r *GameRecommendationRepository) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&model.GameRecommendation{}, id).Error
}
