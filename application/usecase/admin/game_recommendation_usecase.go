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

type GameRecommendationListFilter struct{ SearchWord, Status string }
type GameRecommendationUseCase struct {
	repo *database.GameRecommendationRepository
}

func NewGameRecommendationUseCase(repo *database.GameRecommendationRepository) *GameRecommendationUseCase {
	return &GameRecommendationUseCase{repo: repo}
}

func (u *GameRecommendationUseCase) Create(ctx context.Context, gameName, reason string) (*model.GameRecommendation, error) {
	gameName, reason = strings.TrimSpace(gameName), strings.TrimSpace(reason)
	if gameName == "" || reason == "" {
		return nil, errors.New("gameName and reason are required")
	}
	item := &model.GameRecommendation{GameName: gameName, Reason: reason, Status: "NEW", AdminNote: ""}
	return item, u.repo.Create(ctx, item)
}

func (u *GameRecommendationUseCase) GetByID(ctx context.Context, id int64) (*model.GameRecommendation, error) {
	return u.repo.FindByID(ctx, id)
}
func (u *GameRecommendationUseCase) GetAll(ctx context.Context) ([]model.GameRecommendation, error) {
	return u.repo.FindAll(ctx)
}

func (u *GameRecommendationUseCase) GetWithPagination(ctx context.Context, page, limit int, filter GameRecommendationListFilter) ([]model.GameRecommendation, int64, int, error) {
	searchWord, status := strings.TrimSpace(filter.SearchWord), strings.TrimSpace(filter.Status)
	var whereQuery func(*gorm.DB) *gorm.DB
	if searchWord != "" || status != "" {
		whereQuery = func(db *gorm.DB) *gorm.DB {
			if searchWord != "" {
				like := "%" + searchWord + "%"
				db = db.Where("game_name LIKE ? OR reason LIKE ?", like, like)
			}
			if status != "" {
				db = db.Where("status = ?", status)
			}
			return db
		}
	}
	return usecase.ExecutePaginatedSearch(ctx, page, limit, whereQuery, u.repo.CountAll, u.repo.FindAllWithPagination)
}

func (u *GameRecommendationUseCase) Update(ctx context.Context, id int64, status, adminNote string) (*model.GameRecommendation, error) {
	item, err := u.repo.FindByID(ctx, id)
	if err != nil || item == nil {
		return nil, errors.New("game recommendation not found")
	}
	item.Status, item.AdminNote = strings.TrimSpace(status), strings.TrimSpace(adminNote)
	if item.Status == "" {
		item.Status = "NEW"
	}
	return item, u.repo.Update(ctx, item)
}

func (u *GameRecommendationUseCase) Delete(ctx context.Context, id int64) error {
	return u.repo.Delete(ctx, id)
}
