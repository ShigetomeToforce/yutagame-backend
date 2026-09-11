package admin

import (
	"net/http"
	"strconv"
	"time"
	usecaseAdmin "yutagame-backend/application/usecase/admin"
	"yutagame-backend/domain/model"
	"yutagame-backend/interface/handler"

	"github.com/labstack/echo/v4"
)

type BannerSaveRequest struct {
	Title        string     `json:"title"`
	Placement    string     `json:"placement"`
	LinkURL      string     `json:"linkUrl"`
	OpenInNewTab bool       `json:"openInNewTab"`
	StartsAt     *time.Time `json:"startsAt"`
	EndsAt       *time.Time `json:"endsAt"`
}

type BannerOrderRequest struct {
	Placement string  `json:"placement"`
	IDs       []int64 `json:"ids"`
}
type BannerHandler struct{ useCase *usecaseAdmin.BannerUseCase }

func NewBannerHandler(useCase *usecaseAdmin.BannerUseCase) *BannerHandler {
	return &BannerHandler{useCase: useCase}
}

func (h *BannerHandler) Create(c echo.Context) error {
	var req BannerSaveRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{Message: err.Error()})
	}
	banner, err := h.useCase.Create(c.Request().Context(), &model.Banner{Title: req.Title, Placement: req.Placement, LinkURL: req.LinkURL, OpenInNewTab: req.OpenInNewTab, StartsAt: req.StartsAt, EndsAt: req.EndsAt})
	if err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{Message: err.Error()})
	}
	return c.JSON(http.StatusCreated, banner)
}

func (h *BannerHandler) GetAll(c echo.Context) error {
	items, err := h.useCase.GetAll(c.Request().Context())
	if err != nil {
		return handler.RespondError(c, http.StatusInternalServerError, err)
	}
	if items == nil {
		items = []model.Banner{}
	}
	return c.JSON(http.StatusOK, items)
}

func (h *BannerHandler) GetByID(c echo.Context) error {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	banner, err := h.useCase.GetByID(c.Request().Context(), id)
	if err != nil {
		return handler.RespondError(c, http.StatusInternalServerError, err)
	}
	if banner == nil {
		return c.JSON(http.StatusNotFound, handler.ErrorResponse{Message: "指定されたバナーが見つかりませんでした。"})
	}
	return c.JSON(http.StatusOK, banner)
}

func (h *BannerHandler) Update(c echo.Context) error {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req BannerSaveRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{Message: err.Error()})
	}
	banner, err := h.useCase.Update(c.Request().Context(), id, &model.Banner{Title: req.Title, Placement: req.Placement, LinkURL: req.LinkURL, OpenInNewTab: req.OpenInNewTab, StartsAt: req.StartsAt, EndsAt: req.EndsAt})
	if err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{Message: err.Error()})
	}
	return c.JSON(http.StatusOK, banner)
}

func (h *BannerHandler) Delete(c echo.Context) error {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	banner, err := h.useCase.Delete(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{Message: err.Error()})
	}
	if err := deleteStoredImage(banner.ImageKey); err != nil {
		return handler.RespondError(c, http.StatusInternalServerError, err)
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *BannerHandler) UploadImage(c echo.Context) error {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	file, err := c.FormFile("image")
	if err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{Message: err.Error()})
	}
	banner, err := h.useCase.GetByID(c.Request().Context(), id)
	if err != nil || banner == nil {
		return c.JSON(http.StatusNotFound, handler.ErrorResponse{Message: "指定されたバナーが見つかりませんでした。"})
	}
	imageKey, err := saveUploadedImage(file, "banners", strconv.FormatInt(id, 10))
	if err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{Message: err.Error()})
	}
	updated, err := h.useCase.UpdateImage(c.Request().Context(), id, imageKey)
	if err != nil {
		return handler.RespondError(c, http.StatusInternalServerError, err)
	}
	if banner.ImageKey != imageKey {
		_ = deleteStoredImage(banner.ImageKey)
	}
	return c.JSON(http.StatusOK, updated)
}

func (h *BannerHandler) UpdateOrder(c echo.Context) error {
	var req BannerOrderRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{Message: err.Error()})
	}
	if err := h.useCase.UpdateOrder(c.Request().Context(), req.Placement, req.IDs); err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{Message: err.Error()})
	}
	return c.NoContent(http.StatusNoContent)
}
