package app

import (
	"context"
	"yutagame-backend/domain/model"
	"yutagame-backend/infrastructure/database"
)

type AnnouncementPublicUseCase struct {
	announcementRepo *database.AnnouncementRepository
}

func NewAnnouncementPublicUseCase(announcementRepo *database.AnnouncementRepository) *AnnouncementPublicUseCase {
	return &AnnouncementPublicUseCase{announcementRepo: announcementRepo}
}

func (u *AnnouncementPublicUseCase) GetPublishedAnnouncements(ctx context.Context) ([]model.Announcement, error) {
	return u.announcementRepo.FindPublishedAll(ctx)
}

func (u *AnnouncementPublicUseCase) GetPublishedAnnouncementByID(ctx context.Context, id int64) (*model.Announcement, error) {
	return u.announcementRepo.FindPublishedByID(ctx, id)
}
