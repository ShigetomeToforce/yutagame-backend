package admin

import (
	"net/http"
	"strconv"
	usecaseAdmin "yutagame-backend/application/usecase/admin"
	"yutagame-backend/interface/handler"

	"github.com/labstack/echo/v4"
)

type ContactInquirySaveRequest struct {
	Status    string `json:"status"`
	AdminNote string `json:"adminNote"`
}

type ContactInquiryHandler struct {
	contactUseCase *usecaseAdmin.ContactInquiryUseCase
}

func NewContactInquiryHandler(contactUseCase *usecaseAdmin.ContactInquiryUseCase) *ContactInquiryHandler {
	return &ContactInquiryHandler{contactUseCase: contactUseCase}
}

func (h *ContactInquiryHandler) GetAll(c echo.Context) error {
	return handler.HandleListOrPagination(
		c,
		h.contactUseCase.GetAllContactInquiries,
		h.contactUseCase.GetContactInquiriesWithPagination,
		func(c echo.Context) usecaseAdmin.ContactInquiryListFilter {
			return usecaseAdmin.ContactInquiryListFilter{
				SearchWord: c.QueryParam("q"),
				Status:     c.QueryParam("status"),
			}
		},
	)
}

func (h *ContactInquiryHandler) GetByID(c echo.Context) error {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	ctx := c.Request().Context()
	inquiry, err := h.contactUseCase.GetContactInquiryByID(ctx, id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{Message: err.Error()})
	}
	if inquiry == nil {
		return c.JSON(http.StatusNotFound, handler.ErrorResponse{Message: "指定された問い合わせが見つかりませんでした。"})
	}
	return c.JSON(http.StatusOK, inquiry)
}

func (h *ContactInquiryHandler) Update(c echo.Context) error {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req ContactInquirySaveRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{Message: err.Error()})
	}
	ctx := c.Request().Context()
	inquiry, err := h.contactUseCase.UpdateContactInquiry(ctx, id, req.Status, req.AdminNote)
	if err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{Message: err.Error()})
	}
	return c.JSON(http.StatusOK, inquiry)
}

func (h *ContactInquiryHandler) Delete(c echo.Context) error {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	ctx := c.Request().Context()
	if err := h.contactUseCase.DeleteContactInquiry(ctx, id); err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{Message: err.Error()})
	}
	return c.NoContent(http.StatusNoContent)
}
