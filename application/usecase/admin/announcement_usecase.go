package admin

import (
	"context"
	"errors"
	"regexp"
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
}

func NewAnnouncementUseCase(announcementRepo *database.AnnouncementRepository) *AnnouncementUseCase {
	return &AnnouncementUseCase{announcementRepo: announcementRepo}
}

func trimExcerpt(html string) string {
	re := regexp.MustCompile(`<[^>]+>`)
	text := strings.TrimSpace(re.ReplaceAllString(html, " "))
	if len(text) <= 140 {
		return text
	}
	return text[:140]
}

func (u *AnnouncementUseCase) CreateAnnouncement(ctx context.Context, title, excerpt, bodyHTML, status string) (*model.Announcement, error) {
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
	announcement := &model.Announcement{
		Title:    title,
		Excerpt:  excerpt,
		BodyHTML: bodyHTML,
		Status:   status,
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
	return u.announcementRepo.FindAll(ctx)
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
	return usecase.ExecutePaginatedSearch(ctx, page, limit, whereQuery, u.announcementRepo.CountAll, u.announcementRepo.FindAllWithPagination)
}

func (u *AnnouncementUseCase) UpdateAnnouncement(ctx context.Context, id int64, title, excerpt, bodyHTML, status string) (*model.Announcement, error) {
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
	announcement.Title = title
	announcement.Excerpt = excerpt
	announcement.BodyHTML = bodyHTML
	announcement.Status = status
	if status == "PUBLISHED" && announcement.PublishedAt == nil {
		now := time.Now()
		announcement.PublishedAt = &now
	}
	if status != "PUBLISHED" {
		announcement.PublishedAt = announcement.PublishedAt
	}
	if err := u.announcementRepo.Update(ctx, announcement); err != nil {
		return nil, err
	}
	return announcement, nil
}

func (u *AnnouncementUseCase) DeleteAnnouncement(ctx context.Context, id int64) error {
	return u.announcementRepo.Delete(ctx, id)
}
