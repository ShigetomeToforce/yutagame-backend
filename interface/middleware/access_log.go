package middleware

import (
	"strings"
	usecaseApp "yutagame-backend/application/usecase/app"

	"github.com/labstack/echo/v4"
)

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

func AccessLog(logUseCase *usecaseApp.AccessLogPublicUseCase) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			err := next(c)

			path := c.Request().URL.Path
			if path == "" {
				path = c.Path()
			}

			if strings.HasPrefix(path, "/api/") && !strings.HasPrefix(path, "/api/app/events") {
				query := c.QueryParams()
				visitorID := getCookieValue(c.Request().Header.Get("Cookie"), "visitor_id")
				statusCode := c.Response().Status
				if statusCode == 0 {
					statusCode = 200
				}

				_ = logUseCase.CreateEvent(c.Request().Context(), usecaseApp.PublicAccessEventInput{
					EventType:        "api_hit",
					EventSource:      "backend_api",
					Path:             path,
					Method:           c.Request().Method,
					StatusCode:       statusCode,
					VisitorID:        visitorID,
					IP:               c.RealIP(),
					UserAgent:        c.Request().UserAgent(),
					Referrer:         c.Request().Referer(),
					MachineCode:      query.Get("machineCode"),
					ManufacturerCode: query.Get("manufacturerCode"),
					GenreCode:        query.Get("genreCode"),
					KeywordCode:      query.Get("keywordCode"),
					SearchWord:       query.Get("q"),
				})

				if path == "/api/app/games" {
					_ = logUseCase.CreateEvent(c.Request().Context(), usecaseApp.PublicAccessEventInput{
						EventType:        "search",
						EventSource:      "backend_api",
						Path:             path,
						Method:           c.Request().Method,
						StatusCode:       statusCode,
						VisitorID:        visitorID,
						IP:               c.RealIP(),
						UserAgent:        c.Request().UserAgent(),
						Referrer:         c.Request().Referer(),
						MachineCode:      query.Get("machineCode"),
						ManufacturerCode: query.Get("manufacturerCode"),
						GenreCode:        query.Get("genreCode"),
						KeywordCode:      query.Get("keywordCode"),
						SearchWord:       query.Get("q"),
					})
				}
			}

			return err
		}
	}
}
