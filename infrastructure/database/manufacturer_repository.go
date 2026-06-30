package database

import (
	"context"
	"errors"
	"yutagame-backend/domain/model"

	"gorm.io/gorm"
)

type ManufacturerRepository struct {
	db *gorm.DB
}

func NewManufacturerRepository(db *gorm.DB) *ManufacturerRepository {
	return &ManufacturerRepository{db: db}
}

func (r *ManufacturerRepository) FindAll(ctx context.Context) ([]model.Manufacturer, error) {
	var manufacturers []model.Manufacturer
	err := r.db.WithContext(ctx).Order("id asc").Find(&manufacturers).Error
	return manufacturers, err
}

func (r *ManufacturerRepository) FindByID(ctx context.Context, id int64) (*model.Manufacturer, error) {
	var manufacturer model.Manufacturer
	err := r.db.WithContext(ctx).First(&manufacturer, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &manufacturer, err
}
