package database

import (
	"context"
	"errors"
	"time"
	"yutagame-backend/domain/model"

	"gorm.io/gorm"
)

type ContactInquiryRepository struct {
	db *gorm.DB
}

const contactInquiryOrder = "CASE status WHEN 'NEW' THEN 1 WHEN 'IN_PROGRESS' THEN 2 WHEN 'DONE' THEN 3 ELSE 4 END asc, updated_at desc, id desc"

func NewContactInquiryRepository(db *gorm.DB) *ContactInquiryRepository {
	return &ContactInquiryRepository{db: db}
}

func (r *ContactInquiryRepository) Create(ctx context.Context, inquiry *model.ContactInquiry) error {
	return r.db.WithContext(ctx).Create(inquiry).Error
}

func (r *ContactInquiryRepository) FindByID(ctx context.Context, id int64) (*model.ContactInquiry, error) {
	var inquiry model.ContactInquiry
	err := r.db.WithContext(ctx).First(&inquiry, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &inquiry, err
}

func (r *ContactInquiryRepository) FindAll(ctx context.Context) ([]model.ContactInquiry, error) {
	var inquiries []model.ContactInquiry
	err := r.db.WithContext(ctx).Order(contactInquiryOrder).Find(&inquiries).Error
	return inquiries, err
}

func (r *ContactInquiryRepository) FindAllWithPagination(
	ctx context.Context,
	limit, offset int,
	whereQueries ...func(*gorm.DB) *gorm.DB,
) ([]model.ContactInquiry, error) {
	return ExecuteFindWithPagination[model.ContactInquiry](ctx, r.db, limit, offset, contactInquiryOrder, nil, whereQueries...)
}

func (r *ContactInquiryRepository) CountAll(ctx context.Context, whereQueries ...func(*gorm.DB) *gorm.DB) (int64, error) {
	return ExecuteCount[model.ContactInquiry](ctx, r.db, whereQueries...)
}

func (r *ContactInquiryRepository) CountRange(ctx context.Context, from, to time.Time) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.ContactInquiry{}).
		Where("created_at >= ? AND created_at < ?", from, to).
		Count(&count).Error
	return count, err
}

func (r *ContactInquiryRepository) Update(ctx context.Context, inquiry *model.ContactInquiry) error {
	return r.db.WithContext(ctx).Save(inquiry).Error
}

func (r *ContactInquiryRepository) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&model.ContactInquiry{}, id).Error
}
