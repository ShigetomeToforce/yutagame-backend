package admin

import (
	"net/http"
	"yutagame-backend/application/usecase/admin"
	"yutagame-backend/interface/handler"

	"github.com/labstack/echo/v4"
)

type GenreHandler struct {
	genreUseCase *admin.GenreUseCase
}

func NewGenreHandler(genreUseCase *admin.GenreUseCase) *GenreHandler {
	return &GenreHandler{genreUseCase: genreUseCase}
}

// GetAll ジャンル一覧取得
// @Summary      ジャンル一覧取得
// @Tags         Masters
// @Security     BearerAuth
// @Success      200  {array}  model.Genre
// @Router       /admin/genres [get]
func (h *GenreHandler) GetAll(c echo.Context) error {
	ctx := c.Request().Context()
	genres, err := h.genreUseCase.GetAllGenres(ctx)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, handler.ErrorResponse{
			Message: err.Error(),
		})
	}
	return c.JSON(http.StatusOK, genres)
}
