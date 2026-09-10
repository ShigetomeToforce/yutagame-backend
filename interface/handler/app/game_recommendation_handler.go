package app

import (
	"log"
	"net/http"
	usecaseAdmin "yutagame-backend/application/usecase/admin"
	"yutagame-backend/infrastructure/mail"
	"yutagame-backend/interface/handler"

	"github.com/labstack/echo/v4"
)

type GameRecommendationRequest struct {
	GameName string `json:"gameName"`
	Reason   string `json:"reason"`
}
type GameRecommendationHandler struct {
	useCase  *usecaseAdmin.GameRecommendationUseCase
	notifier mail.GameRecommendationNotifier
}

func NewGameRecommendationHandler(useCase *usecaseAdmin.GameRecommendationUseCase, notifier mail.GameRecommendationNotifier) *GameRecommendationHandler {
	return &GameRecommendationHandler{useCase: useCase, notifier: notifier}
}
func (h *GameRecommendationHandler) Create(c echo.Context) error {
	var req GameRecommendationRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{Message: err.Error()})
	}
	item, err := h.useCase.Create(c.Request().Context(), req.GameName, req.Reason)
	if err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{Message: err.Error()})
	}
	if h.notifier != nil {
		if notifyErr := h.notifier.SendGameRecommendationNotification(c.Request().Context(), item); notifyErr != nil {
			log.Printf("game recommendation notification mail failed: %v", notifyErr)
		}
	}
	return c.JSON(http.StatusCreated, item)
}
