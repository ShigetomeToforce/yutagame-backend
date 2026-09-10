package database

import (
	"context"
	"errors"
	"time"
	"yutagame-backend/domain/model"

	"gorm.io/gorm"
)

type AnnouncementRepository struct {
	db *gorm.DB
}

func NewAnnouncementRepository(db *gorm.DB) *AnnouncementRepository {
	return &AnnouncementRepository{db: db}
}

// publishedOrderClause: 手動で並べ替えられた項目（display_order > 0）を先頭にし、残りは公開日時の新しい順に並べる
const publishedOrderClause = "CASE WHEN display_order > 0 THEN 0 ELSE 1 END, display_order ASC, COALESCE(published_at, created_at) desc, id desc"

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
	now := time.Now()
	err := r.db.WithContext(ctx).
		Where("id = ? AND status = ?", id, "PUBLISHED").
		Where("publish_start_at IS NULL OR publish_start_at <= ?", now).
		Where("publish_end_at IS NULL OR publish_end_at >= ?", now).
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
	now := time.Now()
	err := r.db.WithContext(ctx).
		Where("status = ?", "PUBLISHED").
		Where("publish_start_at IS NULL OR publish_start_at <= ?", now).
		Where("publish_end_at IS NULL OR publish_end_at >= ?", now).
		Order(publishedOrderClause).
		Find(&announcements).Error
	return announcements, err
}

// FindPublishedForOrdering は期間に関係なくステータスがPUBLISHEDのものを並べ替え対象として返す
func (r *AnnouncementRepository) FindPublishedForOrdering(ctx context.Context) ([]model.Announcement, error) {
	var announcements []model.Announcement
	err := r.db.WithContext(ctx).
		Where("status = ?", "PUBLISHED").
		Order(publishedOrderClause).
		Find(&announcements).Error
	return announcements, err
}

// UpdateOrder は指定された順番で display_order を 1から振り直す
func (r *AnnouncementRepository) UpdateOrder(ctx context.Context, ids []int64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for i, id := range ids {
			if err := tx.Model(&model.Announcement{}).Where("id = ?", id).Update("display_order", i+1).Error; err != nil {
				return err
			}
		}
		return nil
	})
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
	// map指定のUpdatesはゼロ値(nilのpublish_start_at等)もそのまま反映されるため、構造体経由のSaveより確実
	return r.db.WithContext(ctx).Model(&model.Announcement{}).Where("id = ?", announcement.ID).Updates(map[string]interface{}{
		"title":            announcement.Title,
		"excerpt":          announcement.Excerpt,
		"body_html":        announcement.BodyHTML,
		"status":           announcement.Status,
		"published_at":     announcement.PublishedAt,
		"publish_start_at": announcement.PublishStartAt,
		"publish_end_at":   announcement.PublishEndAt,
	}).Error
}

func (r *AnnouncementRepository) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&model.Announcement{}, id).Error
}

func (r *AnnouncementRepository) FindPublishedIDs(ctx context.Context) ([]int64, error) {
	var ids []int64
	now := time.Now()
	err := r.db.WithContext(ctx).
		Model(&model.Announcement{}).
		Where("status = ?", "PUBLISHED").
		Where("publish_start_at IS NULL OR publish_start_at <= ?", now).
		Where("publish_end_at IS NULL OR publish_end_at >= ?", now).
		Order(publishedOrderClause).
		Pluck("id", &ids).Error
	return ids, err
}
