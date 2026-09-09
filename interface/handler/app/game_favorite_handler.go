package app

import (
	"net/http"
	"strings"
	usecaseApp "yutagame-backend/application/usecase/app"
	"yutagame-backend/interface/handler"

	"github.com/labstack/echo/v4"
)

type GameFavoriteHandler struct {
	favoriteUseCase *usecaseApp.GameFavoriteUseCase
}

func NewGameFavoriteHandler(favoriteUseCase *usecaseApp.GameFavoriteUseCase) *GameFavoriteHandler {
	return &GameFavoriteHandler{favoriteUseCase: favoriteUseCase}
}

type GameFavoriteRequest struct {
	VisitorID string `json:"visitorId"`
}

func (h *GameFavoriteHandler) Push(c echo.Context) error {
	visitorID := strings.TrimSpace(c.Request().Header.Get("X-Visitor-Id"))
	if visitorID == "" {
		var req GameFavoriteRequest
		if err := c.Bind(&req); err == nil {
			visitorID = strings.TrimSpace(req.VisitorID)
		}
	}

	result, err := h.favoriteUseCase.PushToday(c.Request().Context(), c.Param("code"), visitorID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{Message: err.Error()})
	}
	return c.JSON(http.StatusCreated, result)
}

func (h *GameFavoriteHandler) GetStatus(c echo.Context) error {
	visitorID := strings.TrimSpace(c.Request().Header.Get("X-Visitor-Id"))
	if visitorID == "" {
		visitorID = strings.TrimSpace(c.QueryParam("visitorId"))
	}

	result, err := h.favoriteUseCase.GetStatus(c.Request().Context(), c.Param("code"), visitorID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{Message: err.Error()})
	}
	return c.JSON(http.StatusOK, result)
}
