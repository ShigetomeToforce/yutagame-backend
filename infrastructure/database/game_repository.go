package database

import (
	"context"
	"errors"
	"yutagame-backend/domain/model"

	"gorm.io/gorm"
)

// =========================================================================
// 構造体＆コンストラクタ
// =========================================================================

// GameRepository ゲーム情報に関するデータベース操作を担当するリポジトリ
type GameRepository struct {
	db *gorm.DB
}

// NewGameRepository GameRepositoryの新しいインスタンスを生成するコンストラクタ
func NewGameRepository(db *gorm.DB) *GameRepository {
	return &GameRepository{db: db}
}

// =========================================================================
// C: Create (作成)
// =========================================================================

// Create 新しいゲーム情報をデータベースに登録する
func (r *GameRepository) Create(ctx context.Context, g *model.Game) error {
	return r.db.WithContext(ctx).Create(g).Error
}

// =========================================================================
// R: Read (取得)
// =========================================================================

// FindByID ゲームID（主キー）を指定して、該当するゲーム情報を1件取得する
func (r *GameRepository) FindByID(ctx context.Context, id int64) (*model.Game, error) {
	var game model.Game
	err := r.db.WithContext(ctx).
		Preload("Manufacturer").
		Preload("Machine").
		Preload("Genre").
		Preload("Keywords").
		First(&game, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &game, err
}

// FindByCode コードを指定して、該当するゲーム情報を1件取得する
func (r *GameRepository) FindByCode(ctx context.Context, code string) (*model.Game, error) {
	var game model.Game
	err := r.db.WithContext(ctx).
		Preload("Manufacturer").
		Preload("Machine").
		Preload("Genre").
		Preload("Keywords").
		Where("code = ?", code).
		First(&game).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &game, err
}

// FindAll 登録されているすべてのゲーム情報をID昇順で取得する（ページングなし）
func (r *GameRepository) FindAll(ctx context.Context) ([]model.Game, error) {
	var games []model.Game
	query := r.db.WithContext(ctx).
		Preload("Manufacturer").
		Preload("Machine").
		Preload("Genre").
		Preload("Keywords").
		Order("id asc")
	err := query.Find(&games).Error
	if err != nil {
		return nil, err
	}
	return games, nil
}

// FindAllWithPagination 指定された件数（limit）と開始位置（offset）に応じて、ゲーム情報を発売日昇順で取得する
func (r *GameRepository) FindAllWithPagination(
	ctx context.Context,
	limit, offset int,
	whereQueries ...func(*gorm.DB) *gorm.DB,
) ([]model.Game, error) {
	modifier := func(db *gorm.DB) *gorm.DB {
		return db.Preload("Manufacturer").
			Preload("Machine").
			Preload("Genre").
			Preload("Keywords")
	}
	return ExecuteFindWithPagination[model.Game](
		ctx, r.db, limit, offset, "release_date asc", modifier, whereQueries...)
}

// CountAll ページングの総ページ数計算のため、条件に合致する機種情報の総件数を取得する
func (r *GameRepository) CountAll(ctx context.Context, whereQueries ...func(*gorm.DB) *gorm.DB) (int64, error) {
	return ExecuteCount[model.Game](ctx, r.db, whereQueries...)
}

// =========================================================================
// U: Update (更新)
// =========================================================================

// Update は既存のゲーム情報を更新します
func (r *GameRepository) Update(ctx context.Context, g *model.Game) error {
	return r.db.WithContext(ctx).Save(g).Error
}

// UpdateImageKey はゲーム画像キーのみを更新する
func (r *GameRepository) UpdateImageKey(ctx context.Context, id int64, imageKey *string) error {
	return r.db.WithContext(ctx).
		Model(&model.Game{}).
		Where("id = ?", id).
		Update("image_key", imageKey).Error
}

// =========================================================================
// D: Delete (削除)
// =========================================================================

// Delete はゲームIDを指定して、該当するゲーム情報を物理削除する。
// データベース側の ON DELETE CASCADE 設定により、中間テーブル（game_keywords）の紐付けデータもMySQLが自動で連動削除
func (r *GameRepository) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&model.Game{}, id).Error
}

// =========================================================================
// O: Other (その他)
// =========================================================================

// ゲーム情報とキーワード情報の中間テーブルを更新
func (r *GameRepository) ReplaceKeywords(ctx context.Context, gameID int64, keywordIDs []int64) error {
	var keywords []model.Keyword
	if len(keywordIDs) == 0 {
		return r.db.WithContext(ctx).Model(&model.Game{ID: gameID}).Association("Keywords").Clear()
	}

	if err := r.db.WithContext(ctx).
		Where("id IN ?", keywordIDs).
		Find(&keywords).Error; err != nil {
		return err
	}

	return r.db.WithContext(ctx).
		Model(&model.Game{ID: gameID}).
		Association("Keywords").
		Replace(keywords)
}
