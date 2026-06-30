package database

import (
	"context"
	"errors"
	"yutagame-backend/domain/model"

	"gorm.io/gorm"
)

type GenreRepository struct {
	db *gorm.DB
}

func NewGenreRepository(db *gorm.DB) *GenreRepository {
	return &GenreRepository{db: db}
}

func (r *GenreRepository) FindAll(ctx context.Context) ([]model.Genre, error) {
	var genres []model.Genre
	err := r.db.WithContext(ctx).Order("id asc").Find(&genres).Error
	return genres, err
}

func (r *GenreRepository) FindByID(ctx context.Context, id int64) (*model.Genre, error) {
	var genre model.Genre
	err := r.db.WithContext(ctx).First(&genre, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &genre, err
}
