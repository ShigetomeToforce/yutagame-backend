package admin

import (
	"errors"
	"net/http"
	"strconv"
	"time"
	_ "time/tzdata"
	usecaseAdmin "yutagame-backend/application/usecase/admin"
	"yutagame-backend/interface/handler"

	"github.com/labstack/echo/v4"
)

const publishDateTimeLayout = "2006-01-02T15:04"

// jstLocation: datetime-localはタイムゾーン情報を持たないブラウザのローカル時刻(JST想定)を送ってくるため、
// コンテナのOSタイムゾーン(time.Local。開発環境ではUTCになりがち)に依存せず常にJSTとして解釈する
var jstLocation = func() *time.Location {
	loc, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		return time.FixedZone("JST", 9*60*60)
	}
	return loc
}()

type AnnouncementSaveRequest struct {
	Title          string `json:"title"`
	Excerpt        string `json:"excerpt"`
	BodyHTML       string `json:"bodyHtml"`
	Status         string `json:"status"`
	PublishStartAt string `json:"publishStartAt"`
	PublishEndAt   string `json:"publishEndAt"`
}

// parsePublishDateTime は <input type="datetime-local"> の "2006-01-02T15:04" 形式をJSTとして解釈する
func parsePublishDateTime(value string) (*time.Time, error) {
	if value == "" {
		return nil, nil
	}
	parsed, err := time.ParseInLocation(publishDateTimeLayout, value, jstLocation)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

type AnnouncementHandler struct {
	announcementUseCase *usecaseAdmin.AnnouncementUseCase
}

func NewAnnouncementHandler(announcementUseCase *usecaseAdmin.AnnouncementUseCase) *AnnouncementHandler {
	return &AnnouncementHandler{announcementUseCase: announcementUseCase}
}

func (h *AnnouncementHandler) Create(c echo.Context) error {
	var req AnnouncementSaveRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{Message: err.Error()})
	}
	publishStartAt, err := parsePublishDateTime(req.PublishStartAt)
	if err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{Message: "公開開始日時の形式が正しくありません。"})
	}
	publishEndAt, err := parsePublishDateTime(req.PublishEndAt)
	if err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{Message: "公開終了日時の形式が正しくありません。"})
	}
	ctx := c.Request().Context()
	announcement, err := h.announcementUseCase.CreateAnnouncement(ctx, req.Title, req.Excerpt, req.BodyHTML, req.Status, publishStartAt, publishEndAt)
	if err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{Message: err.Error()})
	}
	return c.JSON(http.StatusCreated, announcement)
}

func (h *AnnouncementHandler) GetByID(c echo.Context) error {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	ctx := c.Request().Context()
	announcement, err := h.announcementUseCase.GetAnnouncementByID(ctx, id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{Message: err.Error()})
	}
	if announcement == nil {
		return c.JSON(http.StatusNotFound, handler.ErrorResponse{Message: "指定されたお知らせが見つかりませんでした。"})
	}
	return c.JSON(http.StatusOK, announcement)
}

func (h *AnnouncementHandler) GetAll(c echo.Context) error {
	return handler.HandleListOrPagination(
		c,
		h.announcementUseCase.GetAllAnnouncements,
		h.announcementUseCase.GetAnnouncementsWithPagination,
		func(c echo.Context) usecaseAdmin.AnnouncementListFilter {
			return usecaseAdmin.AnnouncementListFilter{
				SearchWord: c.QueryParam("q"),
				Status:     c.QueryParam("status"),
			}
		},
	)
}

func (h *AnnouncementHandler) Update(c echo.Context) error {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req AnnouncementSaveRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{Message: err.Error()})
	}
	publishStartAt, err := parsePublishDateTime(req.PublishStartAt)
	if err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{Message: "公開開始日時の形式が正しくありません。"})
	}
	publishEndAt, err := parsePublishDateTime(req.PublishEndAt)
	if err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{Message: "公開終了日時の形式が正しくありません。"})
	}
	ctx := c.Request().Context()
	announcement, err := h.announcementUseCase.UpdateAnnouncement(ctx, id, req.Title, req.Excerpt, req.BodyHTML, req.Status, publishStartAt, publishEndAt)
	if err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{Message: err.Error()})
	}
	return c.JSON(http.StatusOK, announcement)
}

func (h *AnnouncementHandler) Delete(c echo.Context) error {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	ctx := c.Request().Context()
	if err := h.announcementUseCase.DeleteAnnouncement(ctx, id); err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{Message: err.Error()})
	}
	return c.NoContent(http.StatusNoContent)
}

// GetPublishedForOrdering は表示順を並び替えるための公開中お知らせ一覧を返す（期間は考慮しない）
func (h *AnnouncementHandler) GetPublishedForOrdering(c echo.Context) error {
	ctx := c.Request().Context()
	items, err := h.announcementUseCase.GetPublishedAnnouncementsForOrdering(ctx)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, handler.ErrorResponse{Message: err.Error()})
	}
	return c.JSON(http.StatusOK, items)
}

type AnnouncementOrderRequest struct {
	IDs []int64 `json:"ids"`
}

// UpdateOrder は公開中お知らせの表示優先度を並び替える
func (h *AnnouncementHandler) UpdateOrder(c echo.Context) error {
	var req AnnouncementOrderRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{Message: err.Error()})
	}
	ctx := c.Request().Context()
	if err := h.announcementUseCase.UpdateAnnouncementOrder(ctx, req.IDs); err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{Message: err.Error()})
	}
	return c.NoContent(http.StatusNoContent)
}

// UploadContentImage 本文中に挿入する画像をアップロードする（お知らせIDに紐付かない汎用アップロード）
func (h *AnnouncementHandler) UploadContentImage(c echo.Context) error {
	fileHeader, err := c.FormFile("image")
	if err != nil {
		return c.JSON(http.StatusBadRequest, handler.ErrorResponse{Message: "画像ファイルを選択してください。"})
	}

	imageKey, err := saveUploadedImage(fileHeader, "announcements", "content")
	if err != nil {
		status := http.StatusBadRequest
		if !errors.Is(err, errImageRequired) && !errors.Is(err, errImageTooLarge) && !errors.Is(err, errImageTypeDenied) {
			status = http.StatusInternalServerError
		}
		return c.JSON(status, handler.ErrorResponse{Message: err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{"imageKey": imageKey})
}
