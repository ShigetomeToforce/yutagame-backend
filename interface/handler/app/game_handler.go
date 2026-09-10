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

func NewGameHandler(
	gameUseCase *usecaseApp.GamePublicUseCase,
	analyticsLogUseCase *usecaseApp.AnalyticsLogUseCase,
) *GameHandler {
	return &GameHandler{gameUseCase: gameUseCase, analyticsLogUseCase: analyticsLogUseCase}
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

func (h *GameHandler) GetMachines(c echo.Context) error {
	ctx := c.Request().Context()
	items, err := h.gameUseCase.GetMachines(ctx)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, handler.ErrorResponse{Message: err.Error()})
	}
	return c.JSON(http.StatusOK, items)
}

func (h *GameHandler) GetGenres(c echo.Context) error {
	ctx := c.Request().Context()
	items, err := h.gameUseCase.GetGenres(ctx)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, handler.ErrorResponse{Message: err.Error()})
	}
	return c.JSON(http.StatusOK, items)
}

func (h *GameHandler) GetManufacturers(c echo.Context) error {
	ctx := c.Request().Context()
	items, err := h.gameUseCase.GetManufacturers(ctx)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, handler.ErrorResponse{Message: err.Error()})
	}
	return c.JSON(http.StatusOK, items)
}

func (h *GameHandler) GetKeywords(c echo.Context) error {
	ctx := c.Request().Context()
	items, err := h.gameUseCase.GetKeywords(ctx)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, handler.ErrorResponse{Message: err.Error()})
	}
	return c.JSON(http.StatusOK, items)
}

func (h *GameHandler) GetTop(c echo.Context) error {
	releaseLimit := parseLimit(c.QueryParam("releaseLimit"), 5)
	recentLimit := parseLimit(c.QueryParam("recentLimit"), 8)
	randomLimit := parseLimit(c.QueryParam("randomLimit"), 8)

	ctx := c.Request().Context()
	topContents, err := h.gameUseCase.GetTopContents(ctx, releaseLimit, recentLimit, randomLimit)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, handler.ErrorResponse{Message: err.Error()})
	}

	return c.JSON(http.StatusOK, topContents)
}

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
		return c.JSON(http.StatusInternalServerError, handler.ErrorResponse{Message: err.Error()})
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

func (h *GameHandler) GetByCode(c echo.Context) error {
	code := strings.TrimSpace(c.Param("code"))
	if code == "" {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{Message: "code を指定してください。"})
	}

	ctx := c.Request().Context()
	game, err := h.gameUseCase.GetGameByCode(ctx, code)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, handler.ErrorResponse{Message: err.Error()})
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

func (h *GameHandler) GetRanking(c echo.Context) error {
	ctx := c.Request().Context()
	items, err := h.gameUseCase.GetActiveRanking(ctx)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, handler.ErrorResponse{Message: err.Error()})
	}
	return c.JSON(http.StatusOK, items)
}

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
		return c.JSON(http.StatusInternalServerError, handler.ErrorResponse{Message: err.Error()})
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
