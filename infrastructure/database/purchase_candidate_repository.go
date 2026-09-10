package database

import (
	"context"
	"errors"
	"yutagame-backend/domain/model"

	"gorm.io/gorm"
)

type PurchaseCandidateRepository struct {
	db *gorm.DB
}

func NewPurchaseCandidateRepository(db *gorm.DB) *PurchaseCandidateRepository {
	return &PurchaseCandidateRepository{db: db}
}

func preloadPurchaseCandidateRelations(db *gorm.DB) *gorm.DB {
	return db.Preload("Manufacturer").Preload("Machine").Preload("Genre")
}

func (r *PurchaseCandidateRepository) Create(ctx context.Context, item *model.PurchaseCandidate) error {
	return r.db.WithContext(ctx).Create(item).Error
}

func (r *PurchaseCandidateRepository) FindByID(ctx context.Context, id int64) (*model.PurchaseCandidate, error) {
	var item model.PurchaseCandidate
	err := preloadPurchaseCandidateRelations(r.db.WithContext(ctx)).First(&item, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &item, err
}

func (r *PurchaseCandidateRepository) FindByCode(ctx context.Context, code string) (*model.PurchaseCandidate, error) {
	var item model.PurchaseCandidate
	err := preloadPurchaseCandidateRelations(r.db.WithContext(ctx)).Where("code = ?", code).First(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &item, err
}

func (r *PurchaseCandidateRepository) FindAllByPurchased(ctx context.Context, isPurchased bool) ([]model.PurchaseCandidate, error) {
	var items []model.PurchaseCandidate
	err := preloadPurchaseCandidateRelations(r.db.WithContext(ctx)).
		Where("is_purchased = ?", isPurchased).
		Order("display_order asc, id asc").
		Find(&items).Error
	return items, err
}

func (r *PurchaseCandidateRepository) Update(ctx context.Context, item *model.PurchaseCandidate) error {
	return r.db.WithContext(ctx).Model(&model.PurchaseCandidate{}).Where("id = ?", item.ID).Updates(map[string]interface{}{
		"name":              item.Name,
		"kana":              item.Kana,
		"code":              item.Code,
		"list_price":        item.ListPrice,
		"official_site_url": item.OfficialSiteURL,
		"youtube_url":       item.YouTubeURL,
		"release_date_text": item.ReleaseDateText,
		"manufacturer_id":   item.ManufacturerID,
		"machine_id":        item.MachineID,
		"genre_id":          item.GenreID,
		"is_purchased":      item.IsPurchased,
	}).Error
}

func (r *PurchaseCandidateRepository) UpdateImage(ctx context.Context, id int64, imageKey *string) (*model.PurchaseCandidate, error) {
	if err := r.db.WithContext(ctx).Model(&model.PurchaseCandidate{}).Where("id = ?", id).Update("image_key", imageKey).Error; err != nil {
		return nil, err
	}
	return r.FindByID(ctx, id)
}

func (r *PurchaseCandidateRepository) UpdateOrder(ctx context.Context, isPurchased bool, ids []int64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for index, id := range ids {
			if err := tx.Model(&model.PurchaseCandidate{}).Where("id = ? AND is_purchased = ?", id, isPurchased).Update("display_order", index+1).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *PurchaseCandidateRepository) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&model.PurchaseCandidate{}, id).Error
}
