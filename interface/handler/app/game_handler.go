package app

import (
	"net/http"
	"strconv"
	"strings"
	usecaseApp "yutagame-backend/application/usecase/app"
	"yutagame-backend/interface/handler"

	"github.com/labstack/echo/v4"
)

type GameHandler struct {
	gameUseCase         *usecaseApp.GamePublicUseCase
	analyticsLogUseCase *usecaseApp.AnalyticsLogUseCase
}

type PaginatedGameResponse struct {
	Data       any   `json:"data"`
	TotalCount int64 `json:"totalCount"`
	TotalPages int   `json:"totalPages"`
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
}

type PageViewRequest struct {
	VisitorID string `json:"visitorId"`
	PagePath  string `json:"pagePath"`
}

func NewGameHandler(
	gameUseCase *usecaseApp.GamePublicUseCase,
	analyticsLogUseCase *usecaseApp.AnalyticsLogUseCase,
) *GameHandler {
	return &GameHandler{gameUseCase: gameUseCase, analyticsLogUseCase: analyticsLogUseCase}
}

// RecordPageView godoc
// @Summary ページ閲覧を日次記録
// @Description 同じ訪問者による同じページの閲覧は、日本時間で1日1回だけPVとして記録します。
// @Tags Analytics
// @Accept json
// @Produce json
// @Param request body PageViewRequest true "ページ閲覧情報"
// @Success 204
// @Failure 400 {object} handler.ErrorResponse
// @Failure 500 {object} handler.ErrorResponse
// @Router /app/page-views [post]
func (h *GameHandler) RecordPageView(c echo.Context) error {
	var request PageViewRequest
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{Message: "リクエスト形式が正しくありません。"})
	}
	if h.analyticsLogUseCase == nil {
		return c.NoContent(http.StatusNoContent)
	}
	if err := h.analyticsLogUseCase.RecordPageView(c.Request().Context(), usecaseApp.PageViewLogInput{
		VisitorID: request.VisitorID,
		IP:        c.RealIP(),
		PagePath:  request.PagePath,
	}); err != nil {
		return handler.RespondError(c, http.StatusInternalServerError, err)
	}
	return c.NoContent(http.StatusNoContent)
}

func getCookieValue(rawCookie, key string) string {
	parts := strings.Split(rawCookie, ";")
	for _, part := range parts {
		pair := strings.SplitN(strings.TrimSpace(part), "=", 2)
		if len(pair) != 2 {
			continue
		}
		if pair[0] == key {
			return pair[1]
		}
	}
	return ""
}

func resolveVisitorID(c echo.Context) string {
	visitorID := strings.TrimSpace(c.QueryParam("visitorId"))
	if visitorID != "" {
		return visitorID
	}
	return strings.TrimSpace(getCookieValue(c.Request().Header.Get("Cookie"), "visitor_id"))
}

// GetMachines godoc
// @Summary 機種カタログ取得
// @Tags Public Games
// @Produce json
// @Success 200 {array} usecaseApp.CatalogItem
// @Router /app/catalog/machines [get]
func (h *GameHandler) GetMachines(c echo.Context) error {
	ctx := c.Request().Context()
	items, err := h.gameUseCase.GetMachines(ctx)
	if err != nil {
		return handler.RespondError(c, http.StatusInternalServerError, err)
	}
	return c.JSON(http.StatusOK, items)
}

// GetGenres godoc
// @Summary ジャンルカタログ取得
// @Tags Public Games
// @Produce json
// @Success 200 {array} usecaseApp.CatalogItem
// @Router /app/catalog/genres [get]
func (h *GameHandler) GetGenres(c echo.Context) error {
	ctx := c.Request().Context()
	items, err := h.gameUseCase.GetGenres(ctx)
	if err != nil {
		return handler.RespondError(c, http.StatusInternalServerError, err)
	}
	return c.JSON(http.StatusOK, items)
}

// GetManufacturers godoc
// @Summary メーカーカタログ取得
// @Tags Public Games
// @Produce json
// @Success 200 {array} usecaseApp.CatalogItem
// @Router /app/catalog/manufacturers [get]
func (h *GameHandler) GetManufacturers(c echo.Context) error {
	ctx := c.Request().Context()
	items, err := h.gameUseCase.GetManufacturers(ctx)
	if err != nil {
		return handler.RespondError(c, http.StatusInternalServerError, err)
	}
	return c.JSON(http.StatusOK, items)
}

// GetKeywords godoc
// @Summary キーワード一覧取得
// @Tags Public Games
// @Produce json
// @Success 200 {array} usecaseApp.KeywordItem
// @Router /app/keywords [get]
func (h *GameHandler) GetKeywords(c echo.Context) error {
	ctx := c.Request().Context()
	items, err := h.gameUseCase.GetKeywords(ctx)
	if err != nil {
		return handler.RespondError(c, http.StatusInternalServerError, err)
	}
	return c.JSON(http.StatusOK, items)
}

// GetSiteStats godoc
// @Summary 公開サイト統計取得
// @Tags Public Games
// @Produce json
// @Success 200 {object} usecaseApp.SiteStats
// @Router /app/site-stats [get]
func (h *GameHandler) GetSiteStats(c echo.Context) error {
	stats, err := h.gameUseCase.GetSiteStats(c.Request().Context())
	if err != nil {
		return handler.RespondError(c, http.StatusInternalServerError, err)
	}
	return c.JSON(http.StatusOK, stats)
}

// GetTop godoc
// @Summary TOP画面コンテンツ取得
// @Tags Public Games
// @Produce json
// @Param releaseLimit query int false "発売日枠の件数" default(5)
// @Param recentLimit query int false "新着枠の件数" default(8)
// @Param randomLimit query int false "ランダム・お気に入り枠の件数" default(8)
// @Success 200 {object} usecaseApp.TopContents
// @Router /app/top [get]
func (h *GameHandler) GetTop(c echo.Context) error {
	releaseLimit := parseLimit(c.QueryParam("releaseLimit"), 5)
	recentLimit := parseLimit(c.QueryParam("recentLimit"), 8)
	randomLimit := parseLimit(c.QueryParam("randomLimit"), 8)

	ctx := c.Request().Context()
	topContents, err := h.gameUseCase.GetTopContents(ctx, releaseLimit, recentLimit, randomLimit)
	if err != nil {
		return handler.RespondError(c, http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusOK, topContents)
}

// Search godoc
// @Summary ゲーム検索
// @Tags Public Games
// @Produce json
// @Param q query string false "ゲーム名・カナ"
// @Param machineCode query string false "機種コード"
// @Param genreCode query string false "ジャンルコード"
// @Param manufacturerCode query string false "メーカーコード"
// @Param keywordCode query string false "キーワードコード"
// @Param sort query string false "並び順"
// @Param page query int false "ページ番号" default(1)
// @Param limit query int false "表示件数" default(20)
// @Param visitorId query string false "匿名訪問者ID"
// @Success 200 {object} PaginatedGameResponse
// @Router /app/games [get]
func (h *GameHandler) Search(c echo.Context) error {
	page := parsePositiveInt(c.QueryParam("page"), 1)
	limit := parseLimit(c.QueryParam("limit"), 20)

	filter := usecaseApp.GameListFilter{
		SearchWord:       c.QueryParam("q"),
		MachineCode:      c.QueryParam("machineCode"),
		GenreCode:        c.QueryParam("genreCode"),
		ManufacturerCode: c.QueryParam("manufacturerCode"),
		KeywordCode:      c.QueryParam("keywordCode"),
		Sort:             c.QueryParam("sort"),
	}

	ctx := c.Request().Context()
	games, totalCount, totalPages, err := h.gameUseCase.SearchGames(ctx, page, limit, filter)
	if err != nil {
		return handler.RespondError(c, http.StatusInternalServerError, err)
	}

	if h.analyticsLogUseCase != nil {
		if logErr := h.analyticsLogUseCase.RecordSearch(c.Request().Context(), usecaseApp.SearchLogInput{
			VisitorID:        resolveVisitorID(c),
			IP:               c.RealIP(),
			MachineCode:      filter.MachineCode,
			ManufacturerCode: filter.ManufacturerCode,
			GenreCode:        filter.GenreCode,
			KeywordCode:      filter.KeywordCode,
			SearchWord:       filter.SearchWord,
		}); logErr != nil {
			c.Logger().Warnf("search log write failed: %v", logErr)
		}
	}

	return c.JSON(http.StatusOK, PaginatedGameResponse{
		Data:       games,
		TotalCount: totalCount,
		TotalPages: totalPages,
		Page:       page,
		Limit:      limit,
	})
}

// GetByCode godoc
// @Summary ゲーム詳細取得
// @Tags Public Games
// @Produce json
// @Param code path string true "ゲームコード"
// @Param visitorId query string false "匿名訪問者ID"
// @Success 200 {object} model.Game
// @Failure 404 {object} handler.ErrorResponse
// @Router /app/games/{code} [get]
func (h *GameHandler) GetByCode(c echo.Context) error {
	code := strings.TrimSpace(c.Param("code"))
	if code == "" {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{Message: "code を指定してください。"})
	}

	ctx := c.Request().Context()
	game, err := h.gameUseCase.GetGameByCode(ctx, code)
	if err != nil {
		return handler.RespondError(c, http.StatusInternalServerError, err)
	}
	if game == nil {
		return c.JSON(http.StatusNotFound, handler.ErrorResponse{Message: "指定されたゲームが見つかりませんでした。"})
	}

	if h.analyticsLogUseCase != nil {
		if logErr := h.analyticsLogUseCase.RecordGameView(c.Request().Context(), usecaseApp.GameViewLogInput{
			VisitorID: resolveVisitorID(c),
			IP:        c.RealIP(),
			GameCode:  game.Code,
		}); logErr != nil {
			c.Logger().Warnf("game view log write failed: %v", logErr)
		}
	}

	return c.JSON(http.StatusOK, game)
}

// GetRanking godoc
// @Summary 公開中ランキング取得
// @Tags Public Rankings
// @Produce json
// @Success 200 {array} model.Game
// @Router /app/rankings [get]
func (h *GameHandler) GetRanking(c echo.Context) error {
	ctx := c.Request().Context()
	items, err := h.gameUseCase.GetActiveRanking(ctx)
	if err != nil {
		return handler.RespondError(c, http.StatusInternalServerError, err)
	}
	return c.JSON(http.StatusOK, items)
}

// GetRankingPage godoc
// @Summary 集計ランキング取得
// @Tags Public Rankings
// @Produce json
// @Param type query string false "ランキング種別"
// @Param period query string false "集計期間"
// @Param page query int false "ページ番号" default(1)
// @Param limit query int false "表示件数" default(20)
// @Success 200 {object} usecaseApp.PublicRankingResponse
// @Router /app/rankings/page [get]
func (h *GameHandler) GetRankingPage(c echo.Context) error {
	page := parsePositiveInt(c.QueryParam("page"), 1)
	limit := parseLimit(c.QueryParam("limit"), 20)
	result, err := h.gameUseCase.GetPublicRankingPage(
		c.Request().Context(),
		strings.TrimSpace(c.QueryParam("type")),
		strings.TrimSpace(c.QueryParam("period")),
		page,
		limit,
	)
	if err != nil {
		return handler.RespondError(c, http.StatusInternalServerError, err)
	}
	return c.JSON(http.StatusOK, result)
}

func parsePositiveInt(raw string, fallback int) int {
	v, err := strconv.Atoi(raw)
	if err != nil || v < 1 {
		return fallback
	}
	return v
}

func parseLimit(raw string, fallback int) int {
	v := parsePositiveInt(raw, fallback)
	if v > 100 {
		return 100
	}
	return v
}
