package handler

import (
	"net/http"
	"strconv"
	"yutagame-backend/application/usecase"
	"yutagame-backend/domain/model"

	"github.com/labstack/echo/v4"
)

type MachineHandler struct {
	machineUseCase *usecase.MachineUseCase
}

func NewMachineHandler(machineUseCase *usecase.MachineUseCase) *MachineHandler {
	return &MachineHandler{machineUseCase: machineUseCase}
}

// GetAll は機種一覧を返すAPIエンドポイントです
func (h *MachineHandler) GetAll(c echo.Context) error {
	ctx := c.Request().Context()
	machines, err := h.machineUseCase.GetAllMachines(ctx)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, machines)
}

// GetByID は指定されたIDの機種を1件返すAPIエンドポイントです
func (h *MachineHandler) GetByID(c echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid id format"})
	}

	ctx := c.Request().Context()
	machine, err := h.machineUseCase.GetMachineByID(ctx, id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}
	if machine == nil {
		return c.JSON(http.StatusNotFound, echo.Map{"error": "machine not found"})
	}

	return c.JSON(http.StatusOK, machine)
}

// Create は機種を新規登録するAPIエンドポイントです（管理画面用）
func (h *MachineHandler) Create(c echo.Context) error {
	var m model.Machine
	if err := c.Bind(&m); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}

	ctx := c.Request().Context()
	if err := h.machineUseCase.CreateMachine(ctx, &m); err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, m)
}
