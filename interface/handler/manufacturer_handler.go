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

func (h *ManufacturerHandler) GetAll(c echo.Context) error {
	ctx := c.Request().Context()
	manufacturers, err := h.manufacturerUseCase.GetAllManufacturers(ctx)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, manufacturers)
}
