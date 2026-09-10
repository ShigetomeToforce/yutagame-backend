package admin

import (
	"context"
	"errors"
	"strings"
	"yutagame-backend/application/usecase"
	"yutagame-backend/domain/model"
	"yutagame-backend/infrastructure/database"

	"gorm.io/gorm"
)

type ContactInquiryListFilter struct {
	SearchWord string
	Status     string
}

type ContactInquiryUseCase struct {
	contactRepo *database.ContactInquiryRepository
}

func NewContactInquiryUseCase(contactRepo *database.ContactInquiryRepository) *ContactInquiryUseCase {
	return &ContactInquiryUseCase{contactRepo: contactRepo}
}

func (u *ContactInquiryUseCase) CreateContactInquiry(ctx context.Context, name, email, subject, message string) (*model.ContactInquiry, error) {
	name = strings.TrimSpace(name)
	email = strings.TrimSpace(email)
	subject = strings.TrimSpace(subject)
	message = strings.TrimSpace(message)
	if subject == "" || message == "" {
		return nil, errors.New("subject and message are required")
	}
	inquiry := &model.ContactInquiry{
		Name:      name,
		Email:     email,
		Subject:   subject,
		Message:   message,
		Status:    "NEW",
		AdminNote: "",
	}
	if err := u.contactRepo.Create(ctx, inquiry); err != nil {
		return nil, err
	}
	return inquiry, nil
}

func (u *ContactInquiryUseCase) GetContactInquiryByID(ctx context.Context, id int64) (*model.ContactInquiry, error) {
	return u.contactRepo.FindByID(ctx, id)
}

func (u *ContactInquiryUseCase) GetAllContactInquiries(ctx context.Context) ([]model.ContactInquiry, error) {
	return u.contactRepo.FindAll(ctx)
}

func (u *ContactInquiryUseCase) GetContactInquiriesWithPagination(
	ctx context.Context,
	page, limit int,
	filter ContactInquiryListFilter,
) ([]model.ContactInquiry, int64, int, error) {
	var whereQuery func(*gorm.DB) *gorm.DB
	searchWord := strings.TrimSpace(filter.SearchWord)
	status := strings.TrimSpace(filter.Status)
	if searchWord != "" || status != "" {
		whereQuery = func(db *gorm.DB) *gorm.DB {
			if searchWord != "" {
				likeQuery := "%" + searchWord + "%"
				db = db.Where("name LIKE ? OR email LIKE ? OR subject LIKE ? OR message LIKE ?", likeQuery, likeQuery, likeQuery, likeQuery)
			}
			if status != "" {
				db = db.Where("status = ?", status)
			}
			return db
		}
	}
	return usecase.ExecutePaginatedSearch(ctx, page, limit, whereQuery, u.contactRepo.CountAll, u.contactRepo.FindAllWithPagination)
}

func (u *ContactInquiryUseCase) UpdateContactInquiry(ctx context.Context, id int64, status, adminNote string) (*model.ContactInquiry, error) {
	inquiry, err := u.contactRepo.FindByID(ctx, id)
	if err != nil || inquiry == nil {
		return nil, errors.New("contact inquiry not found")
	}
	inquiry.Status = strings.TrimSpace(status)
	inquiry.AdminNote = strings.TrimSpace(adminNote)
	if inquiry.Status == "" {
		inquiry.Status = "NEW"
	}
	if err := u.contactRepo.Update(ctx, inquiry); err != nil {
		return nil, err
	}
	return inquiry, nil
}

func (u *ContactInquiryUseCase) DeleteContactInquiry(ctx context.Context, id int64) error {
	return u.contactRepo.Delete(ctx, id)
}
