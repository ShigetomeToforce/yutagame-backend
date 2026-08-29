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

// GameSaveRequest ゲーム情報保存時にクライアントから送信されるJSONリクエスト
type GameSaveRequest struct {
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

// GameHandler ゲームに関連するHTTPリクエストの受付とレスポンスの制御を担当するハンドラー
type GameHandler struct {
	gameUseCase *admin.GameUseCase
}

// NewGameHandler GameHandlerの新しいインスタンスを生成するコンストラクタ
func NewGameHandler(gameUseCase *admin.GameUseCase) *GameHandler {
	return &GameHandler{gameUseCase: gameUseCase}
}

// =========================================================================
// 🛠️ Game Management CRUD (ゲーム管理エンドポイント) - ガードあり
// =========================================================================

// -------------------------------------------------------------------------
// C: Create (作成)
// -------------------------------------------------------------------------

// Create ゲーム新規登録
// @Summary      ゲーム新規登録
// @Description  新しいゲームの情報をデータベースに登録します。
// @Tags         Games
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body   GameSaveRequest true "ゲーム登録情報"
// @Success      201  {object}  model.Game
// @Router       /admin/games [post]
func (h *GameHandler) Create(c echo.Context) error {
	var req GameSaveRequest
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

	game := model.Game{
		Name:            req.Name,
		Kana:            req.Kana,
		Overview:        req.Overview,
		Code:            req.Code,
		ManufacturerID:  req.ManufacturerID,
		MachineID:       req.MachineID,
		GenreID:         req.GenreID,
		SubGenre:        req.SubGenre,
		CatchCopy:       req.CatchCopy,
		SubCatch:        req.SubCatch,
		ListPrice:       req.ListPrice,
		ReleaseDate:     releaseDate,
		OfficialSiteURL: req.OfficialSiteURL,
		YouTubeURL:      req.YouTubeURL,
		IsPlay:          req.IsPlay,
		IsClear:         req.IsClear,
		IsFavourite:     req.IsFavourite,
	}

	ctx := c.Request().Context()
	if err := h.gameUseCase.CreateGame(ctx, &game, req.KeywordIDs); err != nil {
		return c.JSON(http.StatusInternalServerError, handler.ErrorResponse{
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusCreated, game)
}

// -------------------------------------------------------------------------
// R: Read (取得)
// -------------------------------------------------------------------------

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
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	ctx := c.Request().Context()

	game, err := h.gameUseCase.GetGameByID(ctx, id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{
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

// GetAll ゲーム一覧取得
// @Summary      ゲーム一覧取得
// @Description  条件に従いゲーム情報の一覧を取得します。
// @Tags         Games
// @Produce      json
// @Security     BearerAuth
// @Param        page  			 query   int  	false "ページ番号 (指定するとページングモード)"
// @Param        limit 			 query   int  	false "表示件数 (10, 30, 50)"
// @Param        q     			 query   string false "自由入力のテキスト検索（ゲーム名）"
// @Param        manufacturerIDs query   []int  false "メーカーIDの複数指定"
// @Param        machineIDs      query   []int  false "機種IDの複数指定"
// @Param        genreIDs        query   []int  false "ジャンルIDの複数指定"
// @Param        keywordIDs      query   []int  false "キーワードIDの複数指定"
// @Success      200   {array}   model.Game "page未指定時"
// @Success      200   {object}  handler.PaginatedResponse[model.Game] "page指定時"
// @Router       /admin/games [get]
func (h *GameHandler) GetAll(c echo.Context) error {
	// 💡 共通のジェネリクス関数に全件・ページングの各ユースケース関数を渡して処理を委ねる
	return handler.HandleListOrPagination(
		c,
		h.gameUseCase.GetAllGames,
		h.gameUseCase.GetGamesWithPagination,
		func(c echo.Context) admin.GameListFilter {
			return admin.GameListFilter{
				SearchWord:      c.QueryParam("q"),
				ManufacturerIDs: handler.ParseInt64Array(c, "manufacturerIDs"),
				MachineIDs:      handler.ParseInt64Array(c, "machineIDs"),
				GenreIDs:        handler.ParseInt64Array(c, "genreIDs"),
				KeywordIDs:      handler.ParseInt64Array(c, "keywordIDs"),
				IsPlay:          handler.ParseOptionalBool(c, "isPlay"),
				IsClear:         handler.ParseOptionalBool(c, "isClear"),
				IsFavourite:     handler.ParseOptionalBool(c, "isFavourite"),
			}
		},
	)
}

// -------------------------------------------------------------------------
// U: Update (更新)
// -------------------------------------------------------------------------

// Update ゲーム情報更新
// @Summary      ゲーム情報更新
// @Description  指定されたIDのゲーム情報更新します。
// @Tags         Games
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id      path   int  true  "ゲームID"
// @Param        request body   GameSaveRequest true "ゲーム更新情報"
// @Success      200  {object}  model.Game
// @Failure      400  {object}  handler.ErrorResponse "エラー"
// @Router       /admin/games/{id} [put]
func (h *GameHandler) Update(c echo.Context) error {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req GameSaveRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{Message: err.Error()})
	}

	releaseDate, err := time.Parse("2006-01-02", req.ReleaseDate)
	if err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{
			Message: "releaseDate の形式が不正です。YYYY-MM-DD を指定してください。",
		})
	}

	game := model.Game{
		ID:              id,
		Name:            req.Name,
		Kana:            req.Kana,
		Overview:        req.Overview,
		Code:            req.Code,
		ManufacturerID:  req.ManufacturerID,
		MachineID:       req.MachineID,
		GenreID:         req.GenreID,
		SubGenre:        req.SubGenre,
		CatchCopy:       req.CatchCopy,
		SubCatch:        req.SubCatch,
		ListPrice:       req.ListPrice,
		ReleaseDate:     releaseDate,
		OfficialSiteURL: req.OfficialSiteURL,
		YouTubeURL:      req.YouTubeURL,
		IsPlay:          req.IsPlay,
		IsClear:         req.IsClear,
		IsFavourite:     req.IsFavourite,
	}

	ctx := c.Request().Context()
	if err := h.gameUseCase.UpdateGame(ctx, &game, req.KeywordIDs); err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusOK, game)
}

// -------------------------------------------------------------------------
// D: Delete (削除)
// -------------------------------------------------------------------------

// Delete ゲーム情報削除
// @Summary      ゲーム情報削除
// @Description  指定されたIDのゲーム情報を削除します。
// @Tags         Games
// @Security     BearerAuth
// @Param        id   path      int  true  "ゲームID"
// @Success      204  "No Content"
// @Failure      400  {object}  handler.ErrorResponse "エラー"
// @Router       /admin/games/{id} [delete]
func (h *GameHandler) Delete(c echo.Context) error {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	ctx := c.Request().Context()

	if err := h.gameUseCase.DeleteGame(ctx, id); err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{Message: err.Error()})
	}
	return c.NoContent(http.StatusNoContent)
}
