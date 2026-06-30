package usecase

import (
	"context"
	"yutagame-backend/domain/model"
	"yutagame-backend/infrastructure/database"
)

type GenreUseCase struct {
	genreRepo *database.GenreRepository
}

func NewGenreUseCase(genreRepo *database.GenreRepository) *GenreUseCase {
	return &GenreUseCase{genreRepo: genreRepo}
}

func (u *GenreUseCase) GetAllGenres(ctx context.Context) ([]model.Genre, error) {
	return u.genreRepo.FindAll(ctx)
}
