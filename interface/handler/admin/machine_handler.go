package admin

import (
	"net/http"
	"strconv"
	"time"
	"yutagame-backend/application/usecase/admin"
	"yutagame-backend/interface/handler"

	"yutagame-backend/domain/model"

	"github.com/labstack/echo/v4"
)

// =========================================================================
// 構造体＆コンストラクタ
// =========================================================================

// MachineSaveRequest 機種情報保存時にクライアントから送信されるJSONリクエスト
type MachineSaveRequest struct {
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

// MachineHandler 機種に関連するHTTPリクエストの受付とレスポンスの制御を担当するハンドラー
type MachineHandler struct {
	machineUseCase *admin.MachineUseCase
}

// NewMachineHandler MachineHandlerの新しいインスタンスを生成するコンストラクタ
func NewMachineHandler(machineUseCase *admin.MachineUseCase) *MachineHandler {
	return &MachineHandler{machineUseCase: machineUseCase}
}

// =========================================================================
// 🛠️ Machine Management CRUD (ゲーム管理エンドポイント) - ガードあり
// =========================================================================

// -------------------------------------------------------------------------
// C: Create (作成)
// -------------------------------------------------------------------------

// Create 機種新規登録
// @Summary      機種新規登録
// @Description  新しい機種情報をデータベースに登録します。
// @Tags         Machines
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body   MachineSaveRequest true "機種登録情報"
// @Success      201  {object}  model.Machine
// @Router       /admin/machines [post]
func (h *MachineHandler) Create(c echo.Context) error {
	var req MachineSaveRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{
			Message: err.Error(),
		})
	}

	releaseDate, err := time.Parse("2006-01-02", req.ReleaseDate)
	if err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{
			Message: "releaseDate の形式が不正です。YYYY-MM-DD を指定してください。",
		})
	}

	machine := model.Machine{
		Name:           req.Name,
		Kana:           req.Kana,
		Overview:       req.Overview,
		Code:           req.Code,
		Abbreviation:   req.Abbreviation,
		ManufacturerID: req.ManufacturerID,
		ReleaseDate:    releaseDate,
		SortOrder:      req.SortOrder,
	}

	ctx := c.Request().Context()
	if err := h.machineUseCase.CreateMachine(ctx, &machine); err != nil {
		return c.JSON(http.StatusInternalServerError, handler.ErrorResponse{
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusCreated, machine)
}

// -------------------------------------------------------------------------
// R: Read (取得)
// -------------------------------------------------------------------------

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
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
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

// GetAll 機種一覧取得
// @Summary      機種一覧取得
// @Description  登録されているすべてのハードウェア（機種）マスタを取得します。
// @Tags         Machines
// @Produce      json
// @Security     BearerAuth
// @Param        page  			 query   int  	false "ページ番号 (指定するとページングモード)"
// @Param        limit 			 query   int  	false "表示件数 (10, 30, 50)"
// @Param        q     			 query   string false "自由入力のテキスト検索（ゲーム名）"
// @Param        manufacturerIDs query   []int  false "メーカーIDの複数指定"
// @Success      200  {array}  model.Machine "page未指定時"
// @Success      200   {object}  handler.PaginatedResponse[model.Machine] "page指定時"
// @Router       /admin/machines [get]
func (h *MachineHandler) GetAll(c echo.Context) error {
	// 💡 共通のジェネリクス関数に全件・ページングの各ユースケース関数を渡して処理を委ねる
	return handler.HandleListOrPagination(
		c,
		h.machineUseCase.GetAllMachines,
		h.machineUseCase.GetMachinesWithPagination,
		func(c echo.Context) admin.MachineListFilter {
			return admin.MachineListFilter{
				SearchWord:      c.QueryParam("q"),
				ManufacturerIDs: handler.ParseInt64Array(c, "manufacturerIDs"),
			}
		},
	)
}

// -------------------------------------------------------------------------
// U: Update (更新)
// -------------------------------------------------------------------------

// Update 機種情報更新
// @Summary      機種情報更新
// @Description  指定されたIDの機種情報更新します。
// @Tags         Machines
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id      path   int  true  "機種ID"
// @Param        request body   MachineSaveRequest true "ゲーム更新情報"
// @Success      200  {object}  model.Machine
// @Failure      400  {object}  handler.ErrorResponse "エラー"
// @Router       /admin/machines/{id} [put]
func (h *MachineHandler) Update(c echo.Context) error {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req MachineSaveRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{Message: err.Error()})
	}

	releaseDate, err := time.Parse("2006-01-02", req.ReleaseDate)
	if err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{
			Message: "releaseDate の形式が不正です。YYYY-MM-DD を指定してください。",
		})
	}

	machine := model.Machine{
		ID:             id,
		Name:           req.Name,
		Kana:           req.Kana,
		Overview:       req.Overview,
		Code:           req.Code,
		Abbreviation:   req.Abbreviation,
		ManufacturerID: req.ManufacturerID,
		ReleaseDate:    releaseDate,
		SortOrder:      req.SortOrder,
	}

	ctx := c.Request().Context()
	if err := h.machineUseCase.UpdateMachine(ctx, &machine); err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusOK, machine)
}

// -------------------------------------------------------------------------
// D: Delete (削除)
// -------------------------------------------------------------------------

// Delete 機種情報削除
// @Summary      機種情報削除
// @Description  指定されたIDの機種情報を削除します。
// @Tags         Machines
// @Security     BearerAuth
// @Param        id   path      int  true  "機種ID"
// @Success      204  "No Content"
// @Failure      400  {object}  handler.ErrorResponse "エラー"
// @Router       /admin/machines/{id} [delete]
func (h *MachineHandler) Delete(c echo.Context) error {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	ctx := c.Request().Context()

	if err := h.machineUseCase.DeleteMachine(ctx, id); err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{Message: err.Error()})
	}
	return c.NoContent(http.StatusNoContent)
}
