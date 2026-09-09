package database

import (
	"context"
	"fmt"
	"strings"
	"time"
	"yutagame-backend/domain/model"

	"gorm.io/gorm"
)

type AccessLogTopItem struct {
	Value string `json:"value"`
	Count int64  `json:"count"`
}

type AccessLogAggregate struct {
	PageViews       int64 `json:"pageViews"`
	UniqueVisitors  int64 `json:"uniqueVisitors"`
	APICalls        int64 `json:"apiCalls"`
	SearchCount     int64 `json:"searchCount"`
	AffiliateClicks int64 `json:"affiliateClicks"`
	ErrorCount      int64 `json:"errorCount"`
}

type AccessLogRepository struct {
	db *gorm.DB
}

func NewAccessLogRepository(db *gorm.DB) *AccessLogRepository {
	return &AccessLogRepository{db: db}
}

func (r *AccessLogRepository) Create(ctx context.Context, item *model.AccessLog) error {
	return r.db.WithContext(ctx).Create(item).Error
}

func (r *AccessLogRepository) ExistsRecentDuplicate(
	ctx context.Context,
	visitorID, eventType, path string,
	within time.Duration,
) (bool, error) {
	if strings.TrimSpace(visitorID) == "" || strings.TrimSpace(eventType) == "" || strings.TrimSpace(path) == "" {
		return false, nil
	}

	threshold := time.Now().Add(-within)
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.AccessLog{}).
		Where("visitor_id = ? AND event_type = ? AND path = ? AND created_at >= ?", visitorID, eventType, path, threshold).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *AccessLogRepository) FindAll(ctx context.Context) ([]model.AccessLog, error) {
	var rows []model.AccessLog
	err := r.db.WithContext(ctx).
		Order("created_at desc, id desc").
		Limit(200).
		Find(&rows).Error
	return rows, err
}

func (r *AccessLogRepository) FindAllWithPagination(
	ctx context.Context,
	limit, offset int,
	whereQueries ...func(*gorm.DB) *gorm.DB,
) ([]model.AccessLog, error) {
	return ExecuteFindWithPagination[model.AccessLog](ctx, r.db, limit, offset, "created_at desc, id desc", nil, whereQueries...)
}

func (r *AccessLogRepository) CountAll(ctx context.Context, whereQueries ...func(*gorm.DB) *gorm.DB) (int64, error) {
	return ExecuteCount[model.AccessLog](ctx, r.db, whereQueries...)
}

func (r *AccessLogRepository) AggregateRange(ctx context.Context, from, to time.Time) (AccessLogAggregate, error) {
	var row AccessLogAggregate
	err := r.db.WithContext(ctx).
		Model(&model.AccessLog{}).
		Select(`
			SUM(CASE WHEN event_type = 'page_view' THEN 1 ELSE 0 END) AS page_views,
			COUNT(DISTINCT NULLIF(visitor_id, '')) AS unique_visitors,
			SUM(CASE WHEN event_type = 'api_hit' THEN 1 ELSE 0 END) AS api_calls,
			SUM(CASE WHEN event_type = 'search' THEN 1 ELSE 0 END) AS search_count,
			SUM(CASE WHEN event_type = 'affiliate_click' THEN 1 ELSE 0 END) AS affiliate_clicks,
			SUM(CASE WHEN event_type = 'error' OR status_code >= 500 THEN 1 ELSE 0 END) AS error_count
		`).
		Where("created_at >= ? AND created_at < ?", from, to).
		Scan(&row).Error
	return row, err
}

func (r *AccessLogRepository) TopByField(
	ctx context.Context,
	from, to time.Time,
	field string,
	limit int,
) ([]AccessLogTopItem, error) {
	column, ok := map[string]string{
		"machineCode":      "machine_code",
		"manufacturerCode": "manufacturer_code",
		"genreCode":        "genre_code",
		"keywordCode":      "keyword_code",
		"searchWord":       "search_word",
		"path":             "path",
	}[field]
	if !ok {
		return nil, fmt.Errorf("unsupported top field: %s", field)
	}

	var rows []AccessLogTopItem
	err := r.db.WithContext(ctx).
		Model(&model.AccessLog{}).
		Select(fmt.Sprintf("%s AS value, COUNT(*) AS count", column)).
		Where("created_at >= ? AND created_at < ?", from, to).
		Where(fmt.Sprintf("%s <> ''", column)).
		Group(column).
		Order("count DESC").
		Limit(limit).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	return rows, nil
}
