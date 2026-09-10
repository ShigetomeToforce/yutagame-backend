package database

import (
	"context"
	"errors"
	"time"
	"yutagame-backend/domain/model"

	"gorm.io/gorm"
)

type BannerRepository struct {
	db *gorm.DB
}

func NewBannerRepository(db *gorm.DB) *BannerRepository {
	return &BannerRepository{db: db}
}

func (r *BannerRepository) Create(ctx context.Context, banner *model.Banner) error {
	return r.db.WithContext(ctx).Create(banner).Error
}

func (r *BannerRepository) FindByID(ctx context.Context, id int64) (*model.Banner, error) {
	var banner model.Banner
	err := r.db.WithContext(ctx).First(&banner, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &banner, err
}

func (r *BannerRepository) FindAll(ctx context.Context) ([]model.Banner, error) {
	var banners []model.Banner
	err := r.db.WithContext(ctx).Order("placement asc, display_order asc, id asc").Find(&banners).Error
	return banners, err
}

func (r *BannerRepository) FindVisibleByPlacement(ctx context.Context, placement string, now time.Time) ([]model.Banner, error) {
	var banners []model.Banner
	err := r.db.WithContext(ctx).
		Where("placement = ? AND (starts_at IS NULL OR starts_at <= ?) AND (ends_at IS NULL OR ends_at >= ?)", placement, now, now).
		Order("display_order asc, id asc").Find(&banners).Error
	return banners, err
}

func (r *BannerRepository) Update(ctx context.Context, banner *model.Banner) error {
	return r.db.WithContext(ctx).Save(banner).Error
}

func (r *BannerRepository) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&model.Banner{}, id).Error
}

func (r *BannerRepository) UpdateOrder(ctx context.Context, placement string, ids []int64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for index, id := range ids {
			if err := tx.Model(&model.Banner{}).Where("id = ? AND placement = ?", id, placement).
				Update("display_order", index+1).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *BannerRepository) IncrementClickCount(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Model(&model.Banner{}).Where("id = ?", id).
		UpdateColumn("click_count", gorm.Expr("click_count + ?", 1)).Error
}
