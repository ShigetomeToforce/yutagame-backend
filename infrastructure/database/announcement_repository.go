package database

import (
	"context"
	"errors"
	"yutagame-backend/domain/model"

	"gorm.io/gorm"
)

type AnnouncementRepository struct {
	db *gorm.DB
}

func NewAnnouncementRepository(db *gorm.DB) *AnnouncementRepository {
	return &AnnouncementRepository{db: db}
}

func (r *AnnouncementRepository) Create(ctx context.Context, announcement *model.Announcement) error {
	return r.db.WithContext(ctx).Create(announcement).Error
}

func (r *AnnouncementRepository) FindByID(ctx context.Context, id int64) (*model.Announcement, error) {
	var announcement model.Announcement
	err := r.db.WithContext(ctx).First(&announcement, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &announcement, err
}

func (r *AnnouncementRepository) FindPublishedByID(ctx context.Context, id int64) (*model.Announcement, error) {
	var announcement model.Announcement
	err := r.db.WithContext(ctx).
		Where("id = ? AND status = ?", id, "PUBLISHED").
		First(&announcement).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &announcement, err
}

func (r *AnnouncementRepository) FindAll(ctx context.Context) ([]model.Announcement, error) {
	var announcements []model.Announcement
	err := r.db.WithContext(ctx).Order("COALESCE(published_at, created_at) desc, id desc").Find(&announcements).Error
	return announcements, err
}

func (r *AnnouncementRepository) FindPublishedAll(ctx context.Context) ([]model.Announcement, error) {
	var announcements []model.Announcement
	err := r.db.WithContext(ctx).
		Where("status = ?", "PUBLISHED").
		Order("COALESCE(published_at, created_at) desc, id desc").
		Find(&announcements).Error
	return announcements, err
}

func (r *AnnouncementRepository) FindAllWithPagination(
	ctx context.Context,
	limit, offset int,
	whereQueries ...func(*gorm.DB) *gorm.DB,
) ([]model.Announcement, error) {
	return ExecuteFindWithPagination[model.Announcement](ctx, r.db, limit, offset, "COALESCE(published_at, created_at) desc, id desc", nil, whereQueries...)
}

func (r *AnnouncementRepository) CountAll(ctx context.Context, whereQueries ...func(*gorm.DB) *gorm.DB) (int64, error) {
	return ExecuteCount[model.Announcement](ctx, r.db, whereQueries...)
}

func (r *AnnouncementRepository) Update(ctx context.Context, announcement *model.Announcement) error {
	return r.db.WithContext(ctx).Save(announcement).Error
}

func (r *AnnouncementRepository) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&model.Announcement{}, id).Error
}

func (r *AnnouncementRepository) FindPublishedIDs(ctx context.Context) ([]int64, error) {
	var ids []int64
	err := r.db.WithContext(ctx).
		Model(&model.Announcement{}).
		Where("status = ?", "PUBLISHED").
		Order("COALESCE(published_at, created_at) desc, id desc").
		Pluck("id", &ids).Error
	return ids, err
}
