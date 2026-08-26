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

// KeywordSaveRequest キーワードの新規登録および情報更新時に共通で利用するリクエストデータ
type KeywordSaveRequest struct {
	Name        string `json:"name"`
	Kana        string `json:"kana"`
	Overview    string `json:"overview"`
	Code        string `json:"code"`
	KeywordType string `json:"keywordType"`
	SortOrder   int32  `json:"sortOrder"`
}

// KeywordHandler キーワードに関連するHTTPリクエストの受付とレスポンスの制御を担当するハンドラー
type KeywordHandler struct {
	keywordUseCase *admin.KeywordUseCase
}

// NewKeywordHandler KeywordHandlerの新しいインスタンスを生成するコンストラクタ
func NewKeywordHandler(keywordUseCase *admin.KeywordUseCase) *KeywordHandler {
	return &KeywordHandler{keywordUseCase: keywordUseCase}
}

// =========================================================================
// 🛠️ Keyword Management CRUD (キーワード管理エンドポイント) - ガードあり
// =========================================================================

// -------------------------------------------------------------------------
// C: Create (作成)
// -------------------------------------------------------------------------

// Create キーワード新規登録
// @Summary      キーワード新規登録
// @Description  新しいキーワードを作成します。
// @Tags         Keywords
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body   KeywordSaveRequest true "キーワード登録情報"
// @Success      201  {object}  model.Keyword
// @Failure      400  {object}  handler.ErrorResponse "バリデーションエラー"
// @Router       /admin/keywords [post]
func (h *KeywordHandler) Create(c echo.Context) error {
	var req KeywordSaveRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{Message: err.Error()})
	}

	keyword := &model.Keyword{
		Name:        req.Name,
		Kana:        req.Kana,
		Overview:    req.Overview,
		Code:        req.Code,
		KeywordType: req.KeywordType,
		SortOrder:   req.SortOrder,
	}

	ctx := c.Request().Context()
	if err := h.keywordUseCase.CreateKeyword(ctx, keyword); err != nil {
		return c.JSON(http.StatusInternalServerError, handler.ErrorResponse{Message: err.Error()})
	}

	return c.JSON(http.StatusCreated, keyword)
}

// -------------------------------------------------------------------------
// R: Read (取得)
// -------------------------------------------------------------------------

// GetByID ジャンル詳細取得
// @Summary      ジャンル詳細取得
// @Description  指定されたIDのジャンル情報を取得します。
// @Tags         Keywords
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "キーワードID"
// @Success      200  {object}  model.Keyword
// @Failure      404  {object}  handler.ErrorResponse "未検出エラー"
// @Router       /admin/keywords/{id} [get]
func (h *KeywordHandler) GetByID(c echo.Context) error {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	ctx := c.Request().Context()

	keyword, err := h.keywordUseCase.GetKeywordByID(ctx, id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{
			Message: err.Error(),
		})
	}
	if keyword == nil {
		return c.JSON(http.StatusNotFound, handler.ErrorResponse{
			Message: "指定されたIDのキーワード情報が見つかりませんでした。",
		})
	}

	return c.JSON(http.StatusOK, keyword)
}

// GetAll キーワード一覧取得
// @Summary      キーワード一覧取得
// @Tags         Keywords
// @Produce      json
// @Security     BearerAuth
// @Param        page  query     int  false  "ページ番号 (指定するとページングモード)"
// @Param        limit query     int  false  "表示件数 (10, 30, 50)"
// @Param        q     query     string false "検索キーワード (名前またはカナの部分一致)"
// @Success      200   {array}   model.Genre "page未指定時"
// @Success      200   {object}  handler.PaginatedResponse[model.Keyword] "page指定時"
// @Router       /admin/keywords [get]
func (h *KeywordHandler) GetAll(c echo.Context) error {
	// 💡 共通のジェネリクス関数に全件・ページングの各ユースケース関数を渡して処理を委ねる
	return handler.HandleListOrPagination(
		c,
		h.keywordUseCase.GetAllKeywords,
		h.keywordUseCase.GetKeywordsWithPagination,
		func(c echo.Context) admin.KeywordListFilter {
			return admin.KeywordListFilter{
				SearchWord: c.QueryParam("q"),
			}
		},
	)
}

// -------------------------------------------------------------------------
// U: Update (更新)
// -------------------------------------------------------------------------

// Update キーワード情報更新
// @Summary      キーワード情報更新
// @Description  指定されたIDのキーワードの情報を更新します。
// @Tags         Keywords
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id      path   int  true  "キーワードID"
// @Param        request body   KeywordSaveRequest true "キーワード更新情報"
// @Success      200  {object}  model.Keyword
// @Failure      400  {object}  handler.ErrorResponse "エラー"
// @Router       /admin/keyword/{id} [put]
func (h *KeywordHandler) Update(c echo.Context) error {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req KeywordSaveRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{Message: err.Error()})
	}

	keyword := &model.Keyword{
		ID:          id,
		Name:        req.Name,
		Kana:        req.Kana,
		Overview:    req.Overview,
		Code:        req.Code,
		KeywordType: req.KeywordType,
		SortOrder:   req.SortOrder,
	}

	ctx := c.Request().Context()
	if err := h.keywordUseCase.UpdateKeyword(ctx, keyword); err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{Message: err.Error()})
	}

	return c.JSON(http.StatusOK, keyword)
}

// -------------------------------------------------------------------------
// D: Delete (削除)
// -------------------------------------------------------------------------

// Delete キーワード削除
// @Summary      キーワード削除
// @Description  指定されたIDのキーワードを削除します。
// @Tags         Keywords
// @Security     BearerAuth
// @Param        id   path      int  true  "キーワードID"
// @Success      204  "No Content"
// @Failure      400  {object}  handler.ErrorResponse "エラー"
// @Router       /admin/keywords/{id} [delete]
func (h *KeywordHandler) Delete(c echo.Context) error {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	ctx := c.Request().Context()

	if err := h.keywordUseCase.DeleteKeyword(ctx, id); err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{Message: err.Error()})
	}
	return c.NoContent(http.StatusNoContent)
}
