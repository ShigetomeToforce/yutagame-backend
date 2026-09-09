package admin

import (
	"net/http"
	"strconv"
	usecaseAdmin "yutagame-backend/application/usecase/admin"
	"yutagame-backend/interface/handler"

	"github.com/labstack/echo/v4"
)

type AnnouncementSaveRequest struct {
	Title    string `json:"title"`
	Excerpt  string `json:"excerpt"`
	BodyHTML string `json:"bodyHtml"`
	Status   string `json:"status"`
}

type AnnouncementHandler struct {
	announcementUseCase *usecaseAdmin.AnnouncementUseCase
}

func NewAnnouncementHandler(announcementUseCase *usecaseAdmin.AnnouncementUseCase) *AnnouncementHandler {
	return &AnnouncementHandler{announcementUseCase: announcementUseCase}
}

func (h *AnnouncementHandler) Create(c echo.Context) error {
	var req AnnouncementSaveRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{Message: err.Error()})
	}
	ctx := c.Request().Context()
	announcement, err := h.announcementUseCase.CreateAnnouncement(ctx, req.Title, req.Excerpt, req.BodyHTML, req.Status)
	if err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{Message: err.Error()})
	}
	return c.JSON(http.StatusCreated, announcement)
}

func (h *AnnouncementHandler) GetByID(c echo.Context) error {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	ctx := c.Request().Context()
	announcement, err := h.announcementUseCase.GetAnnouncementByID(ctx, id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{Message: err.Error()})
	}
	if announcement == nil {
		return c.JSON(http.StatusNotFound, handler.ErrorResponse{Message: "指定されたお知らせが見つかりませんでした。"})
	}
	return c.JSON(http.StatusOK, announcement)
}

func (h *AnnouncementHandler) GetAll(c echo.Context) error {
	return handler.HandleListOrPagination(
		c,
		h.announcementUseCase.GetAllAnnouncements,
		h.announcementUseCase.GetAnnouncementsWithPagination,
		func(c echo.Context) usecaseAdmin.AnnouncementListFilter {
			return usecaseAdmin.AnnouncementListFilter{
				SearchWord: c.QueryParam("q"),
				Status:     c.QueryParam("status"),
			}
		},
	)
}

func (h *AnnouncementHandler) Update(c echo.Context) error {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req AnnouncementSaveRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{Message: err.Error()})
	}
	ctx := c.Request().Context()
	announcement, err := h.announcementUseCase.UpdateAnnouncement(ctx, id, req.Title, req.Excerpt, req.BodyHTML, req.Status)
	if err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{Message: err.Error()})
	}
	return c.JSON(http.StatusOK, announcement)
}

func (h *AnnouncementHandler) Delete(c echo.Context) error {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	ctx := c.Request().Context()
	if err := h.announcementUseCase.DeleteAnnouncement(ctx, id); err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{Message: err.Error()})
	}
	return c.NoContent(http.StatusNoContent)
}
