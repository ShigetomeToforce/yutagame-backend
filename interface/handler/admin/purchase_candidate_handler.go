package admin

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	usecaseAdmin "yutagame-backend/application/usecase/admin"
	"yutagame-backend/domain/model"
	"yutagame-backend/interface/handler"

	"github.com/labstack/echo/v4"
)

type PurchaseCandidateSaveRequest struct {
	Name            string `json:"name"`
	Kana            string `json:"kana"`
	Code            string `json:"code"`
	ListPrice       *int32 `json:"listPrice"`
	OfficialSiteURL string `json:"officialSiteUrl"`
	YouTubeURL      string `json:"youtubeUrl"`
	ReleaseDateText string `json:"releaseDateText"`
	ManufacturerID  int64  `json:"manufacturerId"`
	MachineID       int64  `json:"machineId"`
	GenreID         *int64 `json:"genreId"`
	IsPurchased     bool   `json:"isPurchased"`
}

type PurchaseCandidateOrderRequest struct {
	IsPurchased bool    `json:"isPurchased"`
	IDs         []int64 `json:"ids"`
}

type PurchaseCandidateHandler struct {
	useCase *usecaseAdmin.PurchaseCandidateUseCase
}

func NewPurchaseCandidateHandler(useCase *usecaseAdmin.PurchaseCandidateUseCase) *PurchaseCandidateHandler {
	return &PurchaseCandidateHandler{useCase: useCase}
}

func candidateFromRequest(req PurchaseCandidateSaveRequest) *model.PurchaseCandidate {
	return &model.PurchaseCandidate{Name: req.Name, Kana: req.Kana, Code: req.Code, ListPrice: req.ListPrice, OfficialSiteURL: req.OfficialSiteURL, YouTubeURL: req.YouTubeURL, ReleaseDateText: req.ReleaseDateText, ManufacturerID: req.ManufacturerID, MachineID: req.MachineID, GenreID: req.GenreID, IsPurchased: req.IsPurchased}
}

func (h *PurchaseCandidateHandler) GetAll(c echo.Context) error {
	isPurchased := false
	if strings.EqualFold(strings.TrimSpace(c.QueryParam("isPurchased")), "true") {
		isPurchased = true
	}
	items, err := h.useCase.GetAllByPurchased(c.Request().Context(), isPurchased)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, handler.ErrorResponse{Message: err.Error()})
	}
	return c.JSON(http.StatusOK, items)
}

func (h *PurchaseCandidateHandler) GetByID(c echo.Context) error {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	item, err := h.useCase.GetByID(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{Message: err.Error()})
	}
	if item == nil {
		return c.JSON(http.StatusNotFound, handler.ErrorResponse{Message: "指定された購入候補が見つかりませんでした。"})
	}
	return c.JSON(http.StatusOK, item)
}

func (h *PurchaseCandidateHandler) Create(c echo.Context) error {
	var req PurchaseCandidateSaveRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{Message: err.Error()})
	}
	item, err := h.useCase.Create(c.Request().Context(), candidateFromRequest(req))
	if err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{Message: err.Error()})
	}
	return c.JSON(http.StatusCreated, item)
}

func (h *PurchaseCandidateHandler) Update(c echo.Context) error {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req PurchaseCandidateSaveRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{Message: err.Error()})
	}
	item, err := h.useCase.Update(c.Request().Context(), id, candidateFromRequest(req))
	if err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{Message: err.Error()})
	}
	return c.JSON(http.StatusOK, item)
}

func (h *PurchaseCandidateHandler) Delete(c echo.Context) error {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	item, err := h.useCase.GetByID(c.Request().Context(), id)
	if err != nil || item == nil {
		return c.JSON(http.StatusNotFound, handler.ErrorResponse{Message: "指定された購入候補が見つかりませんでした。"})
	}
	if err := h.useCase.Delete(c.Request().Context(), id); err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{Message: err.Error()})
	}
	if item.ImageKey != nil {
		_ = deleteStoredImage(*item.ImageKey)
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *PurchaseCandidateHandler) UpdateOrder(c echo.Context) error {
	var req PurchaseCandidateOrderRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{Message: err.Error()})
	}
	if err := h.useCase.UpdateOrder(c.Request().Context(), req.IsPurchased, req.IDs); err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{Message: err.Error()})
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *PurchaseCandidateHandler) UploadImage(c echo.Context) error {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	fileHeader, err := c.FormFile("image")
	if err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{Message: "画像ファイルを選択してください。"})
	}
	item, err := h.useCase.GetByID(c.Request().Context(), id)
	if err != nil || item == nil {
		return c.JSON(http.StatusNotFound, handler.ErrorResponse{Message: "指定された購入候補が見つかりませんでした。"})
	}
	imageKey, err := saveUploadedImage(fileHeader, "purchase-candidates", strconv.FormatInt(id, 10))
	if err != nil {
		status := http.StatusBadRequest
		if !errors.Is(err, errImageRequired) && !errors.Is(err, errImageTooLarge) && !errors.Is(err, errImageTypeDenied) {
			status = http.StatusInternalServerError
		}
		return c.JSON(status, handler.ErrorResponse{Message: err.Error()})
	}
	updated, err := h.useCase.UpdateImage(c.Request().Context(), id, imageKey)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, handler.ErrorResponse{Message: err.Error()})
	}
	if item.ImageKey != nil && *item.ImageKey != imageKey {
		_ = deleteStoredImage(*item.ImageKey)
	}
	return c.JSON(http.StatusOK, updated)
}

func (h *PurchaseCandidateHandler) DeleteImage(c echo.Context) error {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	item, err := h.useCase.GetByID(c.Request().Context(), id)
	if err != nil || item == nil {
		return c.JSON(http.StatusNotFound, handler.ErrorResponse{Message: "指定された購入候補が見つかりませんでした。"})
	}
	updated, err := h.useCase.DeleteImage(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, handler.ErrorResponse{Message: err.Error()})
	}
	if item.ImageKey != nil {
		_ = deleteStoredImage(*item.ImageKey)
	}
	return c.JSON(http.StatusOK, updated)
}
