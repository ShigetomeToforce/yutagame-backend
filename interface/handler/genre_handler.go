package handler

import (
	"net/http"
	"yutagame-backend/application/usecase"

	"github.com/labstack/echo/v4"
)

type GenreHandler struct {
	genreUseCase *usecase.GenreUseCase
}

func NewGenreHandler(genreUseCase *usecase.GenreUseCase) *GenreHandler {
	return &GenreHandler{genreUseCase: genreUseCase}
}

func (h *GenreHandler) GetAll(c echo.Context) error {
	ctx := c.Request().Context()
	genres, err := h.genreUseCase.GetAllGenres(ctx)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, genres)
}
