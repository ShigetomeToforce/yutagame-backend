package app

import (
	"context"
	"yutagame-backend/infrastructure/database"
)

type SitemapData struct {
	GameCodes       []string `json:"gameCodes"`
	AnnouncementIDs []int64  `json:"announcementIds"`
	FeatureCodes    []string `json:"featureCodes"`
}

type SiteUseCase struct {
	gameRepo         *database.GameRepository
	announcementRepo *database.AnnouncementRepository
	featureRepo      *database.FeatureRepository
}

func NewSiteUseCase(gameRepo *database.GameRepository, announcementRepo *database.AnnouncementRepository, featureRepo *database.FeatureRepository) *SiteUseCase {
	return &SiteUseCase{gameRepo: gameRepo, announcementRepo: announcementRepo, featureRepo: featureRepo}
}

func (u *SiteUseCase) GetSitemapData(ctx context.Context) (*SitemapData, error) {
	gameCodes, err := u.gameRepo.FindAllCodes(ctx)
	if err != nil {
		return nil, err
	}
	announcementIDs, err := u.announcementRepo.FindPublishedIDs(ctx)
	if err != nil {
		return nil, err
	}
	featureCodes, err := u.featureRepo.FindPublishedCodes(ctx)
	if err != nil {
		return nil, err
	}
	return &SitemapData{GameCodes: gameCodes, AnnouncementIDs: announcementIDs, FeatureCodes: featureCodes}, nil
}
