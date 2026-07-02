package admin

import (
	"context"
	"yutagame-backend/domain/model"
	"yutagame-backend/infrastructure/database"
)

type KeywordUseCase struct {
	keywordRepo *database.KeywordRepository
}

func NewKeywordUseCase(keywordRepo *database.KeywordRepository) *KeywordUseCase {
	return &KeywordUseCase{keywordRepo: keywordRepo}
}

func (u *KeywordUseCase) GetAllKeywords(ctx context.Context) ([]model.Keyword, error) {
	return u.keywordRepo.FindAll(ctx)
}

func (u *KeywordUseCase) GetKeywordByID(ctx context.Context, id int64) (*model.Keyword, error) {
	return u.keywordRepo.FindByID(ctx, id)
}
