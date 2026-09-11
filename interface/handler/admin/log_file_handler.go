package admin

import (
	"net/http"
	"strconv"
	"strings"
	"yutagame-backend/infrastructure/filelog"
	"yutagame-backend/interface/handler"

	"github.com/labstack/echo/v4"
)

const logRetentionDays = 30

type LogFileHandler struct {
	logger *filelog.DailyLogger
}

func NewLogFileHandler(logger *filelog.DailyLogger) *LogFileHandler {
	return &LogFileHandler{logger: logger}
}

func (h *LogFileHandler) GetAll(c echo.Context) error {
	if err := h.logger.CleanupOldFiles(logRetentionDays); err != nil {
		return handler.RespondError(c, http.StatusInternalServerError, err)
	}

	page := 1
	if raw := strings.TrimSpace(c.QueryParam("page")); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n > 0 {
			page = n
		}
	}

	limit := 30
	if raw := strings.TrimSpace(c.QueryParam("limit")); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n > 0 {
			limit = n
		}
	}

	result, err := h.logger.Read(filelog.ReadOptions{
		Scope: c.QueryParam("scope"),
		Kind:  c.QueryParam("kind"),
		Date:  c.QueryParam("date"),
		Level: c.QueryParam("level"),
		Query: c.QueryParam("q"),
		Page:  page,
		Limit: limit,
	})
	if err != nil {
		return handler.RespondError(c, http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusOK, result)
}
