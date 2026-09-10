package admin

import (
	"context"
	"errors"
	"strings"
	"time"
	"yutagame-backend/application/usecase"
	"yutagame-backend/domain/model"
	"yutagame-backend/infrastructure/database"

	"gorm.io/gorm"
)

type FeatureListFilter struct {
	SearchWord string
	Status     string
}

type FeatureUseCase struct {
	featureRepo   *database.FeatureRepository
	accessLogRepo *database.ContentAccessLogRepository
}

func NewFeatureUseCase(featureRepo *database.FeatureRepository, accessLogRepo *database.ContentAccessLogRepository) *FeatureUseCase {
	return &FeatureUseCase{featureRepo: featureRepo, accessLogRepo: accessLogRepo}
}

func (u *FeatureUseCase) attachAccessCounts(ctx context.Context, items []model.Feature) []model.Feature {
	if len(items) == 0 || u.accessLogRepo == nil {
		return items
	}
	keys := make([]string, 0, len(items))
	for _, item := range items {
		keys = append(keys, item.Code)
	}
	counts, err := u.accessLogRepo.CountByContentKeys(ctx, "feature", keys)
	if err != nil {
		return items
	}
	for i := range items {
		items[i].AccessCount = counts[items[i].Code]
	}
	return items
}

func normalizeFeatureGameIDs(gameIDs []int64) []int64 {
	result := make([]int64, 0, len(gameIDs))
	seen := map[int64]struct{}{}
	for _, id := range gameIDs {
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

func (u *FeatureUseCase) CreateFeature(ctx context.Context, code, title, excerpt, bodyHTML, thumbnailImageKey, status string, publishStartAt, publishEndAt *time.Time, gameIDs []int64) (*model.Feature, error) {
	code = strings.TrimSpace(code)
	title = strings.TrimSpace(title)
	bodyHTML = strings.TrimSpace(bodyHTML)
	thumbnailImageKey = strings.TrimSpace(thumbnailImageKey)
	status = strings.TrimSpace(status)
	if code == "" || title == "" || bodyHTML == "" {
		return nil, errors.New("code, title and bodyHtml are required")
	}
	if err := validateDuplicateCode(ctx, code, 0, u.featureRepo.FindByCode); err != nil {
		if errors.Is(err, ErrCodeAlreadyExists) {
			return nil, errors.New("指定されたコードは既に使用されています。")
		}
		return nil, err
	}
	if excerpt = strings.TrimSpace(excerpt); excerpt == "" {
		excerpt = trimExcerpt(bodyHTML)
	}
	if status == "" {
		status = "DRAFT"
	}
	if publishStartAt != nil && publishEndAt != nil && publishEndAt.Before(*publishStartAt) {
		return nil, errors.New("publishEndAt must be after publishStartAt")
	}
	var thumbnail *string
	if thumbnailImageKey != "" {
		thumbnail = &thumbnailImageKey
	}
	feature := &model.Feature{Code: code, Title: title, Excerpt: excerpt, BodyHTML: bodyHTML, ThumbnailImageKey: thumbnail, Status: status, PublishStartAt: publishStartAt, PublishEndAt: publishEndAt}
	if status == "PUBLISHED" {
		now := time.Now()
		feature.PublishedAt = &now
	}
	if err := u.featureRepo.Create(ctx, feature); err != nil {
		return nil, err
	}
	if err := u.featureRepo.ReplaceGames(ctx, feature.ID, normalizeFeatureGameIDs(gameIDs)); err != nil {
		return nil, err
	}
	return u.featureRepo.FindByID(ctx, feature.ID)
}

func (u *FeatureUseCase) GetFeatureByID(ctx context.Context, id int64) (*model.Feature, error) {
	return u.featureRepo.FindByID(ctx, id)
}

func (u *FeatureUseCase) GetAllFeatures(ctx context.Context) ([]model.Feature, error) {
	items, err := u.featureRepo.FindAll(ctx)
	return u.attachAccessCounts(ctx, items), err
}

func (u *FeatureUseCase) GetFeaturesWithPagination(ctx context.Context, page, limit int, filter FeatureListFilter) ([]model.Feature, int64, int, error) {
	var whereQuery func(*gorm.DB) *gorm.DB
	searchWord := strings.TrimSpace(filter.SearchWord)
	status := strings.TrimSpace(filter.Status)
	if searchWord != "" || status != "" {
		whereQuery = func(db *gorm.DB) *gorm.DB {
			if searchWord != "" {
				likeQuery := "%" + searchWord + "%"
				db = db.Where("code LIKE ? OR title LIKE ? OR excerpt LIKE ?", likeQuery, likeQuery, likeQuery)
			}
			if status != "" {
				db = db.Where("status = ?", status)
			}
			return db
		}
	}
	items, totalCount, totalPages, err := usecase.ExecutePaginatedSearch(ctx, page, limit, whereQuery, u.featureRepo.CountAll, u.featureRepo.FindAllWithPagination)
	return u.attachAccessCounts(ctx, items), totalCount, totalPages, err
}

func (u *FeatureUseCase) UpdateFeature(ctx context.Context, id int64, code, title, excerpt, bodyHTML, thumbnailImageKey, status string, publishStartAt, publishEndAt *time.Time, gameIDs []int64) (*model.Feature, error) {
	feature, err := u.featureRepo.FindByID(ctx, id)
	if err != nil || feature == nil {
		return nil, errors.New("feature not found")
	}
	code = strings.TrimSpace(code)
	title = strings.TrimSpace(title)
	bodyHTML = strings.TrimSpace(bodyHTML)
	thumbnailImageKey = strings.TrimSpace(thumbnailImageKey)
	status = strings.TrimSpace(status)
	if code == "" || title == "" || bodyHTML == "" {
		return nil, errors.New("code, title and bodyHtml are required")
	}
	if err := validateDuplicateCode(ctx, code, id, u.featureRepo.FindByCode); err != nil {
		if errors.Is(err, ErrCodeAlreadyExists) {
			return nil, errors.New("指定されたコードは既に使用されています。")
		}
		return nil, err
	}
	if excerpt = strings.TrimSpace(excerpt); excerpt == "" {
		excerpt = trimExcerpt(bodyHTML)
	}
	if publishStartAt != nil && publishEndAt != nil && publishEndAt.Before(*publishStartAt) {
		return nil, errors.New("publishEndAt must be after publishStartAt")
	}
	var thumbnail *string
	if thumbnailImageKey != "" {
		thumbnail = &thumbnailImageKey
	}
	feature.Code = code
	feature.Title = title
	feature.Excerpt = excerpt
	feature.BodyHTML = bodyHTML
	feature.ThumbnailImageKey = thumbnail
	feature.Status = status
	feature.PublishStartAt = publishStartAt
	feature.PublishEndAt = publishEndAt
	if status == "PUBLISHED" && feature.PublishedAt == nil {
		now := time.Now()
		feature.PublishedAt = &now
	}
	if err := u.featureRepo.Update(ctx, feature); err != nil {
		return nil, err
	}
	if err := u.featureRepo.ReplaceGames(ctx, id, normalizeFeatureGameIDs(gameIDs)); err != nil {
		return nil, err
	}
	return u.featureRepo.FindByID(ctx, id)
}

func (u *FeatureUseCase) UpdateThumbnailImage(ctx context.Context, id int64, imageKey string) (*model.Feature, error) {
	return u.featureRepo.UpdateThumbnailImage(ctx, id, imageKey)
}

func (u *FeatureUseCase) GetPublishedFeaturesForOrdering(ctx context.Context) ([]model.Feature, error) {
	return u.featureRepo.FindPublishedForOrdering(ctx)
}

func (u *FeatureUseCase) UpdateFeatureOrder(ctx context.Context, ids []int64) error {
	return u.featureRepo.UpdateOrder(ctx, ids)
}

func (u *FeatureUseCase) DeleteFeature(ctx context.Context, id int64) error {
	return u.featureRepo.Delete(ctx, id)
}
