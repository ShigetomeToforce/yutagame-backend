package app

import (
	"context"
	"yutagame-backend/application/usecase/admin"
	"yutagame-backend/domain/model"
	"yutagame-backend/infrastructure/database"
)

type ContactPublicUseCase struct {
	contactRepo *database.ContactInquiryRepository
}

func NewContactPublicUseCase(contactRepo *database.ContactInquiryRepository) *ContactPublicUseCase {
	return &ContactPublicUseCase{contactRepo: contactRepo}
}

func (u *ContactPublicUseCase) CreateContactInquiry(ctx context.Context, name, email, subject, message string) (*model.ContactInquiry, error) {
	return admin.NewContactInquiryUseCase(u.contactRepo).CreateContactInquiry(ctx, name, email, subject, message)
}
