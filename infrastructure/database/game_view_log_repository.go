package database

import (
	"context"
	"strings"
	"time"
	"yutagame-backend/domain/model"

	"gorm.io/gorm"
)

type GameViewAggregate struct {
	PageViews      int64 `json:"pageViews"`
	UniqueVisitors int64 `json:"uniqueVisitors"`
}

type GameViewLogRepository struct {
	db *gorm.DB
}

func NewGameViewLogRepository(db *gorm.DB) *GameViewLogRepository {
	return &GameViewLogRepository{db: db}
}

func (r *GameViewLogRepository) Create(ctx context.Context, item *model.GameViewLog) error {
	return r.db.WithContext(ctx).Create(item).Error
}

func (r *GameViewLogRepository) ExistsDailyGameView(
	ctx context.Context,
	visitorID, ipHash, gameCode string,
	now time.Time,
) (bool, error) {
	visitorID = strings.TrimSpace(visitorID)
	ipHash = strings.TrimSpace(ipHash)
	gameCode = strings.TrimSpace(gameCode)
	if gameCode == "" || (visitorID == "" && ipHash == "") {
		return false, nil
	}

	dayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	dayEnd := dayStart.AddDate(0, 0, 1)

	query := r.db.WithContext(ctx).
		Model(&model.GameViewLog{}).
		Where("game_code = ? AND created_at >= ? AND created_at < ?", gameCode, dayStart, dayEnd)

	if visitorID != "" {
		query = query.Where("visitor_id = ?", visitorID)
	} else {
		query = query.Where("visitor_id = '' AND ip_hash = ?", ipHash)
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *GameViewLogRepository) AggregateRange(ctx context.Context, from, to time.Time) (GameViewAggregate, error) {
	var row GameViewAggregate
	err := r.db.WithContext(ctx).
		Model(&model.GameViewLog{}).
		Select(`
			COUNT(*) AS page_views,
			COUNT(DISTINCT NULLIF(COALESCE(NULLIF(visitor_id, ''), CONCAT('ip:', NULLIF(ip_hash, ''))), '')) AS unique_visitors
		`).
		Where("created_at >= ? AND created_at < ?", from, to).
		Scan(&row).Error
	return row, err
}

func (r *GameViewLogRepository) AggregateRangeWithMonthlyUniquePVUU(
	ctx context.Context,
	from, to time.Time,
) (GameViewAggregate, error) {
	var row GameViewAggregate
	err := r.db.WithContext(ctx).
		Model(&model.GameViewLog{}).
		Select(`
			COUNT(DISTINCT CONCAT(COALESCE(NULLIF(visitor_id, ''), CONCAT('ip:', NULLIF(ip_hash, ''))), '|', game_code)) AS page_views,
			COUNT(DISTINCT CASE
				WHEN COALESCE(NULLIF(visitor_id, ''), NULLIF(ip_hash, '')) IS NOT NULL THEN COALESCE(NULLIF(visitor_id, ''), CONCAT('ip:', NULLIF(ip_hash, '')))
				ELSE NULL
			END) AS unique_visitors
		`).
		Where("created_at >= ? AND created_at < ?", from, to).
		Scan(&row).Error
	return row, err
}

func (r *GameViewLogRepository) TopByGameCode(
	ctx context.Context,
	from, to time.Time,
	limit int,
) ([]AccessLogTopItem, error) {
	var rows []AccessLogTopItem
	err := r.db.WithContext(ctx).
		Model(&model.GameViewLog{}).
		Select("game_code AS value, COUNT(*) AS count").
		Where("created_at >= ? AND created_at < ?", from, to).
		Where("game_code <> ''").
		Group("game_code").
		Order("count DESC").
		Limit(limit).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *GameViewLogRepository) TopByGameCodeAllTime(
	ctx context.Context,
	limit int,
) ([]AccessLogTopItem, error) {
	var rows []AccessLogTopItem
	err := r.db.WithContext(ctx).
		Model(&model.GameViewLog{}).
		Select("game_code AS value, COUNT(*) AS count").
		Where("game_code <> ''").
		Group("game_code").
		Order("count DESC").
		Limit(limit).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	return rows, nil
}
