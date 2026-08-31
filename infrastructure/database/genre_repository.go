package database

import (
	"context"
	"errors"
	"sort"
	"yutagame-backend/domain/model"

	"gorm.io/gorm"
)

// =========================================================================
// 構造体＆コンストラクタ
// =========================================================================

// GenreRepository ジャンル情報に関するデータベース操作を担当するリポジトリ
type GenreRepository struct {
	db *gorm.DB
}

type GenreWithGameCount struct {
	Code      string  `json:"code"`
	Name      string  `json:"name"`
	ImageKey  *string `json:"imageKey"`
	GameCount int64   `json:"gameCount"`
}

// NewGenreRepository GenreRepositoryの新しいインスタンスを生成するコンストラクタ
func NewGenreRepository(db *gorm.DB) *GenreRepository {
	return &GenreRepository{db: db}
}

// =========================================================================
// C: Create (作成)
// =========================================================================

// Create 新しいジャンル情報をデータベースに登録する
func (r *GenreRepository) Create(ctx context.Context, genre *model.Genre) error {
	return r.db.WithContext(ctx).Create(genre).Error
}

// =========================================================================
// R: Read (取得)
// =========================================================================

// FindByID ジャンルID（主キー）を指定して、該当するジャンル情報を1件取得する
func (r *GenreRepository) FindByID(ctx context.Context, id int64) (*model.Genre, error) {
	var genre model.Genre
	err := r.db.WithContext(ctx).First(&genre, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &genre, err
}

// FindByCode コードを指定して、該当するジャンル情報を1件取得する
func (r *GenreRepository) FindByCode(ctx context.Context, code string) (*model.Genre, error) {
	var genre model.Genre
	err := r.db.WithContext(ctx).Where("code = ?", code).First(&genre).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &genre, err
}

// FindAll 登録されているすべてのジャンル情報をID昇順で取得する（ページングなし）
func (r *GenreRepository) FindAll(ctx context.Context) ([]model.Genre, error) {
	var genres []model.Genre
	err := r.db.WithContext(ctx).Order("id asc").Find(&genres).Error
	return genres, err
}

// FindByIDs 指定されたID群に一致するジャンルをまとめて取得する
func (r *GenreRepository) FindByIDs(ctx context.Context, ids []int64) ([]model.Genre, error) {
	if len(ids) == 0 {
		return []model.Genre{}, nil
	}

	var genres []model.Genre
	err := r.db.WithContext(ctx).
		Where("id IN ?", ids).
		Order("id asc").
		Find(&genres).Error
	if err != nil {
		return nil, err
	}

	// 入力順は保証しないので、ID昇順を明示して返す
	sort.Slice(genres, func(i, j int) bool {
		return genres[i].ID < genres[j].ID
	})

	return genres, nil
}

// FindAllWithPagination 指定された件数（limit）と開始位置（offset）に応じて、ジャンル情報をID昇順で取得する
func (r *GenreRepository) FindAllWithPagination(
	ctx context.Context,
	limit, offset int,
	whereQueries ...func(*gorm.DB) *gorm.DB,
) ([]model.Genre, error) {
	return ExecuteFindWithPagination[model.Genre](ctx, r.db, limit, offset, "id asc", nil, whereQueries...)
}

// CountAll ページングの総ページ数計算のため、条件に合致するジャンル情報の総件数を取得する
func (r *GenreRepository) CountAll(ctx context.Context, whereQueries ...func(*gorm.DB) *gorm.DB) (int64, error) {
	return ExecuteCount[model.Genre](ctx, r.db, whereQueries...)
}

// FindAllWithGameCount 公開画面向けにジャンルごとのゲーム件数を集計して返す
func (r *GenreRepository) FindAllWithGameCount(ctx context.Context) ([]GenreWithGameCount, error) {
	var items []GenreWithGameCount
	err := r.db.WithContext(ctx).
		Table("genres").
		Select(`
			genres.code,
			genres.name,
			genres.image_key,
			COUNT(games.id) AS game_count
		`).
		Joins("LEFT JOIN games ON games.genre_id = genres.id").
		Group("genres.id").
		Order("genres.id asc").
		Find(&items).Error
	return items, err
}

// =========================================================================
// U: Update (更新)
// =========================================================================

// Update 既存のジャンル情報（名前、説明など）を更新する
func (r *GenreRepository) Update(ctx context.Context, genre *model.Genre) error {
	return r.db.WithContext(ctx).Save(genre).Error
}

// UpdateFieldsByID 指定IDのジャンルに対し、指定カラムだけを更新する
func (r *GenreRepository) UpdateFieldsByID(ctx context.Context, id int64, updates map[string]any) error {
	if len(updates) == 0 {
		return nil
	}

	return r.db.WithContext(ctx).
		Model(&model.Genre{}).
		Where("id = ?", id).
		Updates(updates).Error
}

// UpdateImageKey はジャンル画像キーのみを更新する
func (r *GenreRepository) UpdateImageKey(ctx context.Context, id int64, imageKey *string) error {
	return r.db.WithContext(ctx).
		Model(&model.Genre{}).
		Where("id = ?", id).
		Update("image_key", imageKey).Error
}

// =========================================================================
// D: Delete (削除)
// =========================================================================

// Delete ジャンルIDを指定して、該当するジャンル情報を物理削除する
func (r *GenreRepository) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&model.Genre{}, id).Error
}
