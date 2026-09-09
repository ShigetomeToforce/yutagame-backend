package app

import (
	"net/http"
	usecaseApp "yutagame-backend/application/usecase/app"
	"yutagame-backend/interface/handler"

	"github.com/labstack/echo/v4"
)

type SiteHandler struct {
	siteUseCase *usecaseApp.SiteUseCase
}

func NewSiteHandler(siteUseCase *usecaseApp.SiteUseCase) *SiteHandler {
	return &SiteHandler{siteUseCase: siteUseCase}
}

func (h *SiteHandler) GetSitemap(c echo.Context) error {
	ctx := c.Request().Context()
	data, err := h.siteUseCase.GetSitemapData(ctx)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, handler.ErrorResponse{Message: err.Error()})
	}
	return c.JSON(http.StatusOK, data)
}
