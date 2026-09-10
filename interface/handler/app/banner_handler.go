package app

import (
	"net/http"
	"strconv"
	"time"
	"yutagame-backend/domain/model"
	"yutagame-backend/infrastructure/database"
	"yutagame-backend/interface/handler"

	"github.com/labstack/echo/v4"
)

type BannerHandler struct {
	bannerRepo *database.BannerRepository
}

func NewBannerHandler(bannerRepo *database.BannerRepository) *BannerHandler {
	return &BannerHandler{bannerRepo: bannerRepo}
}

func (h *BannerHandler) GetByPlacement(c echo.Context) error {
	items, err := h.bannerRepo.FindVisibleByPlacement(
		c.Request().Context(),
		c.Param("placement"),
		time.Now(),
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, handler.ErrorResponse{Message: err.Error()})
	}
	if items == nil {
		items = []model.Banner{}
	}
	return c.JSON(http.StatusOK, items)
}

func (h *BannerHandler) CountClick(c echo.Context) error {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := h.bannerRepo.IncrementClickCount(c.Request().Context(), id); err != nil {
		return c.JSON(http.StatusInternalServerError, handler.ErrorResponse{Message: err.Error()})
	}
	return c.NoContent(http.StatusNoContent)
}
