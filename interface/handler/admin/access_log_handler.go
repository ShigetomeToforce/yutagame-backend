package admin

import (
	"net/http"
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

func (h *AccessLogHandler) GetAll(c echo.Context) error {
	return handler.HandleListOrPagination(
		c,
		h.accessLogUseCase.GetAllAccessLogs,
		h.accessLogUseCase.GetAccessLogsWithPagination,
		func(c echo.Context) usecaseAdmin.AccessLogListFilter {
			return usecaseAdmin.AccessLogListFilter{
				SearchWord: c.QueryParam("q"),
				EventType:  c.QueryParam("eventType"),
				FromDate:   c.QueryParam("fromDate"),
				ToDate:     c.QueryParam("toDate"),
			}
		},
	)
}

func (h *AccessLogHandler) GetDashboard(c echo.Context) error {
	dashboard, err := h.accessLogUseCase.GetDashboard(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, handler.ErrorResponse{Message: err.Error()})
	}
	return c.JSON(http.StatusOK, dashboard)
}
