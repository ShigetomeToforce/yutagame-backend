package admin

import (
	"context"
	"errors"
	"regexp"
	"strconv"
	"strings"
	"time"
	"yutagame-backend/application/usecase"
	"yutagame-backend/domain/model"
	"yutagame-backend/infrastructure/database"

	"gorm.io/gorm"
)

type AnnouncementListFilter struct {
	SearchWord string
	Status     string
}

type AnnouncementUseCase struct {
	announcementRepo *database.AnnouncementRepository
	accessLogRepo    *database.ContentAccessLogRepository
}

func NewAnnouncementUseCase(announcementRepo *database.AnnouncementRepository, accessLogRepo *database.ContentAccessLogRepository) *AnnouncementUseCase {
	return &AnnouncementUseCase{announcementRepo: announcementRepo, accessLogRepo: accessLogRepo}
}

func (u *AnnouncementUseCase) attachAccessCounts(ctx context.Context, items []model.Announcement) []model.Announcement {
	if len(items) == 0 || u.accessLogRepo == nil {
		return items
	}
	keys := make([]string, 0, len(items))
	for _, item := range items {
		keys = append(keys, strconv.FormatInt(item.ID, 10))
	}
	counts, err := u.accessLogRepo.CountByContentKeys(ctx, "announcement", keys)
	if err != nil {
		return items
	}
	for i := range items {
		items[i].AccessCount = counts[strconv.FormatInt(items[i].ID, 10)]
	}
	return items
}

func trimExcerpt(html string) string {
	re := regexp.MustCompile(`<[^>]+>`)
	text := strings.TrimSpace(re.ReplaceAllString(html, " "))
	runes := []rune(text)
	if len(runes) <= 140 {
		return text
	}
	return string(runes[:140])
}

func (u *AnnouncementUseCase) CreateAnnouncement(ctx context.Context, title, excerpt, bodyHTML, status string, publishStartAt, publishEndAt *time.Time) (*model.Announcement, error) {
	title = strings.TrimSpace(title)
	bodyHTML = strings.TrimSpace(bodyHTML)
	status = strings.TrimSpace(status)
	if title == "" || bodyHTML == "" {
		return nil, errors.New("title and bodyHtml are required")
	}
	if excerpt = strings.TrimSpace(excerpt); excerpt == "" {
		excerpt = trimExcerpt(bodyHTML)
	}
	if status == "" {
		status = "DRAFT"
	}
	if publishStartAt != nil && publishEndAt != nil && publishEndAt.Before(*publishStartAt) {
		return nil, errors.New("publishEndAt must be after publishStartAt")
	}
	announcement := &model.Announcement{
		Title:          title,
		Excerpt:        excerpt,
		BodyHTML:       bodyHTML,
		Status:         status,
		PublishStartAt: publishStartAt,
		PublishEndAt:   publishEndAt,
	}
	if status == "PUBLISHED" {
		now := time.Now()
		announcement.PublishedAt = &now
	}
	if err := u.announcementRepo.Create(ctx, announcement); err != nil {
		return nil, err
	}
	return announcement, nil
}

func (u *AnnouncementUseCase) GetAnnouncementByID(ctx context.Context, id int64) (*model.Announcement, error) {
	return u.announcementRepo.FindByID(ctx, id)
}

func (u *AnnouncementUseCase) GetAllAnnouncements(ctx context.Context) ([]model.Announcement, error) {
	items, err := u.announcementRepo.FindAll(ctx)
	return u.attachAccessCounts(ctx, items), err
}

func (u *AnnouncementUseCase) GetAnnouncementsWithPagination(
	ctx context.Context,
	page, limit int,
	filter AnnouncementListFilter,
) ([]model.Announcement, int64, int, error) {
	var whereQuery func(*gorm.DB) *gorm.DB
	searchWord := strings.TrimSpace(filter.SearchWord)
	status := strings.TrimSpace(filter.Status)
	if searchWord != "" || status != "" {
		whereQuery = func(db *gorm.DB) *gorm.DB {
			if searchWord != "" {
				likeQuery := "%" + searchWord + "%"
				db = db.Where("title LIKE ? OR excerpt LIKE ?", likeQuery, likeQuery)
			}
			if status != "" {
				db = db.Where("status = ?", status)
			}
			return db
		}
	}
	items, totalCount, totalPages, err := usecase.ExecutePaginatedSearch(ctx, page, limit, whereQuery, u.announcementRepo.CountAll, u.announcementRepo.FindAllWithPagination)
	return u.attachAccessCounts(ctx, items), totalCount, totalPages, err
}

func (u *AnnouncementUseCase) UpdateAnnouncement(ctx context.Context, id int64, title, excerpt, bodyHTML, status string, publishStartAt, publishEndAt *time.Time) (*model.Announcement, error) {
	announcement, err := u.announcementRepo.FindByID(ctx, id)
	if err != nil || announcement == nil {
		return nil, errors.New("announcement not found")
	}
	title = strings.TrimSpace(title)
	bodyHTML = strings.TrimSpace(bodyHTML)
	status = strings.TrimSpace(status)
	if title == "" || bodyHTML == "" {
		return nil, errors.New("title and bodyHtml are required")
	}
	if excerpt = strings.TrimSpace(excerpt); excerpt == "" {
		excerpt = trimExcerpt(bodyHTML)
	}
	if publishStartAt != nil && publishEndAt != nil && publishEndAt.Before(*publishStartAt) {
		return nil, errors.New("publishEndAt must be after publishStartAt")
	}
	announcement.Title = title
	announcement.Excerpt = excerpt
	announcement.BodyHTML = bodyHTML
	announcement.Status = status
	announcement.PublishStartAt = publishStartAt
	announcement.PublishEndAt = publishEndAt
	if status == "PUBLISHED" && announcement.PublishedAt == nil {
		now := time.Now()
		announcement.PublishedAt = &now
	}
	if err := u.announcementRepo.Update(ctx, announcement); err != nil {
		return nil, err
	}
	return announcement, nil
}

func (u *AnnouncementUseCase) GetPublishedAnnouncementsForOrdering(ctx context.Context) ([]model.Announcement, error) {
	return u.announcementRepo.FindPublishedForOrdering(ctx)
}

func (u *AnnouncementUseCase) UpdateAnnouncementOrder(ctx context.Context, ids []int64) error {
	return u.announcementRepo.UpdateOrder(ctx, ids)
}

func (u *AnnouncementUseCase) DeleteAnnouncement(ctx context.Context, id int64) error {
	return u.announcementRepo.Delete(ctx, id)
}
