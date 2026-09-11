package database

import (
	"context"
	"strings"
	"time"
	"yutagame-backend/domain/model"

	"gorm.io/gorm"
)

type ContentAccessCounts struct {
	AnnouncementViews int64 `json:"announcementViews"`
	FeatureViews      int64 `json:"featureViews"`
	BannerViews       int64 `json:"bannerViews"`
}

type ContentAccessLogRepository struct {
	db *gorm.DB
}

func NewContentAccessLogRepository(db *gorm.DB) *ContentAccessLogRepository {
	return &ContentAccessLogRepository{db: db}
}

func (r *ContentAccessLogRepository) Create(ctx context.Context, item *model.ContentAccessLog) error {
	return r.db.WithContext(ctx).Create(item).Error
}

func (r *ContentAccessLogRepository) ExistsDailyAccess(ctx context.Context, contentType, contentKey, visitorID, ipHash string, now time.Time) (bool, error) {
	contentType = strings.TrimSpace(contentType)
	contentKey = strings.TrimSpace(contentKey)
	visitorID = strings.TrimSpace(visitorID)
	ipHash = strings.TrimSpace(ipHash)
	if contentType == "" || contentKey == "" || (visitorID == "" && ipHash == "") {
		return false, nil
	}

	dayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	dayEnd := dayStart.AddDate(0, 0, 1)
	query := r.db.WithContext(ctx).
		Model(&model.ContentAccessLog{}).
		Where("content_type = ? AND content_key = ? AND created_at >= ? AND created_at < ?", contentType, contentKey, dayStart, dayEnd)
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

func (r *ContentAccessLogRepository) CountByContentKeys(ctx context.Context, contentType string, keys []string) (map[string]int64, error) {
	result := make(map[string]int64, len(keys))
	if len(keys) == 0 {
		return result, nil
	}
	var rows []struct {
		ContentKey string
		Count      int64
	}
	err := r.db.WithContext(ctx).
		Model(&model.ContentAccessLog{}).
		Select("content_key, COUNT(*) AS count").
		Where("content_type = ? AND content_key IN ?", contentType, keys).
		Group("content_key").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		result[row.ContentKey] = row.Count
	}
	return result, nil
}

func (r *ContentAccessLogRepository) AggregateRange(ctx context.Context, from, to time.Time) (ContentAccessCounts, error) {
	var rows []struct {
		ContentType string
		Count       int64
	}
	err := r.db.WithContext(ctx).
		Model(&model.ContentAccessLog{}).
		Select("content_type, COUNT(*) AS count").
		Where("created_at >= ? AND created_at < ?", from, to).
		Group("content_type").
		Scan(&rows).Error
	if err != nil {
		return ContentAccessCounts{}, err
	}
	return toContentAccessCounts(rows), nil
}

func (r *ContentAccessLogRepository) AggregateDailyRange(ctx context.Context, from, to time.Time) (map[string]ContentAccessCounts, error) {
	var rows []struct {
		Date string
		ContentAccessCounts
	}
	err := r.db.WithContext(ctx).
		Model(&model.ContentAccessLog{}).
		Select(`
			DATE_FORMAT(created_at, '%Y-%m-%d') AS date,
			SUM(CASE WHEN content_type = 'announcement' THEN 1 ELSE 0 END) AS announcement_views,
			SUM(CASE WHEN content_type = 'feature' THEN 1 ELSE 0 END) AS feature_views,
			SUM(CASE WHEN content_type = 'banner' THEN 1 ELSE 0 END) AS banner_views
		`).
		Where("created_at >= ? AND created_at < ?", from, to).
		Group("DATE_FORMAT(created_at, '%Y-%m-%d')").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	result := make(map[string]ContentAccessCounts, len(rows))
	for _, row := range rows {
		result[row.Date] = row.ContentAccessCounts
	}
	return result, nil
}

func (r *ContentAccessLogRepository) AggregateAllTime(ctx context.Context) (ContentAccessCounts, error) {
	var rows []struct {
		ContentType string
		Count       int64
	}
	err := r.db.WithContext(ctx).
		Model(&model.ContentAccessLog{}).
		Select("content_type, COUNT(*) AS count").
		Group("content_type").
		Scan(&rows).Error
	if err != nil {
		return ContentAccessCounts{}, err
	}
	return toContentAccessCounts(rows), nil
}

func toContentAccessCounts(rows []struct {
	ContentType string
	Count       int64
}) ContentAccessCounts {
	var counts ContentAccessCounts
	for _, row := range rows {
		switch row.ContentType {
		case "announcement":
			counts.AnnouncementViews = row.Count
		case "feature":
			counts.FeatureViews = row.Count
		case "banner":
			counts.BannerViews = row.Count
		}
	}
	return counts
}
