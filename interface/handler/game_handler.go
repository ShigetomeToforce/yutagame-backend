package handler

import (
	"net/http"
	"strconv"
	"yutagame-backend/application/usecase"
	"yutagame-backend/domain/model"

	"github.com/labstack/echo/v4"
)

type GameHandler struct {
	gameUseCase *usecase.GameUseCase
}

func NewGameHandler(gameUseCase *usecase.GameUseCase) *GameHandler {
	return &GameHandler{gameUseCase: gameUseCase}
}

// Search はクエリパラメータに応じたゲーム検索・一覧を返すAPIエンドポイントです
func (h *GameHandler) Search(c echo.Context) error {
	// クエリパラメータの取得（未指定やパース失敗時は0にする）
	machineID, _ := strconv.ParseInt(c.QueryParam("machineId"), 10, 64)
	manufacturerID, _ := strconv.ParseInt(c.QueryParam("manufacturerId"), 10, 64)
	keywordID, _ := strconv.ParseInt(c.QueryParam("keywordId"), 10, 64)

	ctx := c.Request().Context()
	games, err := h.gameUseCase.SearchGames(ctx, machineID, manufacturerID, keywordID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, games)
}

// GetByID は指定されたIDのゲーム詳細を返すAPIエンドポイントです
func (h *GameHandler) GetByID(c echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid id format"})
	}

	ctx := c.Request().Context()
	game, err := h.gameUseCase.GetGameByID(ctx, id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}
	if game == nil {
		return c.JSON(http.StatusNotFound, echo.Map{"error": "game not found"})
	}

	return c.JSON(http.StatusOK, game)
}

// Create はゲームを新規登録するAPIエンドポイントです（管理画面用）
func (h *GameHandler) Create(c echo.Context) error {
	var g model.Game
	if err := c.Bind(&g); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}

	ctx := c.Request().Context()
	if err := h.gameUseCase.CreateGame(ctx, &g); err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, g)
}
