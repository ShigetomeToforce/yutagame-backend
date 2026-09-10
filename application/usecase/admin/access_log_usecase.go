package admin

import (
	"context"
	"strings"
	"time"
	"yutagame-backend/infrastructure/database"
)

type AccessLogKPIBucket struct {
	From            string                      `json:"from"`
	To              string                      `json:"to"`
	PageViews       int64                       `json:"pageViews"`
	UniqueVisitors  int64                       `json:"uniqueVisitors"`
	APICalls        int64                       `json:"apiCalls"`
	SearchCount     int64                       `json:"searchCount"`
	ContactCount    int64                       `json:"contactCount"`
	GenreSearches   int64                       `json:"genreSearches"`
	MachineSearches int64                       `json:"machineSearches"`
	MakerSearches   int64                       `json:"makerSearches"`
	KeywordSearches int64                       `json:"keywordSearches"`
	AffiliateClicks int64                       `json:"affiliateClicks"`
	ErrorCount      int64                       `json:"errorCount"`
	TopMachines     []database.AccessLogTopItem `json:"topMachines"`
	TopManufactures []database.AccessLogTopItem `json:"topManufacturers"`
	TopGenres       []database.AccessLogTopItem `json:"topGenres"`
	TopKeywords     []database.AccessLogTopItem `json:"topKeywords"`
	TopSearchWords  []database.AccessLogTopItem `json:"topSearchWords"`
}

type AccessLogSearchBreakdown struct {
	Field string                      `json:"field"`
	Date  string                      `json:"date"`
	Daily []database.AccessLogTopItem `json:"daily"`
	Total []database.AccessLogTopItem `json:"total"`
}

type AccessLogMachineSearchDashboard struct {
	Period string                      `json:"period"`
	Target string                      `json:"target"`
	Items  []database.AccessLogTopItem `json:"items"`
}

type AccessLogDashboard struct {
	Daily   AccessLogKPIBucket `json:"daily"`
	Monthly AccessLogKPIBucket `json:"monthly"`
}

type AccessLogMonthlyRow struct {
	Date            string `json:"date"`
	PageViews       int64  `json:"pageViews"`
	UniqueVisitors  int64  `json:"uniqueVisitors"`
	SearchCount     int64  `json:"searchCount"`
	MachineSearches int64  `json:"machineSearches"`
	MakerSearches   int64  `json:"makerSearches"`
	GenreSearches   int64  `json:"genreSearches"`
	KeywordSearches int64  `json:"keywordSearches"`
	ContactCount    int64  `json:"contactCount"`
}

type AccessLogMonthlyTable struct {
	Month      string                `json:"month"`
	PrevMonth  string                `json:"prevMonth"`
	NextMonth  string                `json:"nextMonth"`
	Rows       []AccessLogMonthlyRow `json:"rows"`
	MonthlySum AccessLogMonthlyRow   `json:"monthlySum"`
}

type AccessLogUseCase struct {
	searchLogRepo    *database.SearchLogRepository
	gameViewLogRepo  *database.GameViewLogRepository
	contactRepo      *database.ContactInquiryRepository
	machineRepo      *database.MachineRepository
	manufacturerRepo *database.ManufacturerRepository
	genreRepo        *database.GenreRepository
	keywordRepo      *database.KeywordRepository
	gameRepo         *database.GameRepository
}

func NewAccessLogUseCase(
	searchLogRepo *database.SearchLogRepository,
	gameViewLogRepo *database.GameViewLogRepository,
	contactRepo *database.ContactInquiryRepository,
	machineRepo *database.MachineRepository,
	manufacturerRepo *database.ManufacturerRepository,
	genreRepo *database.GenreRepository,
	keywordRepo *database.KeywordRepository,
	gameRepo *database.GameRepository,
) *AccessLogUseCase {
	return &AccessLogUseCase{
		searchLogRepo:    searchLogRepo,
		gameViewLogRepo:  gameViewLogRepo,
		contactRepo:      contactRepo,
		machineRepo:      machineRepo,
		manufacturerRepo: manufacturerRepo,
		genreRepo:        genreRepo,
		keywordRepo:      keywordRepo,
		gameRepo:         gameRepo,
	}
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

func parseMonthStart(value string) (time.Time, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, false
	}
	t, err := time.ParseInLocation("2006-01", value, time.Local)
	if err != nil {
		return time.Time{}, false
	}
	return t, true
}

func (u *AccessLogUseCase) buildBucket(ctx context.Context, from, to time.Time) (AccessLogKPIBucket, error) {
	searchAgg, err := u.searchLogRepo.AggregateRange(ctx, from, to)
	if err != nil {
		return AccessLogKPIBucket{}, err
	}
	viewAgg, err := u.gameViewLogRepo.AggregateRange(ctx, from, to)
	if err != nil {
		return AccessLogKPIBucket{}, err
	}
	contactCount, err := u.contactRepo.CountRange(ctx, from, to)
	if err != nil {
		return AccessLogKPIBucket{}, err
	}

	topMachines, err := u.searchLogRepo.TopByField(ctx, from, to, "machineCode", 5)
	if err != nil {
		return AccessLogKPIBucket{}, err
	}
	topManufacturers, err := u.searchLogRepo.TopByField(ctx, from, to, "manufacturerCode", 5)
	if err != nil {
		return AccessLogKPIBucket{}, err
	}
	topGenres, err := u.searchLogRepo.TopByField(ctx, from, to, "genreCode", 5)
	if err != nil {
		return AccessLogKPIBucket{}, err
	}
	topKeywords, err := u.searchLogRepo.TopByField(ctx, from, to, "keywordCode", 5)
	if err != nil {
		return AccessLogKPIBucket{}, err
	}
	topSearchWords, err := u.searchLogRepo.TopByField(ctx, from, to, "searchWord", 5)
	if err != nil {
		return AccessLogKPIBucket{}, err
	}

	return AccessLogKPIBucket{
		From:            from.Format("2006-01-02"),
		To:              to.Add(-time.Nanosecond).Format("2006-01-02"),
		PageViews:       viewAgg.PageViews,
		UniqueVisitors:  viewAgg.UniqueVisitors,
		APICalls:        0,
		SearchCount:     searchAgg.SearchCount,
		ContactCount:    contactCount,
		GenreSearches:   searchAgg.GenreSearches,
		MachineSearches: searchAgg.MachineSearches,
		MakerSearches:   searchAgg.MakerSearches,
		KeywordSearches: searchAgg.KeywordSearches,
		AffiliateClicks: 0,
		ErrorCount:      0,
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

func normalizeBreakdownField(field string) string {
	switch strings.TrimSpace(field) {
	case "machineCode", "manufacturerCode", "genreCode", "keywordCode", "searchWord":
		return strings.TrimSpace(field)
	default:
		return "genreCode"
	}
}

func resolvePeriodRange(period, date, month string) (time.Time, time.Time, string) {
	period = normalizePeriod(period)
	now := time.Now()

	if period == "total" {
		return time.Time{}, time.Time{}, "累計"
	}

	if period == "monthly" {
		monthStart, ok := parseMonthStart(month)
		if !ok {
			monthStart = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		}
		return monthStart, monthStart.AddDate(0, 1, 0), monthStart.Format("2006-01")
	}

	dayStart, ok := parseDateStart(date)
	if !ok {
		dayStart = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	}
	return dayStart, dayStart.AddDate(0, 0, 1), dayStart.Format("2006-01-02")
}

func normalizePeriod(period string) string {
	switch strings.TrimSpace(period) {
	case "daily", "monthly", "total":
		return strings.TrimSpace(period)
	default:
		return "daily"
	}
}

func (u *AccessLogUseCase) toMachineNameItems(
	ctx context.Context,
	items []database.AccessLogTopItem,
) []database.AccessLogTopItem {
	if len(items) == 0 || u.machineRepo == nil {
		return items
	}

	machines, err := u.machineRepo.FindAll(ctx)
	if err != nil {
		return items
	}

	nameByCode := make(map[string]string, len(machines))
	for _, m := range machines {
		code := strings.TrimSpace(m.Code)
		name := strings.TrimSpace(m.Name)
		if code != "" && name != "" {
			nameByCode[code] = name
		}
	}

	converted := make([]database.AccessLogTopItem, 0, len(items))
	for _, item := range items {
		label := strings.TrimSpace(item.Value)
		if name, ok := nameByCode[label]; ok {
			label = name
		}
		converted = append(converted, database.AccessLogTopItem{
			Value: label,
			Count: item.Count,
		})
	}

	return converted
}

func convertItemsByCode(
	items []database.AccessLogTopItem,
	labelByCode map[string]string,
) []database.AccessLogTopItem {
	if len(items) == 0 || len(labelByCode) == 0 {
		return items
	}

	converted := make([]database.AccessLogTopItem, 0, len(items))
	for _, item := range items {
		label := strings.TrimSpace(item.Value)
		if mapped, ok := labelByCode[label]; ok {
			label = mapped
		}
		converted = append(converted, database.AccessLogTopItem{Value: label, Count: item.Count})
	}
	return converted
}

func (u *AccessLogUseCase) toManufacturerNameItems(
	ctx context.Context,
	items []database.AccessLogTopItem,
) []database.AccessLogTopItem {
	if len(items) == 0 || u.manufacturerRepo == nil {
		return items
	}

	rows, err := u.manufacturerRepo.FindAll(ctx)
	if err != nil {
		return items
	}

	labelByCode := make(map[string]string, len(rows))
	for _, row := range rows {
		code := strings.TrimSpace(row.Code)
		name := strings.TrimSpace(row.Name)
		if code != "" && name != "" {
			labelByCode[code] = name
		}
	}
	return convertItemsByCode(items, labelByCode)
}

func (u *AccessLogUseCase) toGenreNameItems(
	ctx context.Context,
	items []database.AccessLogTopItem,
) []database.AccessLogTopItem {
	if len(items) == 0 || u.genreRepo == nil {
		return items
	}

	rows, err := u.genreRepo.FindAll(ctx)
	if err != nil {
		return items
	}

	labelByCode := make(map[string]string, len(rows))
	for _, row := range rows {
		code := strings.TrimSpace(row.Code)
		name := strings.TrimSpace(row.Name)
		if code != "" && name != "" {
			labelByCode[code] = name
		}
	}
	return convertItemsByCode(items, labelByCode)
}

func (u *AccessLogUseCase) toKeywordNameItems(
	ctx context.Context,
	items []database.AccessLogTopItem,
) []database.AccessLogTopItem {
	if len(items) == 0 || u.keywordRepo == nil {
		return items
	}

	rows, err := u.keywordRepo.FindAll(ctx)
	if err != nil {
		return items
	}

	labelByCode := make(map[string]string, len(rows))
	for _, row := range rows {
		code := strings.TrimSpace(row.Code)
		name := strings.TrimSpace(row.Name)
		if code != "" && name != "" {
			labelByCode[code] = name
		}
	}
	return convertItemsByCode(items, labelByCode)
}

func (u *AccessLogUseCase) toGameNameItems(
	ctx context.Context,
	items []database.AccessLogTopItem,
) []database.AccessLogTopItem {
	if len(items) == 0 || u.gameRepo == nil {
		return items
	}

	rows, err := u.gameRepo.FindAll(ctx)
	if err != nil {
		return items
	}

	labelByCode := make(map[string]string, len(rows))
	for _, row := range rows {
		code := strings.TrimSpace(row.Code)
		name := strings.TrimSpace(row.Name)
		if code != "" && name != "" {
			labelByCode[code] = name
		}
	}
	return convertItemsByCode(items, labelByCode)
}

func (u *AccessLogUseCase) mapRankingItemsByField(
	ctx context.Context,
	field string,
	items []database.AccessLogTopItem,
) []database.AccessLogTopItem {
	switch field {
	case "machineCode":
		return u.toMachineNameItems(ctx, items)
	case "manufacturerCode":
		return u.toManufacturerNameItems(ctx, items)
	case "genreCode":
		return u.toGenreNameItems(ctx, items)
	case "keywordCode":
		return u.toKeywordNameItems(ctx, items)
	default:
		return items
	}
}

func (u *AccessLogUseCase) GetSearchBreakdown(
	ctx context.Context,
	field string,
	date string,
	limit int,
) (*AccessLogSearchBreakdown, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	normalizedField := normalizeBreakdownField(field)

	targetDate, ok := parseDateStart(date)
	if !ok {
		now := time.Now()
		targetDate = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	}

	dailyFrom := targetDate
	dailyTo := targetDate.AddDate(0, 0, 1)

	daily, err := u.searchLogRepo.TopByField(ctx, dailyFrom, dailyTo, normalizedField, limit)
	if err != nil {
		return nil, err
	}
	total, err := u.searchLogRepo.TopByFieldAllTime(ctx, normalizedField, limit)
	if err != nil {
		return nil, err
	}

	return &AccessLogSearchBreakdown{
		Field: normalizedField,
		Date:  dailyFrom.Format("2006-01-02"),
		Daily: daily,
		Total: total,
	}, nil
}

func (u *AccessLogUseCase) GetMonthlyTable(
	ctx context.Context,
	month string,
) (*AccessLogMonthlyTable, error) {
	monthStart, ok := parseMonthStart(month)
	if !ok {
		now := time.Now()
		monthStart = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	}

	monthEnd := monthStart.AddDate(0, 1, 0)

	rows := make([]AccessLogMonthlyRow, 0, 31)
	for day := monthStart; day.Before(monthEnd); day = day.AddDate(0, 0, 1) {
		dayEnd := day.AddDate(0, 0, 1)
		searchAgg, err := u.searchLogRepo.AggregateRange(ctx, day, dayEnd)
		if err != nil {
			return nil, err
		}
		viewAgg, err := u.gameViewLogRepo.AggregateRange(ctx, day, dayEnd)
		if err != nil {
			return nil, err
		}
		contactCount, err := u.contactRepo.CountRange(ctx, day, dayEnd)
		if err != nil {
			return nil, err
		}

		rows = append(rows, AccessLogMonthlyRow{
			Date:            day.Format("2006-01-02"),
			PageViews:       viewAgg.PageViews,
			UniqueVisitors:  viewAgg.UniqueVisitors,
			SearchCount:     searchAgg.SearchCount,
			MachineSearches: searchAgg.MachineSearches,
			MakerSearches:   searchAgg.MakerSearches,
			GenreSearches:   searchAgg.GenreSearches,
			KeywordSearches: searchAgg.KeywordSearches,
			ContactCount:    contactCount,
		})
	}

	totalSearchAgg, err := u.searchLogRepo.AggregateRange(ctx, monthStart, monthEnd)
	if err != nil {
		return nil, err
	}
	totalViewAgg, err := u.gameViewLogRepo.AggregateRangeWithMonthlyUniquePVUU(ctx, monthStart, monthEnd)
	if err != nil {
		return nil, err
	}
	totalContactCount, err := u.contactRepo.CountRange(ctx, monthStart, monthEnd)
	if err != nil {
		return nil, err
	}

	monthlySum := AccessLogMonthlyRow{
		Date:            "合計",
		PageViews:       totalViewAgg.PageViews,
		UniqueVisitors:  totalViewAgg.UniqueVisitors,
		SearchCount:     totalSearchAgg.SearchCount,
		MachineSearches: totalSearchAgg.MachineSearches,
		MakerSearches:   totalSearchAgg.MakerSearches,
		GenreSearches:   totalSearchAgg.GenreSearches,
		KeywordSearches: totalSearchAgg.KeywordSearches,
		ContactCount:    totalContactCount,
	}

	return &AccessLogMonthlyTable{
		Month:      monthStart.Format("2006-01"),
		PrevMonth:  monthStart.AddDate(0, -1, 0).Format("2006-01"),
		NextMonth:  monthStart.AddDate(0, 1, 0).Format("2006-01"),
		Rows:       rows,
		MonthlySum: monthlySum,
	}, nil
}

func (u *AccessLogUseCase) GetMachineSearchDashboard(
	ctx context.Context,
	period string,
	date string,
	month string,
	limit int,
) (*AccessLogMachineSearchDashboard, error) {
	return u.GetSearchRankingDashboard(ctx, "machineCode", period, date, month, limit)
}

func (u *AccessLogUseCase) GetSearchRankingDashboard(
	ctx context.Context,
	field string,
	period string,
	date string,
	month string,
	limit int,
) (*AccessLogMachineSearchDashboard, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 200 {
		limit = 200
	}

	field = normalizeBreakdownField(field)
	period = normalizePeriod(period)
	from, to, target := resolvePeriodRange(period, date, month)

	var (
		items []database.AccessLogTopItem
		err   error
	)
	if period == "total" {
		items, err = u.searchLogRepo.TopByFieldAllTime(ctx, field, limit)
	} else {
		items, err = u.searchLogRepo.TopByField(ctx, from, to, field, limit)
	}
	if err != nil {
		return nil, err
	}
	items = u.mapRankingItemsByField(ctx, field, items)
	if items == nil {
		items = []database.AccessLogTopItem{}
	}

	return &AccessLogMachineSearchDashboard{
		Period: period,
		Target: target,
		Items:  items,
	}, nil
}

func (u *AccessLogUseCase) GetGameViewDashboard(
	ctx context.Context,
	period string,
	date string,
	month string,
	limit int,
) (*AccessLogMachineSearchDashboard, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 200 {
		limit = 200
	}

	period = normalizePeriod(period)
	from, to, target := resolvePeriodRange(period, date, month)

	var (
		items []database.AccessLogTopItem
		err   error
	)
	if period == "total" {
		items, err = u.gameViewLogRepo.TopByGameCodeAllTime(ctx, limit)
	} else {
		items, err = u.gameViewLogRepo.TopByGameCode(ctx, from, to, limit)
	}
	if err != nil {
		return nil, err
	}
	items = u.toGameNameItems(ctx, items)
	if items == nil {
		items = []database.AccessLogTopItem{}
	}

	return &AccessLogMachineSearchDashboard{
		Period: period,
		Target: target,
		Items:  items,
	}, nil
}
