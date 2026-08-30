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
	gameUseCase *usecaseApp.GamePublicUseCase
}

type PaginatedGameResponse struct {
	Data       any   `json:"data"`
	TotalCount int64 `json:"totalCount"`
	TotalPages int   `json:"totalPages"`
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
}

func NewGameHandler(gameUseCase *usecaseApp.GamePublicUseCase) *GameHandler {
	return &GameHandler{gameUseCase: gameUseCase}
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
	releaseLimit := parseLimit(c.QueryParam("releaseLimit"), 10)
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
	}

	ctx := c.Request().Context()
	games, totalCount, totalPages, err := h.gameUseCase.SearchGames(ctx, page, limit, filter)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, handler.ErrorResponse{Message: err.Error()})
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

	return c.JSON(http.StatusOK, game)
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
