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

// Push godoc
// @Summary 本日の「推し」を登録
// @Description 同じ訪問者・ゲームの組み合わせは日本時間で1日1回だけ登録します。
// @Tags Public Favorites
// @Accept json
// @Produce json
// @Param code path string true "ゲームコード"
// @Param X-Visitor-Id header string false "匿名訪問者ID"
// @Param request body GameFavoriteRequest false "匿名訪問者ID"
// @Success 201 {object} usecaseApp.GameFavoriteResult
// @Failure 400 {object} handler.ErrorResponse
// @Router /app/games/{code}/favorite [post]
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

// GetStatus godoc
// @Summary 本日の「推し」状態取得
// @Tags Public Favorites
// @Produce json
// @Param code path string true "ゲームコード"
// @Param visitorId query string false "匿名訪問者ID"
// @Param X-Visitor-Id header string false "匿名訪問者ID"
// @Success 200 {object} usecaseApp.GameFavoriteResult
// @Failure 400 {object} handler.ErrorResponse
// @Router /app/games/{code}/favorite [get]
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
