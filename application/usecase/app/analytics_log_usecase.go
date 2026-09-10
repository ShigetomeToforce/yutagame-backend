package app

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"time"
	"yutagame-backend/domain/model"
	"yutagame-backend/infrastructure/database"
)

type SearchLogInput struct {
	VisitorID        string
	IP               string
	MachineCode      string
	ManufacturerCode string
	GenreCode        string
	KeywordCode      string
	SearchWord       string
}

type GameViewLogInput struct {
	VisitorID string
	IP        string
	GameCode  string
}

type AnalyticsLogUseCase struct {
	searchLogRepo   *database.SearchLogRepository
	gameViewLogRepo *database.GameViewLogRepository
}

func NewAnalyticsLogUseCase(
	searchLogRepo *database.SearchLogRepository,
	gameViewLogRepo *database.GameViewLogRepository,
) *AnalyticsLogUseCase {
	return &AnalyticsLogUseCase{
		searchLogRepo:   searchLogRepo,
		gameViewLogRepo: gameViewLogRepo,
	}
}

func hashIP(ip string) string {
	trimmed := strings.TrimSpace(ip)
	if trimmed == "" {
		return ""
	}
	h := sha256.Sum256([]byte(trimmed))
	return hex.EncodeToString(h[:])
}

func (u *AnalyticsLogUseCase) RecordSearch(ctx context.Context, input SearchLogInput) error {
	item := &model.SearchLog{
		VisitorID:        strings.TrimSpace(input.VisitorID),
		IPHash:           hashIP(input.IP),
		MachineCode:      strings.TrimSpace(input.MachineCode),
		ManufacturerCode: strings.TrimSpace(input.ManufacturerCode),
		GenreCode:        strings.TrimSpace(input.GenreCode),
		KeywordCode:      strings.TrimSpace(input.KeywordCode),
		SearchWord:       strings.TrimSpace(input.SearchWord),
	}

	return u.searchLogRepo.Create(ctx, item)
}

func (u *AnalyticsLogUseCase) RecordGameView(ctx context.Context, input GameViewLogInput) error {
	item := &model.GameViewLog{
		VisitorID: strings.TrimSpace(input.VisitorID),
		IPHash:    hashIP(input.IP),
		GameCode:  strings.TrimSpace(input.GameCode),
	}

	if item.GameCode == "" {
		return nil
	}

	exists, err := u.gameViewLogRepo.ExistsDailyGameView(
		ctx,
		item.VisitorID,
		item.IPHash,
		item.GameCode,
		time.Now(),
	)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	return u.gameViewLogRepo.Create(ctx, item)
}
