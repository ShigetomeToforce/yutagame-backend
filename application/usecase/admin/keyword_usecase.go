package admin

import (
	"context"
	"yutagame-backend/application/usecase"
	"yutagame-backend/domain/model"
	"yutagame-backend/infrastructure/database"

	"gorm.io/gorm"
)

// =========================================================================
// 構造体＆コンストラクタ
// =========================================================================

// KeywordListFilter キーワード検索用構造体
type KeywordListFilter struct {
	SearchWord string
}

// KeywordUseCase キーワード管理のビジネスロジックを担当するユースケース
type KeywordUseCase struct {
	keywordRepo *database.KeywordRepository
}

// NewKeywordUseCase KeywordUseCaseの新しいインスタンスを生成するコンストラクタ
func NewKeywordUseCase(keywordRepo *database.KeywordRepository) *KeywordUseCase {
	return &KeywordUseCase{keywordRepo: keywordRepo}
}

// =========================================================================
// Keyword Management CRUD (キーワード管理ロジック) - ルーティングのガード内側で利用
// =========================================================================

// -------------------------------------------------------------------------
// C: Create (作成)
// -------------------------------------------------------------------------

// CreateKeyword 新しいキーワードを作成する
func (u *KeywordUseCase) CreateKeyword(ctx context.Context, g *model.Keyword) error {
	return u.keywordRepo.Create(ctx, g)
}

// -------------------------------------------------------------------------
// R: Read (取得)
// -------------------------------------------------------------------------

// GetKeywordByID キーワードIDを指定して、該当するキーワード情報を1件取得する
func (u *KeywordUseCase) GetKeywordByID(ctx context.Context, id int64) (*model.Keyword, error) {
	return u.keywordRepo.FindByID(ctx, id)
}

// GetAllKeywords 登録されているすべてのキーワード情報を取得する（ページングなしの全件マスターデータ用）
func (u *KeywordUseCase) GetAllKeywords(ctx context.Context) ([]model.Keyword, error) {
	return u.keywordRepo.FindAll(ctx)
}

// GetKeywordsWithPagination 指定されたページ、件数、検索キーワードに基づいて、ページング・検索適用済みのキーワード情報を取得する
func (u *KeywordUseCase) GetKeywordsWithPagination(
	ctx context.Context,
	page, limit int,
	filter KeywordListFilter,
) ([]model.Keyword, int64, int, error) {
	var whereQuery func(*gorm.DB) *gorm.DB
	if filter.SearchWord != "" {
		whereQuery = func(db *gorm.DB) *gorm.DB {
			likeQuery := "%" + filter.SearchWord + "%"
			return db.Where("name LIKE ? OR kana LIKE ?", likeQuery, likeQuery)
		}
	}

	return usecase.ExecutePaginatedSearch(
		ctx, page, limit, whereQuery,
		u.keywordRepo.CountAll,
		u.keywordRepo.FindAllWithPagination,
	)
}

// -------------------------------------------------------------------------
// U: Update (更新)
// -------------------------------------------------------------------------

// UpdateKeyword 既存のキーワード情報を更新する
func (u *KeywordUseCase) UpdateKeyword(ctx context.Context, g *model.Keyword) error {
	return u.keywordRepo.Update(ctx, g)
}

// -------------------------------------------------------------------------
// D: Delete (削除)
// -------------------------------------------------------------------------

// DeleteKeyword キーワードIDを指定して、該当するキーワード情報を削除する
func (u *KeywordUseCase) DeleteKeyword(ctx context.Context, id int64) error {
	return u.keywordRepo.Delete(ctx, id)
}
