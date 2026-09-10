package app

import (
	"context"
	"yutagame-backend/domain/model"
	"yutagame-backend/infrastructure/database"
)

type FeaturePublicUseCase struct {
	featureRepo *database.FeatureRepository
}

func NewFeaturePublicUseCase(featureRepo *database.FeatureRepository) *FeaturePublicUseCase {
	return &FeaturePublicUseCase{featureRepo: featureRepo}
}

func (u *FeaturePublicUseCase) GetPublishedFeatures(ctx context.Context) ([]model.Feature, error) {
	return u.featureRepo.FindPublishedAll(ctx)
}

func (u *FeaturePublicUseCase) GetPublishedFeatureByCode(ctx context.Context, code string) (*model.Feature, error) {
	return u.featureRepo.FindPublishedByCode(ctx, code)
}
