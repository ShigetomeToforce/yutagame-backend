package handler

import (
	"net/http"
	"yutagame-backend/application/usecase"

	"github.com/labstack/echo/v4"
)

type KeywordHandler struct {
	keywordUseCase *usecase.KeywordUseCase
}

func NewKeywordHandler(keywordUseCase *usecase.KeywordUseCase) *KeywordHandler {
	return &KeywordHandler{keywordUseCase: keywordUseCase}
}

// GetAll はキーワード一覧を返すAPIエンドポイントです
func (h *KeywordHandler) GetAll(c echo.Context) error {
	ctx := c.Request().Context()
	keywords, err := h.keywordUseCase.GetAllKeywords(ctx)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, keywords)
}
