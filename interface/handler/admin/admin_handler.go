package admin

import (
	"net/http"
	"strconv"
	"yutagame-backend/application/usecase/admin"
	"yutagame-backend/interface/handler"

	"github.com/labstack/echo/v4"
)

// =========================================================================
// 構造体＆コンストラクタ
// =========================================================================

// LoginRequest ログイン時にクライアントから送信されるJSONリクエスト
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// AdminSaveRequest 管理アカウントの新規登録および情報更新時に共通で利用するリクエストデータ
type AdminSaveRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"` // 更新処理時、空文字の場合は「変更なし」として取り扱われます
	RoleType string `json:"roleType"`
}

// AdminHandler 管理アカウントに関連するHTTPリクエストの受付とレスポンスの制御を担当するハンドラー
type AdminHandler struct {
	adminUseCase *admin.AdminUseCase
}

// NewAdminHandler AdminHandlerの新しいインスタンスを生成するコンストラクタ
func NewAdminHandler(adminUseCase *admin.AdminUseCase) *AdminHandler {
	return &AdminHandler{adminUseCase: adminUseCase}
}

// =========================================================================
// 🔑 Authentication (認証エンドポイント) - ガードなし
// =========================================================================

// Login 管理アカウントログイン
// @Summary      管理アカウントログイン
// @Description  メールアドレスとパスワードでログインし、JWTトークンを発行します。
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request body LoginRequest true "ログイン情報"
// @Success      200 {object} map[string]string "tokenが返ります"
// @Failure      401 {object} handler.ErrorResponse "認証エラー"
// @Router       /admin/login [post]
func (h *AdminHandler) Login(c echo.Context) error {
	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{Message: err.Error()})
	}

	ctx := c.Request().Context()
	token, err := h.adminUseCase.Login(ctx, req.Email, req.Password)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, handler.ErrorResponse{
			Message: "メールアドレスまたはパスワードが違います。",
		})
	}

	return c.JSON(http.StatusOK, echo.Map{"token": token})
}

// =========================================================================
// 🛠️ Admin Management CRUD (管理アカウント管理エンドポイント) - ガードあり
// =========================================================================

// -------------------------------------------------------------------------
// C: Create (作成)
// -------------------------------------------------------------------------

// Create 管理アカウント新規登録
// @Summary      管理アカウント新規登録
// @Description  新しい管理アカウントフアカウントを作成します。
// @Tags         Admins
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body   AdminSaveRequest true "管理アカウント登録情報"
// @Success      201  {object}  model.Admin
// @Failure      400  {object}  handler.ErrorResponse "バリデーション・重複エラー"
// @Router       /admin/admins [post]
func (h *AdminHandler) Create(c echo.Context) error {
	var req AdminSaveRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{Message: err.Error()})
	}

	ctx := c.Request().Context()
	adminData, err := h.adminUseCase.CreateAdmin(ctx, req.Name, req.Email, req.Password, req.RoleType)
	if err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{Message: err.Error()})
	}
	return c.JSON(http.StatusCreated, adminData)
}

// -------------------------------------------------------------------------
// R: Read (取得)
// -------------------------------------------------------------------------

// GetByID 管理アカウント詳細取得
// @Summary      管理アカウント詳細取得
// @Description  指定されたIDの管理アカウント情報を取得します。
// @Tags         Admins
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "管理アカウントID"
// @Success      200  {object}  model.Admin
// @Failure      404  {object}  handler.ErrorResponse "未検出エラー"
// @Router       /admin/admins/{id} [get]
func (h *AdminHandler) GetByID(c echo.Context) error {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	ctx := c.Request().Context()

	admin, err := h.adminUseCase.GetAdminByID(ctx, id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{
			Message: err.Error(),
		})
	}
	if admin == nil {
		return c.JSON(http.StatusNotFound, handler.ErrorResponse{
			Message: "指定されたIDの管理アカウント情報が見つかりませんでした。",
		})
	}

	return c.JSON(http.StatusOK, admin)
}

// GetAll 管理アカウント一覧取得
// @Summary      管理アカウント一覧取得
// @Description  条件に従い管理アカウント情報の一覧を取得します。
// @Tags         Admins
// @Produce      json
// @Security     BearerAuth
// @Param        page  query     int  false  "ページ番号 (指定するとページングモード)"
// @Param        limit query     int  false  "表示件数 (10, 30, 50)"
// @Param        q     query     string false "自由入力のテキスト検索（管理アカウント名）"
// @Success      200   {array}   model.Admin "page未指定時"
// @Success      200   {object}  handler.PaginatedResponse[model.Admin] "page指定時"
// @Router       /admin/admins [get]
func (h *AdminHandler) GetAll(c echo.Context) error {
	// 💡 共通のジェネリクス関数に全件・ページングの各ユースケース関数を渡して処理を委ねる
	return handler.HandleListOrPagination(
		c,
		h.adminUseCase.GetAllAdmins,
		h.adminUseCase.GetAdminsWithPagination,
		func(c echo.Context) admin.AdminListFilter {
			return admin.AdminListFilter{
				SearchWord: c.QueryParam("q"),
			}
		},
	)
}

// -------------------------------------------------------------------------
// U: Update (更新)
// -------------------------------------------------------------------------

// Update 管理アカウント情報更新
// @Summary      管理アカウント情報更新
// @Description  指定されたIDの管理アカウントの名前、メール、パスワード、権限を更新します。
// @Tags         Admins
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id      path   int  true  "管理アカウントID"
// @Param        request body   AdminSaveRequest true "管理アカウント更新情報"
// @Success      200  {object}  model.Admin
// @Failure      400  {object}  handler.ErrorResponse "エラー"
// @Router       /admin/admins/{id} [put]
func (h *AdminHandler) Update(c echo.Context) error {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req AdminSaveRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{Message: err.Error()})
	}

	ctx := c.Request().Context()
	adminData, err := h.adminUseCase.UpdateAdmin(ctx, id, req.Name, req.Email, req.Password, req.RoleType)
	if err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{Message: err.Error()})
	}
	return c.JSON(http.StatusOK, adminData)
}

// -------------------------------------------------------------------------
// D: Delete (削除)
// -------------------------------------------------------------------------

// Delete 管理アカウント削除
// @Summary      管理アカウント削除
// @Description  指定されたIDの管理アカウントアカウントを削除します。
// @Tags         Admins
// @Security     BearerAuth
// @Param        id   path      int  true  "管理アカウントID"
// @Success      204  "No Content"
// @Failure      400  {object}  handler.ErrorResponse "エラー"
// @Router       /admin/admins/{id} [delete]
func (h *AdminHandler) Delete(c echo.Context) error {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	ctx := c.Request().Context()

	if err := h.adminUseCase.DeleteAdmin(ctx, id); err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{Message: err.Error()})
	}
	return c.NoContent(http.StatusNoContent)
}
