package admin

import (
	"net/http"
	"yutagame-backend/application/usecase/admin"
	"yutagame-backend/interface/handler"

	"github.com/labstack/echo/v4"
)

type GameRankingHandler struct {
	useCase *admin.GameRankingUseCase
}

type SaveDraftRequest struct {
	GameIDs []int64 `json:"gameIds"`
}

func NewGameRankingHandler(useCase *admin.GameRankingUseCase) *GameRankingHandler {
	return &GameRankingHandler{useCase: useCase}
}

func (h *GameRankingHandler) GetCurrent(c echo.Context) error {
	items, mode, err := h.useCase.GetCurrent(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, handler.ErrorResponse{Message: err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]any{
		"currentMode": mode,
		"current":     items,
		"draft":       nil,
		"active":      nil,
	})
}

func (h *GameRankingHandler) GetDraft(c echo.Context) error {
	items, err := h.useCase.GetDraft(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, handler.ErrorResponse{Message: err.Error()})
	}
	return c.JSON(http.StatusOK, items)
}

func (h *GameRankingHandler) GetActive(c echo.Context) error {
	items, err := h.useCase.GetActive(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, handler.ErrorResponse{Message: err.Error()})
	}
	return c.JSON(http.StatusOK, items)
}

func (h *GameRankingHandler) SaveDraft(c echo.Context) error {
	var req SaveDraftRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{Message: err.Error()})
	}
	items, err := h.useCase.SaveDraft(c.Request().Context(), req.GameIDs)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, handler.ErrorResponse{Message: err.Error()})
	}
	return c.JSON(http.StatusOK, items)
}

func (h *GameRankingHandler) Publish(c echo.Context) error {
	items, err := h.useCase.PublishDraft(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, handler.ErrorResponse{Message: err.Error()})
	}
	return c.JSON(http.StatusOK, items)
}
