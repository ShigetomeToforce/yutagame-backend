package admin

import (
	"errors"
	"net/http"
	"strconv"
	usecaseAdmin "yutagame-backend/application/usecase/admin"
	"yutagame-backend/interface/handler"

	"github.com/labstack/echo/v4"
)

type FeatureSaveRequest struct {
	Code              string  `json:"code"`
	Title             string  `json:"title"`
	Excerpt           string  `json:"excerpt"`
	BodyHTML          string  `json:"bodyHtml"`
	ThumbnailImageKey string  `json:"thumbnailImageKey"`
	Status            string  `json:"status"`
	PublishStartAt    string  `json:"publishStartAt"`
	PublishEndAt      string  `json:"publishEndAt"`
	GameIDs           []int64 `json:"gameIds"`
}

type FeatureHandler struct {
	featureUseCase *usecaseAdmin.FeatureUseCase
}

func NewFeatureHandler(featureUseCase *usecaseAdmin.FeatureUseCase) *FeatureHandler {
	return &FeatureHandler{featureUseCase: featureUseCase}
}

func (h *FeatureHandler) Create(c echo.Context) error {
	var req FeatureSaveRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{Message: err.Error()})
	}
	publishStartAt, err := parsePublishDateTime(req.PublishStartAt)
	if err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{Message: "公開開始日時の形式が正しくありません。"})
	}
	publishEndAt, err := parsePublishDateTime(req.PublishEndAt)
	if err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{Message: "公開終了日時の形式が正しくありません。"})
	}
	feature, err := h.featureUseCase.CreateFeature(c.Request().Context(), req.Code, req.Title, req.Excerpt, req.BodyHTML, req.ThumbnailImageKey, req.Status, publishStartAt, publishEndAt, req.GameIDs)
	if err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{Message: err.Error()})
	}
	return c.JSON(http.StatusCreated, feature)
}

func (h *FeatureHandler) GetByID(c echo.Context) error {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	feature, err := h.featureUseCase.GetFeatureByID(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{Message: err.Error()})
	}
	if feature == nil {
		return c.JSON(http.StatusNotFound, handler.ErrorResponse{Message: "指定された特集が見つかりませんでした。"})
	}
	return c.JSON(http.StatusOK, feature)
}

func (h *FeatureHandler) GetAll(c echo.Context) error {
	return handler.HandleListOrPagination(
		c,
		h.featureUseCase.GetAllFeatures,
		h.featureUseCase.GetFeaturesWithPagination,
		func(c echo.Context) usecaseAdmin.FeatureListFilter {
			return usecaseAdmin.FeatureListFilter{SearchWord: c.QueryParam("q"), Status: c.QueryParam("status")}
		},
	)
}

func (h *FeatureHandler) Update(c echo.Context) error {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req FeatureSaveRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{Message: err.Error()})
	}
	publishStartAt, err := parsePublishDateTime(req.PublishStartAt)
	if err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{Message: "公開開始日時の形式が正しくありません。"})
	}
	publishEndAt, err := parsePublishDateTime(req.PublishEndAt)
	if err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{Message: "公開終了日時の形式が正しくありません。"})
	}
	feature, err := h.featureUseCase.UpdateFeature(c.Request().Context(), id, req.Code, req.Title, req.Excerpt, req.BodyHTML, req.ThumbnailImageKey, req.Status, publishStartAt, publishEndAt, req.GameIDs)
	if err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{Message: err.Error()})
	}
	return c.JSON(http.StatusOK, feature)
}

func (h *FeatureHandler) Delete(c echo.Context) error {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	feature, err := h.featureUseCase.GetFeatureByID(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{Message: err.Error()})
	}
	if feature == nil {
		return c.JSON(http.StatusNotFound, handler.ErrorResponse{Message: "指定された特集が見つかりませんでした。"})
	}
	if err := h.featureUseCase.DeleteFeature(c.Request().Context(), id); err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{Message: err.Error()})
	}
	if feature.ThumbnailImageKey != nil {
		_ = deleteStoredImage(*feature.ThumbnailImageKey)
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *FeatureHandler) GetPublishedForOrdering(c echo.Context) error {
	items, err := h.featureUseCase.GetPublishedFeaturesForOrdering(c.Request().Context())
	if err != nil {
		return handler.RespondError(c, http.StatusInternalServerError, err)
	}
	return c.JSON(http.StatusOK, items)
}

type FeatureOrderRequest struct {
	IDs []int64 `json:"ids"`
}

func (h *FeatureHandler) UpdateOrder(c echo.Context) error {
	var req FeatureOrderRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{Message: err.Error()})
	}
	if err := h.featureUseCase.UpdateFeatureOrder(c.Request().Context(), req.IDs); err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{Message: err.Error()})
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *FeatureHandler) UploadThumbnailImage(c echo.Context) error {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	fileHeader, err := c.FormFile("image")
	if err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{Message: "画像ファイルを選択してください。"})
	}
	feature, err := h.featureUseCase.GetFeatureByID(c.Request().Context(), id)
	if err != nil || feature == nil {
		return c.JSON(http.StatusNotFound, handler.ErrorResponse{Message: "指定された特集が見つかりませんでした。"})
	}
	imageKey, err := saveUploadedImage(fileHeader, "features", strconv.FormatInt(id, 10))
	if err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{Message: err.Error()})
	}
	updated, err := h.featureUseCase.UpdateThumbnailImage(c.Request().Context(), id, imageKey)
	if err != nil {
		return handler.RespondError(c, http.StatusInternalServerError, err)
	}
	if feature.ThumbnailImageKey != nil && *feature.ThumbnailImageKey != imageKey {
		_ = deleteStoredImage(*feature.ThumbnailImageKey)
	}
	return c.JSON(http.StatusOK, updated)
}

func (h *FeatureHandler) UploadContentImage(c echo.Context) error {
	fileHeader, err := c.FormFile("image")
	if err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{Message: "画像ファイルを選択してください。"})
	}
	imageKey, err := saveUploadedImage(fileHeader, "features", "content")
	if err != nil {
		status := http.StatusBadRequest
		if !errors.Is(err, errImageRequired) && !errors.Is(err, errImageTooLarge) && !errors.Is(err, errImageTypeDenied) {
			status = http.StatusInternalServerError
		}
		return c.JSON(status, handler.ErrorResponse{Message: err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"imageKey": imageKey})
}
