package admin

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"
	"yutagame-backend/domain/model"
	"yutagame-backend/infrastructure/database"
)

var bannerPlacements = map[string]bool{
	"top_above": true, "top_below": true, "search_above": true,
	"search_below": true, "ranking_above": true, "ranking_below": true,
	"game_detail_below": true,
}

type BannerUseCase struct {
	bannerRepo    *database.BannerRepository
	accessLogRepo *database.ContentAccessLogRepository
}

func NewBannerUseCase(bannerRepo *database.BannerRepository, accessLogRepo *database.ContentAccessLogRepository) *BannerUseCase {
	return &BannerUseCase{bannerRepo: bannerRepo, accessLogRepo: accessLogRepo}
}

func (u *BannerUseCase) attachAccessCounts(ctx context.Context, items []model.Banner) []model.Banner {
	if len(items) == 0 || u.accessLogRepo == nil {
		return items
	}
	keys := make([]string, 0, len(items))
	for _, item := range items {
		keys = append(keys, strconv.FormatInt(item.ID, 10))
	}
	counts, err := u.accessLogRepo.CountByContentKeys(ctx, "banner", keys)
	if err != nil {
		return items
	}
	for i := range items {
		items[i].AccessCount = counts[strconv.FormatInt(items[i].ID, 10)]
	}
	return items
}

func validateBanner(title, placement string, startsAt, endsAt *time.Time) error {
	if strings.TrimSpace(title) == "" || !bannerPlacements[placement] {
		return errors.New("title and a valid placement are required")
	}
	if startsAt != nil && endsAt != nil && !endsAt.After(*startsAt) {
		return errors.New("endsAt must be later than startsAt")
	}
	return nil
}

func (u *BannerUseCase) Create(ctx context.Context, banner *model.Banner) (*model.Banner, error) {
	banner.Title = strings.TrimSpace(banner.Title)
	banner.Placement = strings.TrimSpace(banner.Placement)
	banner.LinkURL = strings.TrimSpace(banner.LinkURL)
	if err := validateBanner(banner.Title, banner.Placement, banner.StartsAt, banner.EndsAt); err != nil {
		return nil, err
	}
	if banner.DisplayOrder < 1 {
		banner.DisplayOrder = 1
	}
	if err := u.bannerRepo.Create(ctx, banner); err != nil {
		return nil, err
	}
	return banner, nil
}

func (u *BannerUseCase) GetByID(ctx context.Context, id int64) (*model.Banner, error) {
	return u.bannerRepo.FindByID(ctx, id)
}

func (u *BannerUseCase) GetAll(ctx context.Context) ([]model.Banner, error) {
	items, err := u.bannerRepo.FindAll(ctx)
	return u.attachAccessCounts(ctx, items), err
}

func (u *BannerUseCase) Update(ctx context.Context, id int64, update *model.Banner) (*model.Banner, error) {
	banner, err := u.bannerRepo.FindByID(ctx, id)
	if err != nil || banner == nil {
		return nil, errors.New("banner not found")
	}
	update.Title = strings.TrimSpace(update.Title)
	update.Placement = strings.TrimSpace(update.Placement)
	update.LinkURL = strings.TrimSpace(update.LinkURL)
	if err := validateBanner(update.Title, update.Placement, update.StartsAt, update.EndsAt); err != nil {
		return nil, err
	}
	banner.Title, banner.Placement, banner.LinkURL = update.Title, update.Placement, update.LinkURL
	banner.OpenInNewTab, banner.StartsAt, banner.EndsAt = update.OpenInNewTab, update.StartsAt, update.EndsAt
	if err := u.bannerRepo.Update(ctx, banner); err != nil {
		return nil, err
	}
	return banner, nil
}

func (u *BannerUseCase) Delete(ctx context.Context, id int64) (*model.Banner, error) {
	banner, err := u.bannerRepo.FindByID(ctx, id)
	if err != nil || banner == nil {
		return nil, errors.New("banner not found")
	}
	return banner, u.bannerRepo.Delete(ctx, id)
}

func (u *BannerUseCase) UpdateImage(ctx context.Context, id int64, imageKey string) (*model.Banner, error) {
	banner, err := u.bannerRepo.FindByID(ctx, id)
	if err != nil || banner == nil {
		return nil, errors.New("banner not found")
	}
	banner.ImageKey = imageKey
	return banner, u.bannerRepo.Update(ctx, banner)
}

func (u *BannerUseCase) UpdateOrder(ctx context.Context, placement string, ids []int64) error {
	if !bannerPlacements[placement] {
		return errors.New("invalid placement")
	}
	return u.bannerRepo.UpdateOrder(ctx, placement, ids)
}
