package app

import (
	"net/http"
	"strings"
	usecaseApp "yutagame-backend/application/usecase/app"
	"yutagame-backend/interface/handler"

	"github.com/labstack/echo/v4"
)

type AccessLogCreateRequest struct {
	EventType         string `json:"eventType"`
	EventSource       string `json:"eventSource"`
	Path              string `json:"path"`
	Method            string `json:"method"`
	StatusCode        int    `json:"statusCode"`
	VisitorID         string `json:"visitorId"`
	Referrer          string `json:"referrer"`
	MachineCode       string `json:"machineCode"`
	ManufacturerCode  string `json:"manufacturerCode"`
	GenreCode         string `json:"genreCode"`
	KeywordCode       string `json:"keywordCode"`
	SearchWord        string `json:"searchWord"`
	GameCode          string `json:"gameCode"`
	AffiliateCategory string `json:"affiliateCategory"`
}

type AccessLogHandler struct {
	logUseCase *usecaseApp.AccessLogPublicUseCase
}

func NewAccessLogHandler(logUseCase *usecaseApp.AccessLogPublicUseCase) *AccessLogHandler {
	return &AccessLogHandler{logUseCase: logUseCase}
}

func getCookieValue(rawCookie, key string) string {
	parts := strings.Split(rawCookie, ";")
	for _, part := range parts {
		pair := strings.SplitN(strings.TrimSpace(part), "=", 2)
		if len(pair) != 2 {
			continue
		}
		if pair[0] == key {
			return pair[1]
		}
	}
	return ""
}

func (h *AccessLogHandler) Create(c echo.Context) error {
	var req AccessLogCreateRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{Message: err.Error()})
	}

	visitorID := strings.TrimSpace(req.VisitorID)
	if visitorID == "" {
		visitorID = strings.TrimSpace(getCookieValue(c.Request().Header.Get("Cookie"), "visitor_id"))
	}

	path := strings.TrimSpace(req.Path)
	if path == "" {
		path = c.Request().URL.Path
	}

	method := strings.TrimSpace(req.Method)
	if method == "" {
		method = c.Request().Method
	}

	err := h.logUseCase.CreateEvent(c.Request().Context(), usecaseApp.PublicAccessEventInput{
		EventType:         req.EventType,
		EventSource:       req.EventSource,
		Path:              path,
		Method:            method,
		StatusCode:        req.StatusCode,
		VisitorID:         visitorID,
		IP:                c.RealIP(),
		UserAgent:         c.Request().UserAgent(),
		Referrer:          req.Referrer,
		MachineCode:       req.MachineCode,
		ManufacturerCode:  req.ManufacturerCode,
		GenreCode:         req.GenreCode,
		KeywordCode:       req.KeywordCode,
		SearchWord:        req.SearchWord,
		GameCode:          req.GameCode,
		AffiliateCategory: req.AffiliateCategory,
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, handler.ErrorResponse{Message: err.Error()})
	}

	return c.JSON(http.StatusCreated, map[string]any{"ok": true})
}
