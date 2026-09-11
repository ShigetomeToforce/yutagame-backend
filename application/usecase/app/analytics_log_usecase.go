package app

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"time"
	"yutagame-backend/application/usecase"
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

type ContentAccessLogInput struct {
	VisitorID   string
	IP          string
	ContentType string
	ContentKey  string
}

type PageViewLogInput struct {
	VisitorID string
	IP        string
	PagePath  string
}

type AnalyticsLogUseCase struct {
	searchLogRepo        *database.SearchLogRepository
	gameViewLogRepo      *database.GameViewLogRepository
	contentAccessLogRepo *database.ContentAccessLogRepository
	pageViewLogRepo      *database.PageViewLogRepository
}

func NewAnalyticsLogUseCase(
	searchLogRepo *database.SearchLogRepository,
	gameViewLogRepo *database.GameViewLogRepository,
	contentAccessLogRepo *database.ContentAccessLogRepository,
	pageViewLogRepo *database.PageViewLogRepository,
) *AnalyticsLogUseCase {
	return &AnalyticsLogUseCase{
		searchLogRepo:        searchLogRepo,
		gameViewLogRepo:      gameViewLogRepo,
		contentAccessLogRepo: contentAccessLogRepo,
		pageViewLogRepo:      pageViewLogRepo,
	}
}

func visitorHash(visitorID, ip string) string {
	identity := strings.TrimSpace(visitorID)
	if identity == "" {
		identity = "ip:" + strings.TrimSpace(ip)
	}
	if identity == "" || identity == "ip:" {
		return ""
	}
	h := sha256.Sum256([]byte(identity))
	return hex.EncodeToString(h[:])
}

func normalizePagePath(pagePath string) string {
	pagePath = strings.TrimSpace(pagePath)
	if pagePath == "" || !strings.HasPrefix(pagePath, "/") {
		return ""
	}
	runes := []rune(pagePath)
	if len(runes) > 191 {
		return string(runes[:191])
	}
	return pagePath
}

func (u *AnalyticsLogUseCase) RecordPageView(ctx context.Context, input PageViewLogInput) error {
	if u.pageViewLogRepo == nil {
		return nil
	}
	identity := visitorHash(input.VisitorID, input.IP)
	pagePath := normalizePagePath(input.PagePath)
	if identity == "" || pagePath == "" {
		return nil
	}

	// サーバーのOS設定に依存せず、日本の日付境界で日次集計します。
	now := time.Now().In(usecase.JapanLocation)
	return u.pageViewLogRepo.CreateDaily(ctx, &model.PageViewLog{
		VisitorHash: identity,
		PagePath:    pagePath,
		ViewedOn:    now.Format("2006-01-02"),
	})
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
		time.Now().In(usecase.JapanLocation),
	)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	return u.gameViewLogRepo.Create(ctx, item)
}

func (u *AnalyticsLogUseCase) RecordContentAccess(ctx context.Context, input ContentAccessLogInput) error {
	if u.contentAccessLogRepo == nil {
		return nil
	}
	item := &model.ContentAccessLog{
		VisitorID:   strings.TrimSpace(input.VisitorID),
		IPHash:      hashIP(input.IP),
		ContentType: strings.TrimSpace(input.ContentType),
		ContentKey:  strings.TrimSpace(input.ContentKey),
	}
	if item.ContentType == "" || item.ContentKey == "" {
		return nil
	}
	exists, err := u.contentAccessLogRepo.ExistsDailyAccess(ctx, item.ContentType, item.ContentKey, item.VisitorID, item.IPHash, time.Now().In(usecase.JapanLocation))
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	return u.contentAccessLogRepo.Create(ctx, item)
}
