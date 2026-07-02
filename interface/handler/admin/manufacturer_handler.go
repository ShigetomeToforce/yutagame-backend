package admin

import (
	"net/http"
	"yutagame-backend/application/usecase/admin"
	"yutagame-backend/interface/handler"

	"github.com/labstack/echo/v4"
)

type ManufacturerHandler struct {
	manufacturerUseCase *admin.ManufacturerUseCase
}

func NewManufacturerHandler(manufacturerUseCase *admin.ManufacturerUseCase) *ManufacturerHandler {
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
		return c.JSON(http.StatusInternalServerError, handler.ErrorResponse{
			Message: err.Error(),
		})
	}
	return c.JSON(http.StatusOK, manufacturers)
}
