package database

import (
	"context"
	"time"
	"yutagame-backend/domain/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type PageViewAggregate struct {
	PageViews      int64 `json:"pageViews"`
	UniqueVisitors int64 `json:"uniqueVisitors"`
}

type PageViewLogRepository struct {
	db *gorm.DB
}

func NewPageViewLogRepository(db *gorm.DB) *PageViewLogRepository {
	return &PageViewLogRepository{db: db}
}

// CreateDaily はDBの一意制約を使い、同時リクエストでも二重計上を防ぎます。
func (r *PageViewLogRepository) CreateDaily(ctx context.Context, item *model.PageViewLog) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(item).Error
}

func (r *PageViewLogRepository) AggregateRange(ctx context.Context, from, to time.Time) (PageViewAggregate, error) {
	var row PageViewAggregate
	err := r.db.WithContext(ctx).
		Model(&model.PageViewLog{}).
		Select("COUNT(*) AS page_views, COUNT(DISTINCT visitor_hash) AS unique_visitors").
		Where("viewed_on >= ? AND viewed_on < ?", from, to).
		Scan(&row).Error
	return row, err
}

func (r *PageViewLogRepository) AggregateDailyRange(ctx context.Context, from, to time.Time) (map[string]PageViewAggregate, error) {
	var rows []struct {
		Date string
		PageViewAggregate
	}
	err := r.db.WithContext(ctx).
		Model(&model.PageViewLog{}).
		Select("DATE_FORMAT(viewed_on, '%Y-%m-%d') AS date, COUNT(*) AS page_views, COUNT(DISTINCT visitor_hash) AS unique_visitors").
		Where("viewed_on >= ? AND viewed_on < ?", from, to).
		Group("viewed_on").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	result := make(map[string]PageViewAggregate, len(rows))
	for _, row := range rows {
		result[row.Date] = row.PageViewAggregate
	}
	return result, nil
}
