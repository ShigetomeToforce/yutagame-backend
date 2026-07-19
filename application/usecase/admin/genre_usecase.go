package admin

import (
	"context"
	"yutagame-backend/application/usecase"
	"yutagame-backend/domain/model"
	"yutagame-backend/infrastructure/database"

	"gorm.io/gorm"
)

// GenreUseCase ジャンル管理のビジネスロジックを担当するユースケース
type GenreUseCase struct {
	genreRepo *database.GenreRepository
}

// NewGenreUseCase GenreUseCaseの新しいインスタンスを生成するコンストラクタ
func NewGenreUseCase(genreRepo *database.GenreRepository) *GenreUseCase {
	return &GenreUseCase{genreRepo: genreRepo}
}

// =========================================================================
// 🛠️ Genre Management CRUD (ジャンル管理ロジック) - ルーティングのガード内側で利用
// =========================================================================

// -------------------------------------------------------------------------
// C: Create (作成)
// -------------------------------------------------------------------------

// CreateGenre 新しいジャンルを作成する
func (u *GenreUseCase) CreateGenre(ctx context.Context, g *model.Genre) error {
	return u.genreRepo.Create(ctx, g)
}

// -------------------------------------------------------------------------
// R: Read (取得)
// -------------------------------------------------------------------------

// GetGenreByID ジャンルIDを指定して、該当するジャンル情報を1件取得する
func (u *GenreUseCase) GetGenreByID(ctx context.Context, id int64) (*model.Genre, error) {
	return u.genreRepo.FindByID(ctx, id)
}

// GetAllGenres 登録されているすべてのジャンル情報を取得する（ページングなしの全件マスターデータ用）
func (u *GenreUseCase) GetAllGenres(ctx context.Context) ([]model.Genre, error) {
	return u.genreRepo.FindAll(ctx)
}

// GetGenresWithPagination 指定されたページ、件数、検索キーワードに基づいて、ページング・検索適用済みのジャンル一覧を取得する
func (u *GenreUseCase) GetGenresWithPagination(ctx context.Context, page, limit int, searchWord string) ([]model.Genre, int64, int, error) {
	var whereQuery func(*gorm.DB) *gorm.DB
	if searchWord != "" {
		whereQuery = func(db *gorm.DB) *gorm.DB {
			likeQuery := "%" + searchWord + "%"
			return db.Where("name LIKE ? OR kana LIKE ?", likeQuery, likeQuery)
		}
	}

	return usecase.ExecutePaginatedSearch(
		ctx, page, limit, whereQuery,
		u.genreRepo.CountAll,
		u.genreRepo.FindAllWithPagination,
	)
}

// -------------------------------------------------------------------------
// U: Update (更新)
// -------------------------------------------------------------------------

// UpdateGenre 既存のジャンル情報を更新する
func (u *GenreUseCase) UpdateGenre(ctx context.Context, g *model.Genre) error {
	return u.genreRepo.Update(ctx, g)
}

// -------------------------------------------------------------------------
// D: Delete (削除)
// -------------------------------------------------------------------------

// DeleteGenre ジャンルIDを指定して、該当するジャンル情報を削除する
func (u *GenreUseCase) DeleteGenre(ctx context.Context, id int64) error {
	return u.genreRepo.Delete(ctx, id)
}
