package admin

import (
	"context"
	"errors"
	"strings"
	"yutagame-backend/domain/model"
	"yutagame-backend/infrastructure/database"
)

type PurchaseCandidateUseCase struct {
	repo *database.PurchaseCandidateRepository
}

func NewPurchaseCandidateUseCase(repo *database.PurchaseCandidateRepository) *PurchaseCandidateUseCase {
	return &PurchaseCandidateUseCase{repo: repo}
}

func (u *PurchaseCandidateUseCase) validate(ctx context.Context, item *model.PurchaseCandidate, excludeID int64) error {
	item.Name = strings.TrimSpace(item.Name)
	item.Kana = strings.TrimSpace(item.Kana)
	item.Code = strings.TrimSpace(item.Code)
	item.OfficialSiteURL = strings.TrimSpace(item.OfficialSiteURL)
	item.YouTubeURL = strings.TrimSpace(item.YouTubeURL)
	item.ReleaseDateText = strings.TrimSpace(item.ReleaseDateText)
	if item.Name == "" || item.Kana == "" || item.Code == "" || item.ManufacturerID <= 0 || item.MachineID <= 0 {
		return errors.New("name, kana, code, manufacturerId and machineId are required")
	}
	if err := validateDuplicateCode(ctx, item.Code, excludeID, u.repo.FindByCode); err != nil {
		if errors.Is(err, ErrCodeAlreadyExists) {
			return errors.New("指定されたコードは既に使用されています。")
		}
		return err
	}
	return nil
}

func (u *PurchaseCandidateUseCase) Create(ctx context.Context, item *model.PurchaseCandidate) (*model.PurchaseCandidate, error) {
	item.IsPurchased = false
	if err := u.validate(ctx, item, 0); err != nil {
		return nil, err
	}
	if item.DisplayOrder < 1 {
		item.DisplayOrder = 1
	}
	if err := u.repo.Create(ctx, item); err != nil {
		return nil, err
	}
	return u.repo.FindByID(ctx, item.ID)
}

func (u *PurchaseCandidateUseCase) GetByID(ctx context.Context, id int64) (*model.PurchaseCandidate, error) {
	return u.repo.FindByID(ctx, id)
}

func (u *PurchaseCandidateUseCase) GetByCode(ctx context.Context, code string) (*model.PurchaseCandidate, error) {
	return u.repo.FindByCode(ctx, strings.TrimSpace(code))
}

func (u *PurchaseCandidateUseCase) GetAllByPurchased(ctx context.Context, isPurchased bool) ([]model.PurchaseCandidate, error) {
	items, err := u.repo.FindAllByPurchased(ctx, isPurchased)
	if items == nil {
		items = []model.PurchaseCandidate{}
	}
	return items, err
}

func (u *PurchaseCandidateUseCase) Update(ctx context.Context, id int64, item *model.PurchaseCandidate) (*model.PurchaseCandidate, error) {
	existing, err := u.repo.FindByID(ctx, id)
	if err != nil || existing == nil {
		return nil, errors.New("purchase candidate not found")
	}
	item.ID = id
	if err := u.validate(ctx, item, id); err != nil {
		return nil, err
	}
	if err := u.repo.Update(ctx, item); err != nil {
		return nil, err
	}
	return u.repo.FindByID(ctx, id)
}

func (u *PurchaseCandidateUseCase) UpdateImage(ctx context.Context, id int64, imageKey string) (*model.PurchaseCandidate, error) {
	return u.repo.UpdateImage(ctx, id, &imageKey)
}

func (u *PurchaseCandidateUseCase) DeleteImage(ctx context.Context, id int64) (*model.PurchaseCandidate, error) {
	return u.repo.UpdateImage(ctx, id, nil)
}

func (u *PurchaseCandidateUseCase) UpdateOrder(ctx context.Context, isPurchased bool, ids []int64) error {
	return u.repo.UpdateOrder(ctx, isPurchased, ids)
}

func (u *PurchaseCandidateUseCase) Delete(ctx context.Context, id int64) error {
	return u.repo.Delete(ctx, id)
}
