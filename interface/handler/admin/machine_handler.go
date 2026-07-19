package admin

import (
	"net/http"
	"strconv"
	"yutagame-backend/application/usecase/admin"
	"yutagame-backend/interface/handler"

	"yutagame-backend/domain/model"

	"github.com/labstack/echo/v4"
)

type MachineHandler struct {
	machineUseCase *admin.MachineUseCase
}

func NewMachineHandler(machineUseCase *admin.MachineUseCase) *MachineHandler {
	return &MachineHandler{machineUseCase: machineUseCase}
}

type MachineCreateRequest struct {
	Name           string `json:"name"`
	Kana           string `json:"kana"`
	Overview       string `json:"overview"`
	Code           string `json:"code"`
	Abbreviation   string `json:"abbreviation"`
	ManufacturerID int64  `json:"manufacturerId"`
	MachineType    string `json:"machineType"`
	ReleaseDate    string `json:"releaseDate"` // フロントからは文字列("YYYY-MM-DD")で受け取る
	SortOrder      int32  `json:"sortOrder"`
}

// GetAll 機種一覧取得
// @Summary      機種一覧取得
// @Description  登録されているすべてのハードウェア（機種）マスタを取得します。
// @Tags         Machines
// @Produce      json
// @Security     BearerAuth
// @Success      200  {array}  model.Machine
// @Router       /admin/machines [get]
func (h *MachineHandler) GetAll(c echo.Context) error {
	ctx := c.Request().Context()
	machines, err := h.machineUseCase.GetAllMachines(ctx)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, handler.ErrorResponse{
			Message: err.Error(),
		})
	}
	return c.JSON(http.StatusOK, machines)
}

// GetByID 機種詳細取得
// @Summary      機種詳細取得
// @Description  指定されたIDの機種情報を取得します。
// @Tags         Machines
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "機種ID"
// @Success      200  {object}  model.Machine
// @Router       /admin/machines/{id} [get]
func (h *MachineHandler) GetByID(c echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{
			Message: "不正なID形式です。整数値を指定してください。",
		})
	}

	ctx := c.Request().Context()
	machine, err := h.machineUseCase.GetMachineByID(ctx, id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{
			Message: err.Error(),
		})
	}
	if machine == nil {
		return c.JSON(http.StatusNotFound, handler.ErrorResponse{
			Message: "指定されたIDの機種情報が見つかりませんでした。",
		})
	}

	return c.JSON(http.StatusOK, machine)
}

// Create 機種新規登録
// @Summary      機種新規登録
// @Description  新しいハードウェア（機種）マスタを登録します。
// @Tags         Machines
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body   MachineCreateRequest true "機種登録情報"
// @Success      201  {object}  model.Machine
// @Router       /admin/machines [post]
func (h *MachineHandler) Create(c echo.Context) error {
	var m model.Machine
	if err := c.Bind(&m); err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{
			Message: err.Error(),
		})
	}

	ctx := c.Request().Context()
	if err := h.machineUseCase.CreateMachine(ctx, &m); err != nil {
		return c.JSON(http.StatusInternalServerError, handler.ErrorResponse{
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusCreated, m)
}
