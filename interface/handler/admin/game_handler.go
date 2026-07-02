package admin

import (
	"net/http"
	"strconv"
	"yutagame-backend/application/usecase/admin"
	"yutagame-backend/interface/handler"

	"yutagame-backend/domain/model"

	"github.com/labstack/echo/v4"
)

type GameHandler struct {
	gameUseCase *admin.GameUseCase
}

func NewGameHandler(gameUseCase *admin.GameUseCase) *GameHandler {
	return &GameHandler{gameUseCase: gameUseCase}
}

type GameCreateRequest struct {
	Name            string  `json:"name"`
	Kana            string  `json:"kana"`
	Overview        string  `json:"overview"`
	Code            string  `json:"code"`
	ManufacturerID  int64   `json:"manufacturerId"`
	MachineID       int64   `json:"machineId"`
	GenreID         int64   `json:"genreId"`
	SubGenre        string  `json:"subGenre"`
	CatchCopy       string  `json:"catchCopy"`
	SubCatch        string  `json:"subCatch"`
	ListPrice       int32   `json:"listPrice"`
	ReleaseDate     string  `json:"releaseDate"` // フロントからは文字列("YYYY-MM-DD")で受け取る
	OfficialSiteURL string  `json:"officialSiteUrl"`
	YouTubeURL      string  `json:"youtubeUrl"`
	IsPlay          bool    `json:"isPlay"`
	IsClear         bool    `json:"isClear"`
	IsFavourite     bool    `json:"isFavourite"`
	KeywordIDs      []int64 `json:"keywordIds"` // 💡 紐付けるキーワードのID配列
}

// Search ゲーム検索・一覧取得
// @Summary      ゲーム検索・一覧取得
// @Description  管理画面用のゲーム一覧を取得します（タイトル部分一致検索など）。
// @Tags         Games
// @Produce      json
// @Security     BearerAuth
// @Param        title  query     string  false  "ゲームタイトル（部分一致）"
// @Success      200    {array}   model.Game
// @Router       /admin/games [get]
func (h *GameHandler) Search(c echo.Context) error {
	// クエリパラメータの取得（未指定やパース失敗時は0にする）
	machineID, _ := strconv.ParseInt(c.QueryParam("machineId"), 10, 64)
	manufacturerID, _ := strconv.ParseInt(c.QueryParam("manufacturerId"), 10, 64)
	keywordID, _ := strconv.ParseInt(c.QueryParam("keywordId"), 10, 64)

	ctx := c.Request().Context()
	games, err := h.gameUseCase.SearchGames(ctx, machineID, manufacturerID, keywordID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, handler.ErrorResponse{
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusOK, games)
}

// GetByID ゲーム詳細取得
// @Summary      ゲーム詳細取得
// @Description  指定されたIDのゲーム詳細情報を取得します。
// @Tags         Games
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "ゲームID"
// @Success      200  {object}  model.Game
// @Failure      404  {object}  map[string]string
// @Router       /admin/games/{id} [get]
func (h *GameHandler) GetByID(c echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{
			Message: "不正なID形式です。整数値を指定してください。",
		})
	}

	ctx := c.Request().Context()
	game, err := h.gameUseCase.GetGameByID(ctx, id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, handler.ErrorResponse{
			Message: err.Error(),
		})
	}
	if game == nil {
		return c.JSON(http.StatusNotFound, handler.ErrorResponse{
			Message: "指定されたIDのゲーム情報が見つかりませんでした。",
		})
	}

	return c.JSON(http.StatusOK, game)
}

// Create ゲーム新規登録
// @Summary      ゲーム新規登録
// @Description  新しいパッケージゲームの情報をデータベースに登録します。
// @Tags         Games
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body   handler.GameCreateRequest true "ゲーム登録情報"
// @Success      201  {object}  model.Game
// @Router       /admin/games [post]
func (h *GameHandler) Create(c echo.Context) error {
	var g model.Game
	if err := c.Bind(&g); err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{
			Message: err.Error(),
		})
	}

	ctx := c.Request().Context()
	if err := h.gameUseCase.CreateGame(ctx, &g); err != nil {
		return c.JSON(http.StatusInternalServerError, handler.ErrorResponse{
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusCreated, g)
}
