package admin

import (
	"net/http"
	"strconv"
	"yutagame-backend/application/usecase/admin"
	"yutagame-backend/domain/model"
	"yutagame-backend/interface/handler"

	"github.com/labstack/echo/v4"
)

// =========================================================================
// 構造体＆コンストラクタ
// =========================================================================

// ManufacturerSaveRequest メーカーの新規登録および情報更新時に共通で利用するリクエストデータ
type ManufacturerSaveRequest struct {
	Name     string `json:"name"`
	Kana     string `json:"kana"`
	Overview string `json:"overview"`
	Code     string `json:"code"`
}

// ManufacturerHandler メーカーに関連するHTTPリクエストの受付とレスポンスの制御を担当するハンドラー
type ManufacturerHandler struct {
	manufacturerUseCase *admin.ManufacturerUseCase
}

// NewManufacturerHandler ManufacturerHandlerの新しいインスタンスを生成するコンストラクタ
func NewManufacturerHandler(manufacturerUseCase *admin.ManufacturerUseCase) *ManufacturerHandler {
	return &ManufacturerHandler{manufacturerUseCase: manufacturerUseCase}
}

// =========================================================================
// 🛠️ Manufacturer Management CRUD (ジャンル管理エンドポイント) - ガードあり
// =========================================================================

// -------------------------------------------------------------------------
// C: Create (作成)
// -------------------------------------------------------------------------

// Create メーカー新規登録
// @Summary      メーカー新規登録
// @Description  新しいメーカーを作成します。
// @Tags         Manufacturers
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body   ManufacturerSaveRequest true "メーカー情報"
// @Success      201  {object}  model.Manufacturer
// @Failure      400  {object}  handler.ErrorResponse "バリデーションエラー"
// @Router       /admin/manufacturers [post]
func (h *ManufacturerHandler) Create(c echo.Context) error {
	var req ManufacturerSaveRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{Message: err.Error()})
	}

	manufacturer := &model.Manufacturer{
		Name:     req.Name,
		Kana:     req.Kana,
		Overview: req.Overview,
		Code:     req.Code,
	}

	ctx := c.Request().Context()
	if err := h.manufacturerUseCase.CreateManufacturer(ctx, manufacturer); err != nil {
		return c.JSON(http.StatusInternalServerError, handler.ErrorResponse{Message: err.Error()})
	}

	return c.JSON(http.StatusCreated, manufacturer)
}

// -------------------------------------------------------------------------
// R: Read (取得)
// -------------------------------------------------------------------------

// GetByID メーカー詳細取得
// @Summary      メーカー詳細取得
// @Description  指定されたIDのメーカー情報を取得します。
// @Tags         Manufacturers
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "メーカーID"
// @Success      200  {object}  model.Manufacturer
// @Failure      404  {object}  handler.ErrorResponse "未検出エラー"
// @Router       /admin/manufacturers/{id} [get]
func (h *ManufacturerHandler) GetByID(c echo.Context) error {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	ctx := c.Request().Context()

	manufacturer, err := h.manufacturerUseCase.GetManufacturerByID(ctx, id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{
			Message: err.Error(),
		})
	}
	if manufacturer == nil {
		return c.JSON(http.StatusNotFound, handler.ErrorResponse{
			Message: "指定されたIDのメーカー情報が見つかりませんでした。",
		})
	}

	return c.JSON(http.StatusOK, manufacturer)
}

// GetAll メーカー一覧取得
// @Summary      メーカー一覧取得
// @Tags         Manufacturers
// @Produce      json
// @Security     BearerAuth
// @Param        page  query     int  false  "ページ番号 (指定するとページングモード)"
// @Param        limit query     int  false  "表示件数 (10, 30, 50)"
// @Param        q     query     string false "検索キーワード (名前またはカナの部分一致)"
// @Success      200   {array}   model.Manufacturer "page未指定時"
// @Success      200   {object}  handler.PaginatedResponse[model.Manufacturer] "page指定時"
// @Router       /admin/manufacturers [get]
func (h *ManufacturerHandler) GetAll(c echo.Context) error {
	// 💡 共通のジェネリクス関数に全件・ページングの各ユースケース関数を渡して処理を委ねる
	return handler.HandleListOrPagination(
		c,
		h.manufacturerUseCase.GetAllManufacturers,
		h.manufacturerUseCase.GetManufacturersWithPagination,
		func(c echo.Context) admin.ManufactureListFilter {
			return admin.ManufactureListFilter{
				SearchWord: c.QueryParam("q"),
			}
		},
	)
}

// -------------------------------------------------------------------------
// U: Update (更新)
// -------------------------------------------------------------------------

// Update メーカー情報更新
// @Summary      メーカー情報更新
// @Description  指定されたIDのメーカーの情報を更新します。
// @Tags         Manufacturers
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id      path   int  true  "メーカーID"
// @Param        request body   ManufacturerSaveRequest true "メーカー更新情報"
// @Success      200  {object}  model.Genre
// @Failure      400  {object}  handler.ErrorResponse "エラー"
// @Router       /admin/manufacturers/{id} [put]
func (h *ManufacturerHandler) Update(c echo.Context) error {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req ManufacturerSaveRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{Message: err.Error()})
	}

	manufacturer := &model.Manufacturer{
		ID:       id,
		Name:     req.Name,
		Kana:     req.Kana,
		Overview: req.Overview,
		Code:     req.Code,
	}

	ctx := c.Request().Context()
	if err := h.manufacturerUseCase.UpdateManufacturer(ctx, manufacturer); err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{Message: err.Error()})
	}

	return c.JSON(http.StatusOK, manufacturer)
}

// -------------------------------------------------------------------------
// D: Delete (削除)
// -------------------------------------------------------------------------

// Delete メーカー削除
// @Summary      メーカー削除
// @Description  指定されたIDのメーカーを削除します。
// @Tags         Manufacturers
// @Security     BearerAuth
// @Param        id   path      int  true  "ジメーカID"
// @Success      204  "No Content"
// @Failure      400  {object}  handler.ErrorResponse "エラー"
// @Router       /admin/manufacturers/{id} [delete]
func (h *ManufacturerHandler) Delete(c echo.Context) error {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	ctx := c.Request().Context()

	if err := h.manufacturerUseCase.DeleteManufacturer(ctx, id); err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{Message: err.Error()})
	}
	return c.NoContent(http.StatusNoContent)
}
