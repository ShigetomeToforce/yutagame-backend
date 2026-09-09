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

type PublicAccessEventInput struct {
	EventType         string
	EventSource       string
	Path              string
	Method            string
	StatusCode        int
	VisitorID         string
	IP                string
	UserAgent         string
	Referrer          string
	MachineCode       string
	ManufacturerCode  string
	GenreCode         string
	KeywordCode       string
	SearchWord        string
	GameCode          string
	AffiliateCategory string
}

type AccessLogPublicUseCase struct {
	accessLogRepo *database.AccessLogRepository
}

func NewAccessLogPublicUseCase(accessLogRepo *database.AccessLogRepository) *AccessLogPublicUseCase {
	return &AccessLogPublicUseCase{accessLogRepo: accessLogRepo}
}

func normalizeEventType(eventType string) string {
	switch strings.TrimSpace(eventType) {
	case "page_view", "api_hit", "search", "affiliate_click", "error":
		return strings.TrimSpace(eventType)
	default:
		return "page_view"
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

func (u *AccessLogPublicUseCase) CreateEvent(ctx context.Context, input PublicAccessEventInput) error {
	item := &model.AccessLog{
		EventType:         normalizeEventType(input.EventType),
		EventSource:       strings.TrimSpace(input.EventSource),
		Path:              strings.TrimSpace(input.Path),
		Method:            strings.TrimSpace(input.Method),
		StatusCode:        input.StatusCode,
		VisitorID:         strings.TrimSpace(input.VisitorID),
		IPHash:            hashIP(input.IP),
		UserAgent:         strings.TrimSpace(input.UserAgent),
		Referrer:          strings.TrimSpace(input.Referrer),
		MachineCode:       strings.TrimSpace(input.MachineCode),
		ManufacturerCode:  strings.TrimSpace(input.ManufacturerCode),
		GenreCode:         strings.TrimSpace(input.GenreCode),
		KeywordCode:       strings.TrimSpace(input.KeywordCode),
		SearchWord:        strings.TrimSpace(input.SearchWord),
		GameCode:          strings.TrimSpace(input.GameCode),
		AffiliateCategory: strings.TrimSpace(input.AffiliateCategory),
	}

	if item.Path == "" {
		item.Path = "/"
	}
	if item.Method == "" {
		item.Method = "GET"
	}
	if item.EventSource == "" {
		item.EventSource = "frontend"
	}

	if item.EventType == "page_view" || item.EventType == "search" {
		exists, err := u.accessLogRepo.ExistsRecentDuplicate(ctx, item.VisitorID, item.EventType, item.Path, 5*time.Second)
		if err != nil {
			return err
		}
		if exists {
			return nil
		}
	}

	return u.accessLogRepo.Create(ctx, item)
}
