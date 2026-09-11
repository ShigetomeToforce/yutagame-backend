package database

import (
	"context"
	"fmt"
	"time"
	"yutagame-backend/domain/model"

	"gorm.io/gorm"
)

type AccessLogTopItem struct {
	Value string `json:"value"`
	Count int64  `json:"count"`
}

type SearchLogAggregate struct {
	SearchCount     int64 `json:"searchCount"`
	GenreSearches   int64 `json:"genreSearches"`
	MachineSearches int64 `json:"machineSearches"`
	MakerSearches   int64 `json:"makerSearches"`
	KeywordSearches int64 `json:"keywordSearches"`
}

type SearchLogRepository struct {
	db *gorm.DB
}

func NewSearchLogRepository(db *gorm.DB) *SearchLogRepository {
	return &SearchLogRepository{db: db}
}

func (r *SearchLogRepository) Create(ctx context.Context, item *model.SearchLog) error {
	return r.db.WithContext(ctx).Create(item).Error
}

func (r *SearchLogRepository) AggregateRange(ctx context.Context, from, to time.Time) (SearchLogAggregate, error) {
	var row SearchLogAggregate
	err := r.db.WithContext(ctx).
		Model(&model.SearchLog{}).
		Select(`
			COUNT(*) AS search_count,
			SUM(CASE WHEN genre_code <> '' THEN 1 ELSE 0 END) AS genre_searches,
			SUM(CASE WHEN machine_code <> '' THEN 1 ELSE 0 END) AS machine_searches,
			SUM(CASE WHEN manufacturer_code <> '' THEN 1 ELSE 0 END) AS maker_searches,
			SUM(CASE WHEN keyword_code <> '' THEN 1 ELSE 0 END) AS keyword_searches
		`).
		Where("created_at >= ? AND created_at < ?", from, to).
		Scan(&row).Error
	return row, err
}

func (r *SearchLogRepository) AggregateDailyRange(ctx context.Context, from, to time.Time) (map[string]SearchLogAggregate, error) {
	var rows []struct {
		Date string
		SearchLogAggregate
	}
	err := r.db.WithContext(ctx).
		Model(&model.SearchLog{}).
		Select(`
			DATE_FORMAT(created_at, '%Y-%m-%d') AS date,
			COUNT(*) AS search_count,
			SUM(CASE WHEN genre_code <> '' THEN 1 ELSE 0 END) AS genre_searches,
			SUM(CASE WHEN machine_code <> '' THEN 1 ELSE 0 END) AS machine_searches,
			SUM(CASE WHEN manufacturer_code <> '' THEN 1 ELSE 0 END) AS maker_searches,
			SUM(CASE WHEN keyword_code <> '' THEN 1 ELSE 0 END) AS keyword_searches
		`).
		Where("created_at >= ? AND created_at < ?", from, to).
		Group("DATE(created_at)").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	result := make(map[string]SearchLogAggregate, len(rows))
	for _, row := range rows {
		result[row.Date] = row.SearchLogAggregate
	}
	return result, nil
}

func searchTopColumn(field string) (string, bool) {
	column, ok := map[string]string{
		"machineCode":      "machine_code",
		"manufacturerCode": "manufacturer_code",
		"genreCode":        "genre_code",
		"keywordCode":      "keyword_code",
		"searchWord":       "search_word",
	}[field]
	return column, ok
}

func (r *SearchLogRepository) TopByField(
	ctx context.Context,
	from, to time.Time,
	field string,
	limit int,
) ([]AccessLogTopItem, error) {
	column, ok := searchTopColumn(field)
	if !ok {
		return nil, fmt.Errorf("unsupported top field: %s", field)
	}

	var rows []AccessLogTopItem
	err := r.db.WithContext(ctx).
		Model(&model.SearchLog{}).
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

func (r *SearchLogRepository) TopByFieldAllTime(
	ctx context.Context,
	field string,
	limit int,
) ([]AccessLogTopItem, error) {
	column, ok := searchTopColumn(field)
	if !ok {
		return nil, fmt.Errorf("unsupported top field: %s", field)
	}

	var rows []AccessLogTopItem
	err := r.db.WithContext(ctx).
		Model(&model.SearchLog{}).
		Select(fmt.Sprintf("%s AS value, COUNT(*) AS count", column)).
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
