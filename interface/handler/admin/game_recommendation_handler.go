package admin

import (
	"github.com/labstack/echo/v4"
	"net/http"
	"strconv"
	usecaseAdmin "yutagame-backend/application/usecase/admin"
	"yutagame-backend/interface/handler"
)

type GameRecommendationSaveRequest struct {
	Status    string `json:"status"`
	AdminNote string `json:"adminNote"`
}
type GameRecommendationHandler struct {
	useCase *usecaseAdmin.GameRecommendationUseCase
}

func NewGameRecommendationHandler(useCase *usecaseAdmin.GameRecommendationUseCase) *GameRecommendationHandler {
	return &GameRecommendationHandler{useCase: useCase}
}
func (h *GameRecommendationHandler) GetAll(c echo.Context) error {
	return handler.HandleListOrPagination(c, h.useCase.GetAll, h.useCase.GetWithPagination, func(c echo.Context) usecaseAdmin.GameRecommendationListFilter {
		return usecaseAdmin.GameRecommendationListFilter{SearchWord: c.QueryParam("q"), Status: c.QueryParam("status")}
	})
}
func (h *GameRecommendationHandler) GetByID(c echo.Context) error {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	item, err := h.useCase.GetByID(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{Message: err.Error()})
	}
	if item == nil {
		return c.JSON(http.StatusNotFound, handler.ErrorResponse{Message: "指定されたおすすめゲームが見つかりませんでした。"})
	}
	return c.JSON(http.StatusOK, item)
}
func (h *GameRecommendationHandler) Update(c echo.Context) error {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req GameRecommendationSaveRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{Message: err.Error()})
	}
	item, err := h.useCase.Update(c.Request().Context(), id, req.Status, req.AdminNote)
	if err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{Message: err.Error()})
	}
	return c.JSON(http.StatusOK, item)
}
func (h *GameRecommendationHandler) Delete(c echo.Context) error {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := h.useCase.Delete(c.Request().Context(), id); err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{Message: err.Error()})
	}
	return c.NoContent(http.StatusNoContent)
}
