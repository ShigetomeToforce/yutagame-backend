package admin

import (
	"net/http"
	"strconv"
	"yutagame-backend/application/usecase/admin"
	"yutagame-backend/domain/model"
	"yutagame-backend/interface/handler"

	"github.com/labstack/echo/v4"
)

// GenreHandler ジャンルに関連するHTTPリクエストの受付とレスポンスの制御を担当するハンドラー
type GenreHandler struct {
	genreUseCase *admin.GenreUseCase
}

// NewGenreHandler GenreHandler
func NewGenreHandler(genreUseCase *admin.GenreUseCase) *GenreHandler {
	return &GenreHandler{genreUseCase: genreUseCase}
}

// =========================================================================
// 📦 Request/Response DTO (データ転送構造体)
// =========================================================================

// GenreSaveRequest ジャンルの新規登録および情報更新時に共通で利用するリクエストデータ
type GenreSaveRequest struct {
	Name     string `json:"name"`
	Kana     string `json:"kana"`
	Overview string `json:"overview"`
	Code     string `json:"code"`
}

// =========================================================================
// 🛠️ Genre Management CRUD (ジャンル管理エンドポイント) - ガードあり
// =========================================================================

// -------------------------------------------------------------------------
// C: Create (作成)
// -------------------------------------------------------------------------

// Create ジャンル新規登録
// @Summary      ジャンル新規登録
// @Description  新しいジャンルを作成します。
// @Tags         Genre-Management
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body   GenreSaveRequest true "ジャンル登録情報"
// @Success      201  {object}  model.Genre
// @Failure      400  {object}  handler.ErrorResponse "バリデーションエラー"
// @Router       /admin/genres [post]
func (h *GenreHandler) Create(c echo.Context) error {
	var req GenreSaveRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{Message: err.Error()})
	}

	genre := &model.Genre{
		Name:     req.Name,
		Kana:     req.Kana,
		Overview: req.Overview,
		Code:     req.Code,
	}

	ctx := c.Request().Context()
	if err := h.genreUseCase.CreateGenre(ctx, genre); err != nil {
		return c.JSON(http.StatusInternalServerError, handler.ErrorResponse{Message: err.Error()})
	}

	return c.JSON(http.StatusCreated, genre)
}

// -------------------------------------------------------------------------
// R: Read (取得)
// -------------------------------------------------------------------------

// GetByID ジャンル詳細取得
// @Summary      ジャンル詳細取得
// @Description  指定されたIDのジャンル情報を取得します。
// @Tags         Genre-Management
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "ジャンルID"
// @Success      200  {object}  model.Genre
// @Failure      404  {object}  handler.ErrorResponse "未検出エラー"
// @Router       /admin/genres/{id} [get]
func (h *GenreHandler) GetByID(c echo.Context) error {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	ctx := c.Request().Context()

	genreData, err := h.genreUseCase.GetGenreByID(ctx, id)
	if err != nil || genreData == nil {
		return c.JSON(http.StatusNotFound, handler.ErrorResponse{
			Message: "指定されたIDのジャンル情報が見つかりませんでした。",
		})
	}
	return c.JSON(http.StatusOK, genreData)
}

// GetAll ジャンル一覧取得
// @Summary      ジャンル一覧取得
// @Tags         Genre-Management
// @Produce      json
// @Security     BearerAuth
// @Param        page  query     int  false  "ページ番号 (指定するとページングモード)"
// @Param        limit query     int  false  "表示件数 (10, 30, 50)"
// @Param        q     query     string false "検索キーワード (名前またはカナの部分一致)"
// @Success      200   {array}   model.Genre "page未指定時"
// @Success      200   {object}  handler.PaginatedResponse[model.Genre] "page指定時"
// @Router       /admin/genres [get]
func (h *GenreHandler) GetAll(c echo.Context) error {
	// 💡 共通のジェネリクス関数に全件・ページングの各ユースケース関数を渡して処理を委ねる
	return handler.HandleListOrPagination(
		c,
		h.genreUseCase.GetAllGenres,
		h.genreUseCase.GetGenresWithPagination,
	)
}

// -------------------------------------------------------------------------
// U: Update (更新)
// -------------------------------------------------------------------------

// Update ジャンル情報更新
// @Summary      ジャンル情報更新
// @Description  指定されたIDのジャンルの情報を更新します。
// @Tags         Genre-Management
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id      path   int  true  "ジャンルID"
// @Param        request body   GenreSaveRequest true "ジャンル更新情報"
// @Success      200  {object}  model.Genre
// @Failure      400  {object}  handler.ErrorResponse "エラー"
// @Router       /admin/genres/{id} [put]
func (h *GenreHandler) Update(c echo.Context) error {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req GenreSaveRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{Message: err.Error()})
	}

	genre := &model.Genre{
		ID:       id,
		Name:     req.Name,
		Kana:     req.Kana,
		Overview: req.Overview,
		Code:     req.Code,
	}

	ctx := c.Request().Context()
	if err := h.genreUseCase.UpdateGenre(ctx, genre); err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{Message: err.Error()})
	}

	return c.JSON(http.StatusCreated, genre)
}

// -------------------------------------------------------------------------
// D: Delete (削除)
// -------------------------------------------------------------------------

// Delete 管理者削除
// @Summary      ジャンル削除
// @Description  指定されたIDのジャンルを削除します。
// @Tags         Genre-Management
// @Security     BearerAuth
// @Param        id   path      int  true  "ジャンルID"
// @Success      204  "No Content"
// @Failure      400  {object}  handler.ErrorResponse "エラー"
// @Router       /admin/genres/{id} [delete]
func (h *GenreHandler) Delete(c echo.Context) error {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	ctx := c.Request().Context()

	if err := h.genreUseCase.DeleteGenre(ctx, id); err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{Message: err.Error()})
	}
	return c.NoContent(http.StatusNoContent)
}
