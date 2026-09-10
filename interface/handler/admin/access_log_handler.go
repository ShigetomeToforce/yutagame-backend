package admin

import (
	"net/http"
	"strconv"
	usecaseAdmin "yutagame-backend/application/usecase/admin"
	"yutagame-backend/interface/handler"

	"github.com/labstack/echo/v4"
)

type AccessLogHandler struct {
	accessLogUseCase *usecaseAdmin.AccessLogUseCase
}

func NewAccessLogHandler(accessLogUseCase *usecaseAdmin.AccessLogUseCase) *AccessLogHandler {
	return &AccessLogHandler{accessLogUseCase: accessLogUseCase}
}

func (h *AccessLogHandler) GetDashboard(c echo.Context) error {
	dashboard, err := h.accessLogUseCase.GetDashboard(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, handler.ErrorResponse{Message: err.Error()})
	}
	return c.JSON(http.StatusOK, dashboard)
}

func (h *AccessLogHandler) GetSearchBreakdown(c echo.Context) error {
	limit := 20
	if raw := c.QueryParam("limit"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil {
			limit = parsed
		}
	}

	result, err := h.accessLogUseCase.GetSearchBreakdown(
		c.Request().Context(),
		c.QueryParam("field"),
		c.QueryParam("date"),
		limit,
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, handler.ErrorResponse{Message: err.Error()})
	}

	return c.JSON(http.StatusOK, result)
}

func (h *AccessLogHandler) GetMonthlyTable(c echo.Context) error {
	result, err := h.accessLogUseCase.GetMonthlyTable(
		c.Request().Context(),
		c.QueryParam("month"),
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, handler.ErrorResponse{Message: err.Error()})
	}

	return c.JSON(http.StatusOK, result)
}

func (h *AccessLogHandler) GetMachineSearchDashboard(c echo.Context) error {
	limit := 20
	if raw := c.QueryParam("limit"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil {
			limit = parsed
		}
	}

	result, err := h.accessLogUseCase.GetMachineSearchDashboard(
		c.Request().Context(),
		c.QueryParam("period"),
		c.QueryParam("date"),
		c.QueryParam("month"),
		limit,
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, handler.ErrorResponse{Message: err.Error()})
	}

	return c.JSON(http.StatusOK, result)
}

func (h *AccessLogHandler) GetSearchRankingDashboard(c echo.Context) error {
	limit := 20
	if raw := c.QueryParam("limit"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil {
			limit = parsed
		}
	}

	result, err := h.accessLogUseCase.GetSearchRankingDashboard(
		c.Request().Context(),
		c.QueryParam("field"),
		c.QueryParam("period"),
		c.QueryParam("date"),
		c.QueryParam("month"),
		limit,
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, handler.ErrorResponse{Message: err.Error()})
	}

	return c.JSON(http.StatusOK, result)
}

func (h *AccessLogHandler) GetGameViewDashboard(c echo.Context) error {
	limit := 20
	if raw := c.QueryParam("limit"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil {
			limit = parsed
		}
	}

	result, err := h.accessLogUseCase.GetGameViewDashboard(
		c.Request().Context(),
		c.QueryParam("period"),
		c.QueryParam("date"),
		c.QueryParam("month"),
		limit,
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, handler.ErrorResponse{Message: err.Error()})
	}

	return c.JSON(http.StatusOK, result)
}
