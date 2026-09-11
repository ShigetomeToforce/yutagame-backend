package app

import (
	"net/http"
	"strconv"
	"time"
	usecaseApp "yutagame-backend/application/usecase/app"
	"yutagame-backend/domain/model"
	"yutagame-backend/infrastructure/database"
	"yutagame-backend/interface/handler"

	"github.com/labstack/echo/v4"
)

type BannerHandler struct {
	bannerRepo          *database.BannerRepository
	analyticsLogUseCase *usecaseApp.AnalyticsLogUseCase
}

func NewBannerHandler(bannerRepo *database.BannerRepository, analyticsLogUseCase *usecaseApp.AnalyticsLogUseCase) *BannerHandler {
	return &BannerHandler{bannerRepo: bannerRepo, analyticsLogUseCase: analyticsLogUseCase}
}

func (h *BannerHandler) GetByPlacement(c echo.Context) error {
	items, err := h.bannerRepo.FindVisibleByPlacement(
		c.Request().Context(),
		c.Param("placement"),
		time.Now(),
	)
	if err != nil {
		return handler.RespondError(c, http.StatusInternalServerError, err)
	}
	if items == nil {
		items = []model.Banner{}
	}
	if h.analyticsLogUseCase != nil {
		for _, item := range items {
			if logErr := h.analyticsLogUseCase.RecordContentAccess(c.Request().Context(), usecaseApp.ContentAccessLogInput{
				VisitorID:   resolveVisitorID(c),
				IP:          c.RealIP(),
				ContentType: "banner",
				ContentKey:  strconv.FormatInt(item.ID, 10),
			}); logErr != nil {
				c.Logger().Warnf("banner access log write failed: %v", logErr)
			}
		}
	}
	return c.JSON(http.StatusOK, items)
}

func (h *BannerHandler) CountClick(c echo.Context) error {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := h.bannerRepo.IncrementClickCount(c.Request().Context(), id); err != nil {
		return handler.RespondError(c, http.StatusInternalServerError, err)
	}
	return c.NoContent(http.StatusNoContent)
}
