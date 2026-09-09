package admin

import (
	"context"
	"strings"
	"time"
	"yutagame-backend/application/usecase"
	"yutagame-backend/domain/model"
	"yutagame-backend/infrastructure/database"

	"gorm.io/gorm"
)

type AccessLogListFilter struct {
	SearchWord string
	EventType  string
	FromDate   string
	ToDate     string
}

type AccessLogKPIBucket struct {
	From            string                      `json:"from"`
	To              string                      `json:"to"`
	PageViews       int64                       `json:"pageViews"`
	UniqueVisitors  int64                       `json:"uniqueVisitors"`
	APICalls        int64                       `json:"apiCalls"`
	SearchCount     int64                       `json:"searchCount"`
	AffiliateClicks int64                       `json:"affiliateClicks"`
	ErrorCount      int64                       `json:"errorCount"`
	TopMachines     []database.AccessLogTopItem `json:"topMachines"`
	TopManufactures []database.AccessLogTopItem `json:"topManufacturers"`
	TopGenres       []database.AccessLogTopItem `json:"topGenres"`
	TopKeywords     []database.AccessLogTopItem `json:"topKeywords"`
	TopSearchWords  []database.AccessLogTopItem `json:"topSearchWords"`
}

type AccessLogDashboard struct {
	Daily   AccessLogKPIBucket `json:"daily"`
	Monthly AccessLogKPIBucket `json:"monthly"`
}

type AccessLogUseCase struct {
	accessLogRepo *database.AccessLogRepository
}

func NewAccessLogUseCase(accessLogRepo *database.AccessLogRepository) *AccessLogUseCase {
	return &AccessLogUseCase{accessLogRepo: accessLogRepo}
}

func parseDateStart(value string) (time.Time, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, false
	}
	t, err := time.ParseInLocation("2006-01-02", value, time.Local)
	if err != nil {
		return time.Time{}, false
	}
	return t, true
}

func (u *AccessLogUseCase) GetAllAccessLogs(ctx context.Context) ([]model.AccessLog, error) {
	return u.accessLogRepo.FindAll(ctx)
}

func (u *AccessLogUseCase) GetAccessLogsWithPagination(
	ctx context.Context,
	page, limit int,
	filter AccessLogListFilter,
) ([]model.AccessLog, int64, int, error) {
	searchWord := strings.TrimSpace(filter.SearchWord)
	eventType := strings.TrimSpace(filter.EventType)
	from, hasFrom := parseDateStart(filter.FromDate)
	to, hasTo := parseDateStart(filter.ToDate)

	var whereQuery func(*gorm.DB) *gorm.DB
	if searchWord != "" || eventType != "" || hasFrom || hasTo {
		whereQuery = func(db *gorm.DB) *gorm.DB {
			if searchWord != "" {
				like := "%" + searchWord + "%"
				db = db.Where(
					"path LIKE ? OR search_word LIKE ? OR visitor_id LIKE ? OR game_code LIKE ?",
					like,
					like,
					like,
					like,
				)
			}
			if eventType != "" {
				db = db.Where("event_type = ?", eventType)
			}
			if hasFrom {
				db = db.Where("created_at >= ?", from)
			}
			if hasTo {
				db = db.Where("created_at < ?", to.AddDate(0, 0, 1))
			}
			return db
		}
	}

	return usecase.ExecutePaginatedSearch(ctx, page, limit, whereQuery, u.accessLogRepo.CountAll, u.accessLogRepo.FindAllWithPagination)
}

func (u *AccessLogUseCase) buildBucket(ctx context.Context, from, to time.Time) (AccessLogKPIBucket, error) {
	agg, err := u.accessLogRepo.AggregateRange(ctx, from, to)
	if err != nil {
		return AccessLogKPIBucket{}, err
	}

	topMachines, err := u.accessLogRepo.TopByField(ctx, from, to, "machineCode", 5)
	if err != nil {
		return AccessLogKPIBucket{}, err
	}
	topManufacturers, err := u.accessLogRepo.TopByField(ctx, from, to, "manufacturerCode", 5)
	if err != nil {
		return AccessLogKPIBucket{}, err
	}
	topGenres, err := u.accessLogRepo.TopByField(ctx, from, to, "genreCode", 5)
	if err != nil {
		return AccessLogKPIBucket{}, err
	}
	topKeywords, err := u.accessLogRepo.TopByField(ctx, from, to, "keywordCode", 5)
	if err != nil {
		return AccessLogKPIBucket{}, err
	}
	topSearchWords, err := u.accessLogRepo.TopByField(ctx, from, to, "searchWord", 5)
	if err != nil {
		return AccessLogKPIBucket{}, err
	}

	return AccessLogKPIBucket{
		From:            from.Format("2006-01-02"),
		To:              to.Add(-time.Nanosecond).Format("2006-01-02"),
		PageViews:       agg.PageViews,
		UniqueVisitors:  agg.UniqueVisitors,
		APICalls:        agg.APICalls,
		SearchCount:     agg.SearchCount,
		AffiliateClicks: agg.AffiliateClicks,
		ErrorCount:      agg.ErrorCount,
		TopMachines:     topMachines,
		TopManufactures: topManufacturers,
		TopGenres:       topGenres,
		TopKeywords:     topKeywords,
		TopSearchWords:  topSearchWords,
	}, nil
}

func (u *AccessLogUseCase) GetDashboard(ctx context.Context) (*AccessLogDashboard, error) {
	now := time.Now()
	dailyStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	dailyEnd := dailyStart.AddDate(0, 0, 1)

	monthlyStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	monthlyEnd := monthlyStart.AddDate(0, 1, 0)

	daily, err := u.buildBucket(ctx, dailyStart, dailyEnd)
	if err != nil {
		return nil, err
	}
	monthly, err := u.buildBucket(ctx, monthlyStart, monthlyEnd)
	if err != nil {
		return nil, err
	}

	return &AccessLogDashboard{Daily: daily, Monthly: monthly}, nil
}
