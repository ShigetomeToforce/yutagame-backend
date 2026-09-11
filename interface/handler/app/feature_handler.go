package app

import (
	"net/http"
	"strings"
	usecaseApp "yutagame-backend/application/usecase/app"
	"yutagame-backend/domain/model"
	"yutagame-backend/interface/handler"

	"github.com/labstack/echo/v4"
)

type FeatureHandler struct {
	featureUseCase      *usecaseApp.FeaturePublicUseCase
	analyticsLogUseCase *usecaseApp.AnalyticsLogUseCase
}

func NewFeatureHandler(featureUseCase *usecaseApp.FeaturePublicUseCase, analyticsLogUseCase *usecaseApp.AnalyticsLogUseCase) *FeatureHandler {
	return &FeatureHandler{featureUseCase: featureUseCase, analyticsLogUseCase: analyticsLogUseCase}
}

func (h *FeatureHandler) GetAll(c echo.Context) error {
	items, err := h.featureUseCase.GetPublishedFeatures(c.Request().Context())
	if err != nil {
		return handler.RespondError(c, http.StatusInternalServerError, err)
	}
	if items == nil {
		items = []model.Feature{}
	}
	return c.JSON(http.StatusOK, items)
}

func (h *FeatureHandler) GetByCode(c echo.Context) error {
	code := strings.TrimSpace(c.Param("code"))
	if code == "" {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{Message: "特集コードを指定してください。"})
	}
	item, err := h.featureUseCase.GetPublishedFeatureByCode(c.Request().Context(), code)
	if err != nil {
		return handler.RespondError(c, http.StatusInternalServerError, err)
	}
	if item == nil {
		return c.JSON(http.StatusNotFound, handler.ErrorResponse{Message: "指定された特集が見つかりませんでした。"})
	}
	if h.analyticsLogUseCase != nil {
		if logErr := h.analyticsLogUseCase.RecordContentAccess(c.Request().Context(), usecaseApp.ContentAccessLogInput{
			VisitorID:   resolveVisitorID(c),
			IP:          c.RealIP(),
			ContentType: "feature",
			ContentKey:  item.Code,
		}); logErr != nil {
			c.Logger().Warnf("feature access log write failed: %v", logErr)
		}
	}
	return c.JSON(http.StatusOK, item)
}
