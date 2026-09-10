package database

import (
	"context"
	"errors"
	"math/rand"
	"sort"
	"strings"
	"time"
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
		Preload("Affiliates").
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
		Preload("Affiliates").
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
		Preload("Affiliates").
		Order("id asc")
	err := query.Find(&games).Error
	if err != nil {
		return nil, err
	}
	return games, nil
}

// FindByIDs 指定されたID群に一致するゲームをまとめて取得する
func (r *GameRepository) FindByIDs(ctx context.Context, ids []int64) ([]model.Game, error) {
	if len(ids) == 0 {
		return []model.Game{}, nil
	}

	var games []model.Game
	err := r.db.WithContext(ctx).
		Where("id IN ?", ids).
		Order("id asc").
		Find(&games).Error
	if err != nil {
		return nil, err
	}

	sort.Slice(games, func(i, j int) bool {
		return games[i].ID < games[j].ID
	})

	return games, nil
}

func (r *GameRepository) FindByIDsWithRelations(ctx context.Context, ids []int64) ([]model.Game, error) {
	if len(ids) == 0 {
		return []model.Game{}, nil
	}

	var games []model.Game
	err := r.db.WithContext(ctx).
		Preload("Manufacturer").
		Preload("Machine").
		Preload("Genre").
		Preload("Keywords").
		Preload("Affiliates").
		Where("id IN ?", ids).
		Order("id asc").
		Find(&games).Error
	if err != nil {
		return nil, err
	}

	sort.Slice(games, func(i, j int) bool {
		return games[i].ID < games[j].ID
	})

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
			Preload("Keywords").
			Preload("Affiliates")
	}
	return ExecuteFindWithPagination[model.Game](
		ctx, r.db, limit, offset, "release_date asc", modifier, whereQueries...)
}

func (r *GameRepository) FindAllWithPaginationByOrder(
	ctx context.Context,
	limit, offset int,
	order string,
	whereQueries ...func(*gorm.DB) *gorm.DB,
) ([]model.Game, error) {
	modifier := func(db *gorm.DB) *gorm.DB {
		if strings.Contains(order, "game_ranking_active_entries") {
			db = db.Joins("LEFT JOIN game_ranking_active_entries ON game_ranking_active_entries.game_id = games.id")
		}
		return db.Preload("Manufacturer").
			Preload("Machine").
			Preload("Genre").
			Preload("Keywords").
			Preload("Affiliates")
	}
	return ExecuteFindWithPagination[model.Game](
		ctx, r.db, limit, offset, order, modifier, whereQueries...)
}

// CountAll ページングの総ページ数計算のため、条件に合致する機種情報の総件数を取得する
func (r *GameRepository) CountAll(ctx context.Context, whereQueries ...func(*gorm.DB) *gorm.DB) (int64, error) {
	return ExecuteCount[model.Game](ctx, r.db, whereQueries...)
}

// FindReleasedOnMonthDay 指定日を起点に、同月日のゲームを優先しつつ、
// 足りない分は翌日以降の近い発売日順で補完して取得する
func (r *GameRepository) FindReleasedOnMonthDay(ctx context.Context, month, day, limit int) ([]model.Game, error) {
	if limit < 1 {
		return []model.Game{}, nil
	}

	var games []model.Game
	err := r.db.WithContext(ctx).
		Preload("Manufacturer").
		Preload("Machine").
		Preload("Genre").
		Preload("Keywords").
		Preload("Affiliates").
		Order("release_date asc, id asc").
		Find(&games).Error
	if err != nil {
		return nil, err
	}

	const calendarYear = 2000
	target := time.Date(calendarYear, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	base := time.Date(calendarYear, time.January, 1, 0, 0, 0, 0, time.UTC)
	daysInYear := int(time.Date(calendarYear+1, time.January, 1, 0, 0, 0, 0, time.UTC).Sub(base).Hours() / 24)

	groupMap := map[int][]model.Game{}
	for _, game := range games {
		gameDate := time.Date(calendarYear, game.ReleaseDate.Month(), game.ReleaseDate.Day(), 0, 0, 0, 0, time.UTC)
		distance := int(gameDate.Sub(target).Hours() / 24)
		if distance < 0 {
			distance += daysInYear
		}
		groupMap[distance] = append(groupMap[distance], game)
	}

	distances := make([]int, 0, len(groupMap))
	for distance := range groupMap {
		distances = append(distances, distance)
	}
	sort.Ints(distances)

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	result := make([]model.Game, 0, limit)
	for _, distance := range distances {
		group := groupMap[distance]
		if distance == 0 {
			rng.Shuffle(len(group), func(i, j int) {
				group[i], group[j] = group[j], group[i]
			})
		} else {
			sort.Slice(group, func(i, j int) bool {
				if group[i].ReleaseDate.Equal(group[j].ReleaseDate) {
					return group[i].ID > group[j].ID
				}
				return group[i].ReleaseDate.After(group[j].ReleaseDate)
			})
		}

		if distance == 0 {
			result = append(result, group...)
			if len(result) >= limit {
				return result, nil
			}
			continue
		}

		remaining := limit - len(result)
		if remaining <= 0 {
			break
		}
		if len(group) > remaining {
			group = group[:remaining]
		}
		result = append(result, group...)
		if len(result) >= limit {
			break
		}
	}

	return result, nil
}

// FindRecentlyReleased 発売日が新しいゲームを取得する
func (r *GameRepository) FindRecentlyReleased(ctx context.Context, limit int) ([]model.Game, error) {
	var games []model.Game
	err := r.db.WithContext(ctx).
		Preload("Manufacturer").
		Preload("Machine").
		Preload("Genre").
		Preload("Keywords").
		Preload("Affiliates").
		Order("release_date desc, id desc").
		Limit(limit).
		Find(&games).Error
	if err != nil {
		return nil, err
	}
	return games, nil
}

// FindRecentlyUpdated 登録日が新しいゲームを取得する
func (r *GameRepository) FindRecentlyUpdated(ctx context.Context, limit int) ([]model.Game, error) {
	var games []model.Game
	err := r.db.WithContext(ctx).
		Preload("Manufacturer").
		Preload("Machine").
		Preload("Genre").
		Preload("Keywords").
		Preload("Affiliates").
		Order("created_at desc, id desc").
		Limit(limit).
		Find(&games).Error
	if err != nil {
		return nil, err
	}
	return games, nil
}

// FindRandom ランダムにゲームを取得する
func (r *GameRepository) FindRandom(ctx context.Context, limit int) ([]model.Game, error) {
	var games []model.Game
	err := r.db.WithContext(ctx).
		Preload("Manufacturer").
		Preload("Machine").
		Preload("Genre").
		Preload("Keywords").
		Preload("Affiliates").
		Order("RAND()").
		Limit(limit).
		Find(&games).Error
	if err != nil {
		return nil, err
	}
	return games, nil
}

func (r *GameRepository) FindAllCodes(ctx context.Context) ([]string, error) {
	var codes []string
	err := r.db.WithContext(ctx).
		Model(&model.Game{}).
		Order("id asc").
		Pluck("code", &codes).Error
	return codes, err
}

// =========================================================================
// U: Update (更新)
// =========================================================================

// Update は既存のゲーム情報を更新します
func (r *GameRepository) Update(ctx context.Context, g *model.Game) error {
	return r.db.WithContext(ctx).Save(g).Error
}

// UpdateFieldsByID 指定IDのゲームに対し、指定カラムだけを更新する
func (r *GameRepository) UpdateFieldsByID(ctx context.Context, id int64, updates map[string]any) error {
	if len(updates) == 0 {
		return nil
	}

	return r.db.WithContext(ctx).
		Model(&model.Game{}).
		Where("id = ?", id).
		Updates(updates).Error
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

func (r *GameRepository) ListAffiliatesByGameID(ctx context.Context, gameID int64) ([]model.GameAffiliate, error) {
	var items []model.GameAffiliate
	err := r.db.WithContext(ctx).
		Where("game_id = ?", gameID).
		Order("id asc").
		Find(&items).Error
	if err != nil {
		return nil, err
	}
	return items, nil
}

func (r *GameRepository) FindAffiliateByID(ctx context.Context, id int64) (*model.GameAffiliate, error) {
	var item model.GameAffiliate
	err := r.db.WithContext(ctx).
		First(&item, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *GameRepository) FindAffiliateByGameIDAndCategory(ctx context.Context, gameID int64, category string) (*model.GameAffiliate, error) {
	var item model.GameAffiliate
	err := r.db.WithContext(ctx).
		Where("game_id = ? AND category = ?", gameID, category).
		First(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *GameRepository) CreateAffiliate(ctx context.Context, item *model.GameAffiliate) error {
	return r.db.WithContext(ctx).Create(item).Error
}

func (r *GameRepository) UpdateAffiliate(ctx context.Context, item *model.GameAffiliate) error {
	return r.db.WithContext(ctx).Save(item).Error
}

func (r *GameRepository) DeleteAffiliate(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&model.GameAffiliate{}, id).Error
}
