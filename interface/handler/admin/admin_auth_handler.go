package admin

import (
	"net/http"
	"strconv"
	"yutagame-backend/application/usecase/admin"
	"yutagame-backend/interface/handler"

	"github.com/labstack/echo/v4"
)

// AdminAuthHandler 管理者に関連するHTTPリクエストの受付とレスポンスの制御を担当するハンドラー
type AdminAuthHandler struct {
	adminAuthUseCase *admin.AdminAuthUseCase
}

// NewAdminAuthHandler AdminAuthHandlerの新しいインスタンスを生成するコンストラクタ
func NewAdminAuthHandler(adminAuthUseCase *admin.AdminAuthUseCase) *AdminAuthHandler {
	return &AdminAuthHandler{adminAuthUseCase: adminAuthUseCase}
}

// =========================================================================
// 📦 Request/Response DTO (データ転送構造体)
// =========================================================================

// LoginRequest ログイン時にクライアントから送信されるJSONリクエスト
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// AdminSaveRequest 管理者の新規登録および情報更新時に共通で利用するリクエストデータ
type AdminSaveRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"` // 更新処理時、空文字の場合は「変更なし」として取り扱われます
	RoleType string `json:"roleType"`
}

// =========================================================================
// 🔑 Authentication (認証エンドポイント) - ガードなし
// =========================================================================

// Login 管理者ログイン
// @Summary      管理者ログイン
// @Description  メールアドレスとパスワードでログインし、JWTトークンを発行します。
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request body LoginRequest true "ログイン情報"
// @Success      200 {object} map[string]string "tokenが返ります"
// @Failure      401 {object} handler.ErrorResponse "認証エラー"
// @Router       /admin/login [post]
func (h *AdminAuthHandler) Login(c echo.Context) error {
	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{Message: err.Error()})
	}

	ctx := c.Request().Context()
	token, err := h.adminAuthUseCase.Login(ctx, req.Email, req.Password)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, handler.ErrorResponse{
			Message: "メールアドレスまたはパスワードが違います。",
		})
	}

	return c.JSON(http.StatusOK, echo.Map{"token": token})
}

// =========================================================================
// 🛠️ Admin Management CRUD (管理者管理エンドポイント) - ガードあり
// =========================================================================

// -------------------------------------------------------------------------
// C: Create (作成)
// -------------------------------------------------------------------------

// Create 管理者新規登録
// @Summary      管理者新規登録
// @Description  新しい管理者または一般権限のスタッフアカウントを作成します。
// @Tags         Admin-Management
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body   AdminSaveRequest true "管理者登録情報"
// @Success      201  {object}  model.Admin
// @Failure      400  {object}  handler.ErrorResponse "バリデーション・重複エラー"
// @Router       /admin/admins [post]
func (h *AdminAuthHandler) Create(c echo.Context) error {
	var req AdminSaveRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{Message: err.Error()})
	}

	ctx := c.Request().Context()
	adminData, err := h.adminAuthUseCase.CreateAdmin(ctx, req.Name, req.Email, req.Password, req.RoleType)
	if err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{Message: err.Error()})
	}
	return c.JSON(http.StatusCreated, adminData)
}

// -------------------------------------------------------------------------
// R: Read (取得)
// -------------------------------------------------------------------------

// GetByID 管理者詳細取得
// @Summary      管理者詳細取得
// @Description  指定されたIDの管理者情報を取得します。
// @Tags         Admin-Management
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "管理者ID"
// @Success      200  {object}  model.Admin
// @Failure      404  {object}  handler.ErrorResponse "未検出エラー"
// @Router       /admin/admins/{id} [get]
func (h *AdminAuthHandler) GetByID(c echo.Context) error {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	ctx := c.Request().Context()

	adminData, err := h.adminAuthUseCase.GetAdminByID(ctx, id)
	if err != nil || adminData == nil {
		return c.JSON(http.StatusNotFound, handler.ErrorResponse{
			Message: "指定されたIDの管理者情報が見つかりませんでした。",
		})
	}
	return c.JSON(http.StatusOK, adminData)
}

// GetAll 管理者一覧取得
// @Summary      管理者一覧取得
// @Tags         Admin-Management
// @Produce      json
// @Security     BearerAuth
// @Param        page  query     int  false  "ページ番号 (指定するとページングモード)"
// @Param        limit query     int  false  "表示件数 (10, 30, 50)"
// @Param        q     query     string false "検索キーワード (名前またはメールアドレスの部分一致)"
// @Success      200   {array}   model.Admin "page未指定時"
// @Success      200   {object}  handler.PaginatedResponse[model.Admin] "page指定時"
// @Router       /admin/admins [get]
func (h *AdminAuthHandler) GetAll(c echo.Context) error {
	// 💡 共通のジェネリクス関数に全件・ページングの各ユースケース関数を渡して処理を委ねる
	return handler.HandleListOrPagination(
		c,
		h.adminAuthUseCase.GetAllAdmins,
		h.adminAuthUseCase.GetAdminsWithPagination,
	)
}

// -------------------------------------------------------------------------
// U: Update (更新)
// -------------------------------------------------------------------------

// Update 管理者情報更新
// @Summary      管理者情報更新
// @Description  指定されたIDの管理者の名前、メール、パスワード、権限を更新します。
// @Tags         Admin-Management
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id      path   int  true  "管理者ID"
// @Param        request body   AdminSaveRequest true "管理者更新情報"
// @Success      200  {object}  model.Admin
// @Failure      400  {object}  handler.ErrorResponse "エラー"
// @Router       /admin/admins/{id} [put]
func (h *AdminAuthHandler) Update(c echo.Context) error {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req AdminSaveRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{Message: err.Error()})
	}

	ctx := c.Request().Context()
	adminData, err := h.adminAuthUseCase.UpdateAdmin(ctx, id, req.Name, req.Email, req.Password, req.RoleType)
	if err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{Message: err.Error()})
	}
	return c.JSON(http.StatusOK, adminData)
}

// -------------------------------------------------------------------------
// D: Delete (削除)
// -------------------------------------------------------------------------

// Delete 管理者削除
// @Summary      管理者削除
// @Description  指定されたIDの管理者アカウントを削除します。
// @Tags         Admin-Management
// @Security     BearerAuth
// @Param        id   path      int  true  "管理者ID"
// @Success      204  "No Content"
// @Failure      400  {object}  handler.ErrorResponse "エラー"
// @Router       /admin/admins/{id} [delete]
func (h *AdminAuthHandler) Delete(c echo.Context) error {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	ctx := c.Request().Context()

	if err := h.adminAuthUseCase.DeleteAdmin(ctx, id); err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{Message: err.Error()})
	}
	return c.NoContent(http.StatusNoContent)
}
