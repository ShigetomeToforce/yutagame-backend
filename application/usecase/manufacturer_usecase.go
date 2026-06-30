package usecase

import (
	"context"
	"yutagame-backend/domain/model"
	"yutagame-backend/infrastructure/database"
)

type ManufacturerUseCase struct {
	manufacturerRepo *database.ManufacturerRepository
}

func NewManufacturerUseCase(manufacturerRepo *database.ManufacturerRepository) *ManufacturerUseCase {
	return &ManufacturerUseCase{manufacturerRepo: manufacturerRepo}
}

func (u *ManufacturerUseCase) GetAllManufacturers(ctx context.Context) ([]model.Manufacturer, error) {
	return u.manufacturerRepo.FindAll(ctx)
}
