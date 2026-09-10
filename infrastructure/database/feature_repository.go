package database

import (
	"context"
	"errors"
	"sort"
	"time"
	"yutagame-backend/domain/model"

	"gorm.io/gorm"
)

type FeatureRepository struct {
	db *gorm.DB
}

func NewFeatureRepository(db *gorm.DB) *FeatureRepository {
	return &FeatureRepository{db: db}
}

const featurePublishedOrderClause = "CASE WHEN display_order > 0 THEN 0 ELSE 1 END, display_order ASC, COALESCE(published_at, created_at) desc, id desc"

func preloadFeatureRelations(db *gorm.DB) *gorm.DB {
	return db.Preload("FeatureGames", func(db *gorm.DB) *gorm.DB {
		return db.Order("display_order asc, id asc")
	}).
		Preload("FeatureGames.Game.Manufacturer").
		Preload("FeatureGames.Game.Machine").
		Preload("FeatureGames.Game.Genre").
		Preload("FeatureGames.Game.Keywords")
}

func attachFeatureGames(feature *model.Feature) {
	if feature == nil {
		return
	}
	games := make([]model.Game, 0, len(feature.FeatureGames))
	for _, item := range feature.FeatureGames {
		if item.Game != nil {
			games = append(games, *item.Game)
		}
	}
	feature.Games = games
}

func attachFeatureGamesList(features []model.Feature) {
	for i := range features {
		attachFeatureGames(&features[i])
	}
}

func (r *FeatureRepository) Create(ctx context.Context, feature *model.Feature) error {
	return r.db.WithContext(ctx).Create(feature).Error
}

func (r *FeatureRepository) FindByID(ctx context.Context, id int64) (*model.Feature, error) {
	var feature model.Feature
	err := preloadFeatureRelations(r.db.WithContext(ctx)).First(&feature, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	attachFeatureGames(&feature)
	return &feature, err
}

func (r *FeatureRepository) FindByCode(ctx context.Context, code string) (*model.Feature, error) {
	var feature model.Feature
	err := r.db.WithContext(ctx).Where("code = ?", code).First(&feature).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &feature, err
}

func (r *FeatureRepository) FindPublishedByCode(ctx context.Context, code string) (*model.Feature, error) {
	var feature model.Feature
	now := time.Now()
	err := preloadFeatureRelations(r.db.WithContext(ctx)).
		Where("code = ? AND status = ?", code, "PUBLISHED").
		Where("publish_start_at IS NULL OR publish_start_at <= ?", now).
		Where("publish_end_at IS NULL OR publish_end_at >= ?", now).
		First(&feature).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	attachFeatureGames(&feature)
	return &feature, err
}

func (r *FeatureRepository) FindAll(ctx context.Context) ([]model.Feature, error) {
	var features []model.Feature
	err := preloadFeatureRelations(r.db.WithContext(ctx)).Order("COALESCE(published_at, created_at) desc, id desc").Find(&features).Error
	attachFeatureGamesList(features)
	return features, err
}

func (r *FeatureRepository) FindPublishedForOrdering(ctx context.Context) ([]model.Feature, error) {
	var features []model.Feature
	err := r.db.WithContext(ctx).
		Where("status = ?", "PUBLISHED").
		Order(featurePublishedOrderClause).
		Find(&features).Error
	return features, err
}

func (r *FeatureRepository) FindAllWithPagination(
	ctx context.Context,
	limit, offset int,
	whereQueries ...func(*gorm.DB) *gorm.DB,
) ([]model.Feature, error) {
	modifier := func(db *gorm.DB) *gorm.DB {
		return preloadFeatureRelations(db)
	}
	features, err := ExecuteFindWithPagination[model.Feature](ctx, r.db, limit, offset, "COALESCE(published_at, created_at) desc, id desc", modifier, whereQueries...)
	attachFeatureGamesList(features)
	return features, err
}

func (r *FeatureRepository) CountAll(ctx context.Context, whereQueries ...func(*gorm.DB) *gorm.DB) (int64, error) {
	return ExecuteCount[model.Feature](ctx, r.db, whereQueries...)
}

func (r *FeatureRepository) Update(ctx context.Context, feature *model.Feature) error {
	return r.db.WithContext(ctx).Model(&model.Feature{}).Where("id = ?", feature.ID).Updates(map[string]interface{}{
		"code":                feature.Code,
		"title":               feature.Title,
		"excerpt":             feature.Excerpt,
		"body_html":           feature.BodyHTML,
		"thumbnail_image_key": feature.ThumbnailImageKey,
		"status":              feature.Status,
		"published_at":        feature.PublishedAt,
		"publish_start_at":    feature.PublishStartAt,
		"publish_end_at":      feature.PublishEndAt,
	}).Error
}

func (r *FeatureRepository) UpdateThumbnailImage(ctx context.Context, id int64, imageKey string) (*model.Feature, error) {
	if err := r.db.WithContext(ctx).Model(&model.Feature{}).Where("id = ?", id).Update("thumbnail_image_key", imageKey).Error; err != nil {
		return nil, err
	}
	return r.FindByID(ctx, id)
}

func (r *FeatureRepository) ReplaceGames(ctx context.Context, featureID int64, gameIDs []int64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("feature_id = ?", featureID).Delete(&model.FeatureGame{}).Error; err != nil {
			return err
		}
		for i, gameID := range gameIDs {
			if err := tx.Create(&model.FeatureGame{FeatureID: featureID, GameID: gameID, DisplayOrder: i + 1}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *FeatureRepository) UpdateOrder(ctx context.Context, ids []int64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for i, id := range ids {
			if err := tx.Model(&model.Feature{}).Where("id = ?", id).Update("display_order", i+1).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *FeatureRepository) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&model.Feature{}, id).Error
}

func (r *FeatureRepository) FindPublishedAll(ctx context.Context) ([]model.Feature, error) {
	var features []model.Feature
	now := time.Now()
	err := preloadFeatureRelations(r.db.WithContext(ctx)).
		Where("status = ?", "PUBLISHED").
		Where("publish_start_at IS NULL OR publish_start_at <= ?", now).
		Where("publish_end_at IS NULL OR publish_end_at >= ?", now).
		Order(featurePublishedOrderClause).
		Find(&features).Error
	attachFeatureGamesList(features)
	return features, err
}

func (r *FeatureRepository) FindPublishedCodes(ctx context.Context) ([]string, error) {
	var codes []string
	now := time.Now()
	err := r.db.WithContext(ctx).
		Model(&model.Feature{}).
		Where("status = ?", "PUBLISHED").
		Where("publish_start_at IS NULL OR publish_start_at <= ?", now).
		Where("publish_end_at IS NULL OR publish_end_at >= ?", now).
		Order(featurePublishedOrderClause).
		Pluck("code", &codes).Error
	return codes, err
}

func sortFeaturesByOrder(features []model.Feature) {
	sort.SliceStable(features, func(i, j int) bool {
		return features[i].DisplayOrder < features[j].DisplayOrder
	})
}
