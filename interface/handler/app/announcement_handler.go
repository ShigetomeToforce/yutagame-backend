package app

import (
	"net/http"
	"strconv"
	usecaseApp "yutagame-backend/application/usecase/app"
	"yutagame-backend/interface/handler"

	"github.com/labstack/echo/v4"
)

type AnnouncementHandler struct {
	announcementUseCase *usecaseApp.AnnouncementPublicUseCase
	analyticsLogUseCase *usecaseApp.AnalyticsLogUseCase
}

func NewAnnouncementHandler(announcementUseCase *usecaseApp.AnnouncementPublicUseCase, analyticsLogUseCase *usecaseApp.AnalyticsLogUseCase) *AnnouncementHandler {
	return &AnnouncementHandler{announcementUseCase: announcementUseCase, analyticsLogUseCase: analyticsLogUseCase}
}

func (h *AnnouncementHandler) GetAll(c echo.Context) error {
	ctx := c.Request().Context()
	items, err := h.announcementUseCase.GetPublishedAnnouncements(ctx)
	if err != nil {
		return handler.RespondError(c, http.StatusInternalServerError, err)
	}
	return c.JSON(http.StatusOK, items)
}

func (h *AnnouncementHandler) GetByID(c echo.Context) error {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	ctx := c.Request().Context()
	item, err := h.announcementUseCase.GetPublishedAnnouncementByID(ctx, id)
	if err != nil {
		return handler.RespondError(c, http.StatusInternalServerError, err)
	}
	if item == nil {
		return c.JSON(http.StatusNotFound, handler.ErrorResponse{Message: "指定されたお知らせが見つかりませんでした。"})
	}
	if h.analyticsLogUseCase != nil {
		if logErr := h.analyticsLogUseCase.RecordContentAccess(c.Request().Context(), usecaseApp.ContentAccessLogInput{
			VisitorID:   resolveVisitorID(c),
			IP:          c.RealIP(),
			ContentType: "announcement",
			ContentKey:  strconv.FormatInt(item.ID, 10),
		}); logErr != nil {
			c.Logger().Warnf("announcement access log write failed: %v", logErr)
		}
	}
	return c.JSON(http.StatusOK, item)
}
