package admin

import (
	"net/http"
	"strconv"
	"yutagame-backend/application/usecase/admin"

	"github.com/labstack/echo/v4"
)

type AdminAuthHandler struct {
	adminAuthUseCase *admin.AdminAuthUseCase
}

func NewAdminAuthHandler(adminAuthUseCase *admin.AdminAuthUseCase) *AdminAuthHandler {
	return &AdminAuthHandler{adminAuthUseCase: adminAuthUseCase}
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AdminSaveRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"` // 更新時は空文字なら「変更なし」の扱いにします
	RoleType string `json:"roleType"`
}

// Login 管理者ログイン
// @Summary      管理者ログイン
// @Description  メールアドレスとパスワードでログインし、JWTトークンを発行します。
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request body LoginRequest true "ログイン情報"
// @Success      200 {object} map[string]string "tokenが返ります"
// @Failure      401 {object} map[string]string "認証エラー"
// @Router       /admin/login [post]
func (h *AdminAuthHandler) Login(c echo.Context) error {
	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}

	ctx := c.Request().Context()
	token, err := h.adminAuthUseCase.Login(ctx, req.Email, req.Password)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, echo.Map{"token": token})
}

// GetAll 管理者一覧取得
// @Summary      管理者一覧取得
// @Description  登録されているすべての管理者アカウントのリストを取得します。
// @Tags         Admin-Management
// @Produce      json
// @Security     BearerAuth
// @Success      200  {array}   model.Admin
// @Router       /admin/admins [get]
func (h *AdminAuthHandler) GetAll(c echo.Context) error {
	ctx := c.Request().Context()
	admins, err := h.adminAuthUseCase.GetAllAdmins(ctx)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, admins)
}

// GetByID 管理者詳細取得
// @Summary      管理者詳細取得
// @Description  指定されたIDの管理者情報を取得します。
// @Tags         Admin-Management
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "管理者ID"
// @Success      200  {object}  model.Admin
// @Failure      404  {object}  map[string]string "未検出エラー"
// @Router       /admin/admins/{id} [get]
func (h *AdminAuthHandler) GetByID(c echo.Context) error {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	ctx := c.Request().Context()
	adminData, err := h.adminAuthUseCase.GetAdminByID(ctx, id)
	if err != nil || adminData == nil {
		return c.JSON(http.StatusNotFound, echo.Map{"error": "admin not found"})
	}
	return c.JSON(http.StatusOK, adminData)
}

// Create 管理者新規登録
// @Summary      管理者新規登録
// @Description  新しい管理者または一般権限のスタッフアカウントを作成します。
// @Tags         Admin-Management
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body   AdminSaveRequest true "管理者登録情報"
// @Success      201  {object}  model.Admin
// @Failure      400  {object}  map[string]string "バリデーションエラー"
// @Router       /admin/admins [post]
func (h *AdminAuthHandler) Create(c echo.Context) error {
	var req AdminSaveRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}

	ctx := c.Request().Context()
	adminData, err := h.adminAuthUseCase.CreateAdmin(ctx, req.Name, req.Email, req.Password, req.RoleType)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusCreated, adminData)
}

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
// @Failure      400  {object}  map[string]string "エラー"
// @Router       /admin/admins/{id} [put]
func (h *AdminAuthHandler) Update(c echo.Context) error {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req AdminSaveRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}

	ctx := c.Request().Context()
	adminData, err := h.adminAuthUseCase.UpdateAdmin(ctx, id, req.Name, req.Email, req.Password, req.RoleType)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, adminData)
}

// Delete 管理者削除
// @Summary      管理者削除
// @Description  指定されたIDの管理者アカウントを削除します。
// @Tags         Admin-Management
// @Security     BearerAuth
// @Param        id   path      int  true  "管理者ID"
// @Success      24   "No Content"
// @Failure      400  {object}  map[string]string "エラー"
// @Router       /admin/admins/{id} [delete]
func (h *AdminAuthHandler) Delete(c echo.Context) error {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	ctx := c.Request().Context()
	if err := h.adminAuthUseCase.DeleteAdmin(ctx, id); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}
	return c.NoContent(http.StatusNoContent)
}
