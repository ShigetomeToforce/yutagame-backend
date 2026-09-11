package app

import (
	"context"
	"log"
	"yutagame-backend/application/usecase/admin"
	"yutagame-backend/domain/model"
	"yutagame-backend/infrastructure/database"
	"yutagame-backend/infrastructure/mail"
)

type ContactPublicUseCase struct {
	contactRepo *database.ContactInquiryRepository
	notifier    mail.ContactNotifier
}

func NewContactPublicUseCase(contactRepo *database.ContactInquiryRepository, notifier mail.ContactNotifier) *ContactPublicUseCase {
	return &ContactPublicUseCase{contactRepo: contactRepo, notifier: notifier}
}

func (u *ContactPublicUseCase) CreateContactInquiry(ctx context.Context, name, email, subject, message string) (*model.ContactInquiry, error) {
	// 管理画面と同じ入力検証・保存ルールを使い、公開側だけ仕様がずれないようにします。
	inquiry, err := admin.NewContactInquiryUseCase(u.contactRepo).CreateContactInquiry(ctx, name, email, subject, message)
	if err != nil {
		return nil, err
	}
	if u.notifier != nil {
		// 保存済みの問い合わせをメール障害で失敗扱いにしないため、通知はベストエフォートです。
		if notifyErr := u.notifier.SendContactNotification(ctx, inquiry); notifyErr != nil {
			log.Printf("contact notification mail failed: %v", notifyErr)
		}
	}
	return inquiry, nil
}
