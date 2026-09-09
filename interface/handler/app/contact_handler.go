package app

import (
	"net/http"
	usecaseApp "yutagame-backend/application/usecase/app"
	"yutagame-backend/interface/handler"

	"github.com/labstack/echo/v4"
)

type ContactRequest struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	Subject string `json:"subject"`
	Message string `json:"message"`
}

type ContactHandler struct {
	contactUseCase *usecaseApp.ContactPublicUseCase
}

func NewContactHandler(contactUseCase *usecaseApp.ContactPublicUseCase) *ContactHandler {
	return &ContactHandler{contactUseCase: contactUseCase}
}

func (h *ContactHandler) Create(c echo.Context) error {
	var req ContactRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{Message: err.Error()})
	}
	ctx := c.Request().Context()
	item, err := h.contactUseCase.CreateContactInquiry(ctx, req.Name, req.Email, req.Subject, req.Message)
	if err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{Message: err.Error()})
	}
	return c.JSON(http.StatusCreated, item)
}
