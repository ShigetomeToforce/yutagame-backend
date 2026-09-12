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
	ctx := c.Request().Context()
	draft, err := h.useCase.GetDraft(ctx)
	if err != nil {
		return handler.RespondError(c, http.StatusInternalServerError, err)
	}
	active, err := h.useCase.GetActive(ctx)
	if err != nil {
		return handler.RespondError(c, http.StatusInternalServerError, err)
	}
	return c.JSON(http.StatusOK, map[string]any{
		"draft":  draft,
		"active": active,
	})
}

func (h *GameRankingHandler) GetDraft(c echo.Context) error {
	items, err := h.useCase.GetDraft(c.Request().Context())
	if err != nil {
		return handler.RespondError(c, http.StatusInternalServerError, err)
	}
	return c.JSON(http.StatusOK, items)
}

func (h *GameRankingHandler) GetActive(c echo.Context) error {
	items, err := h.useCase.GetActive(c.Request().Context())
	if err != nil {
		return handler.RespondError(c, http.StatusInternalServerError, err)
	}
	return c.JSON(http.StatusOK, items)
}

func (h *GameRankingHandler) GetEditOrder(c echo.Context) error {
	items, err := h.useCase.GetDefaultOrder(c.Request().Context())
	if err != nil {
		return handler.RespondError(c, http.StatusInternalServerError, err)
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
		return handler.RespondError(c, http.StatusInternalServerError, err)
	}
	return c.JSON(http.StatusOK, items)
}

func (h *GameRankingHandler) DiscardDraft(c echo.Context) error {
	if err := h.useCase.DiscardDraft(c.Request().Context()); err != nil {
		return handler.RespondError(c, http.StatusInternalServerError, err)
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *GameRankingHandler) Publish(c echo.Context) error {
	items, err := h.useCase.PublishDraft(c.Request().Context())
	if err != nil {
		return handler.RespondError(c, http.StatusInternalServerError, err)
	}
	return c.JSON(http.StatusOK, items)
}
