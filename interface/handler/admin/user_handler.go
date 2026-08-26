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

// UserSaveRequest ユーザーの新規登録および情報更新時に共通で利用するリクエストデータ
type UserSaveRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"` // 更新処理時、空文字の場合は「変更なし」として取り扱われます
}

// UserHandler ユーザーに関連するHTTPリクエストの受付とレスポンスの制御を担当するハンドラー
type UserHandler struct {
	userUseCase *admin.UserUseCase
}

// NewUserHandler UserHandlerの新しいインスタンスを生成するコンストラクタ
func NewUserHandler(userUseCase *admin.UserUseCase) *UserHandler {
	return &UserHandler{userUseCase: userUseCase}
}

// =========================================================================
// 🛠️ Admin Management CRUD (管理者管理エンドポイント) - ガードあり
// =========================================================================

// -------------------------------------------------------------------------
// C: Create (作成)
// -------------------------------------------------------------------------

// Create ユーザー新規登録
// @Summary      ユーザー新規登録
// @Description  新しいユーザーアカウントを作成します。
// @Tags         Users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body   UserSaveRequest true "ユーザー登録情報"
// @Success      201  {object}  model.User
// @Failure      400  {object}  handler.ErrorResponse "バリデーション・重複エラー"
// @Router       /admin/users [post]
func (h *UserHandler) Create(c echo.Context) error {
	var req UserSaveRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{Message: err.Error()})
	}

	ctx := c.Request().Context()
	user, err := h.userUseCase.CreateUser(ctx, req.Name, req.Email, req.Password)
	if err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{Message: err.Error()})
	}
	return c.JSON(http.StatusCreated, user)
}

// -------------------------------------------------------------------------
// R: Read (取得)
// -------------------------------------------------------------------------

// GetByID ユーザー詳細取得
// @Summary      ユーザー詳細取得
// @Description  指定されたIDのユーザー情報を取得します。
// @Tags         Users
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "ユーザーID"
// @Success      200  {object}  model.User
// @Failure      404  {object}  handler.ErrorResponse "未検出エラー"
// @Router       /admin/users/{id} [get]
func (h *UserHandler) GetByID(c echo.Context) error {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	ctx := c.Request().Context()

	admin, err := h.userUseCase.GetUserByID(ctx, id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{
			Message: err.Error(),
		})
	}
	if admin == nil {
		return c.JSON(http.StatusNotFound, handler.ErrorResponse{
			Message: "指定されたIDのユーザー情報が見つかりませんでした。",
		})
	}

	return c.JSON(http.StatusOK, admin)
}

// GetAll ユーザー一覧取得
// @Summary      ユーザー一覧取得
// @Description  条件に従いユーザー情報の一覧を取得します。
// @Tags         Users
// @Produce      json
// @Security     BearerAuth
// @Param        page  query     int  false  "ページ番号 (指定するとページングモード)"
// @Param        limit query     int  false  "表示件数 (10, 30, 50)"
// @Param        q     query     string false "自由入力のテキスト検索（管理者名）"
// @Success      200   {array}   model.User "page未指定時"
// @Success      200   {object}  handler.PaginatedResponse[model.User] "page指定時"
// @Router       /admin/users [get]
func (h *UserHandler) GetAll(c echo.Context) error {
	// 💡 共通のジェネリクス関数に全件・ページングの各ユースケース関数を渡して処理を委ねる
	return handler.HandleListOrPagination(
		c,
		h.userUseCase.GetAllUsers,
		h.userUseCase.GetUsersWithPagination,
		func(c echo.Context) admin.UserListFilter {
			return admin.UserListFilter{
				SearchWord: c.QueryParam("q"),
			}
		},
	)
}

// -------------------------------------------------------------------------
// U: Update (更新)
// -------------------------------------------------------------------------

// Update ユーザー情報更新
// @Summary      ユーザー情報更新
// @Description  指定されたIDの管理者の名前、メール、パスワードを更新します。
// @Tags         Users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id      path   int  true  "ユーザーID"
// @Param        request body   UserSaveRequest true "ユーザー更新情報"
// @Success      200  {object}  model.User
// @Failure      400  {object}  handler.ErrorResponse "エラー"
// @Router       /admin/users/{id} [put]
func (h *UserHandler) Update(c echo.Context) error {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req UserSaveRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{Message: err.Error()})
	}

	ctx := c.Request().Context()
	adminData, err := h.userUseCase.UpdateUser(ctx, id, req.Name, req.Email, req.Password)
	if err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{Message: err.Error()})
	}
	return c.JSON(http.StatusOK, adminData)
}

// -------------------------------------------------------------------------
// D: Delete (削除)
// -------------------------------------------------------------------------

// Delete ユーザー削除
// @Summary      ユーザー削除
// @Description  指定されたIDのユーザー情報を削除します。
// @Tags         Users
// @Security     BearerAuth
// @Param        id   path      int  true  "管理者ID"
// @Success      204  "No Content"
// @Failure      400  {object}  handler.ErrorResponse "エラー"
// @Router       /admin/users/{id} [delete]
func (h *UserHandler) Delete(c echo.Context) error {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	ctx := c.Request().Context()

	if err := h.userUseCase.DeleteUser(ctx, id); err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{Message: err.Error()})
	}
	return c.NoContent(http.StatusNoContent)
}
