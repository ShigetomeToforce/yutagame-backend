package handler

import (
	"net/http"
	"yutagame-backend/application/usecase"

	"github.com/labstack/echo/v4"
)

type ManufacturerHandler struct {
	manufacturerUseCase *usecase.ManufacturerUseCase
}

func NewManufacturerHandler(manufacturerUseCase *usecase.ManufacturerUseCase) *ManufacturerHandler {
	return &ManufacturerHandler{manufacturerUseCase: manufacturerUseCase}
}

// GetAll メーカー一覧取得
// @Summary      メーカー一覧取得
// @Tags         Masters
// @Security     BearerAuth
// @Success      200  {array}  model.Manufacturer
// @Router       /admin/manufacturers [get]
func (h *ManufacturerHandler) GetAll(c echo.Context) error {
	ctx := c.Request().Context()
	manufacturers, err := h.manufacturerUseCase.GetAllManufacturers(ctx)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, manufacturers)
}
