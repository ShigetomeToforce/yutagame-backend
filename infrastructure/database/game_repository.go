package database

import (
	"context"
	"errors"
	"yutagame-backend/domain/model"

	"gorm.io/gorm"
)

type GameRepository struct {
	db *gorm.DB
}

func NewGameRepository(db *gorm.DB) *GameRepository {
	return &GameRepository{db: db}
}

// FindAll は条件に応じてゲーム一覧を取得します（絞り込み ＆ キーワード同時取得対応）
func (r *GameRepository) FindAll(ctx context.Context, machineID int64, manufacturerID int64, keywordID int64) ([]model.Game, error) {
	var games []model.Game

	// 💡 Preload("Keywords") と書くだけで、GORMが中間テーブルを自動JOINして
	// 構造体の中の Keywords スライスにデータを全自動で詰め込んでくれます！
	tx := r.db.WithContext(ctx).Preload("Keywords").Order("release_date asc")

	// 機種IDによる絞り込み
	if machineID > 0 {
		tx = tx.Where("machine_id = ?", machineID)
	}

	// メーカーIDによる絞り込み
	if manufacturerID > 0 {
		tx = tx.Where("manufacturer_id = ?", manufacturerID)
	}

	// 💡 中間テーブル化の恩恵：特定のキーワードIDでの絞り込みも爆速かつスマートに実現可能
	if keywordID > 0 {
		// game_keywords 中間テーブルに存在する game_id を安全にサブクエリで絞り込みます
		tx = tx.Where("id IN (SELECT game_id FROM game_keywords WHERE keyword_id = ?)", keywordID)
	}

	// クエリ実行
	err := tx.Find(&games).Error
	if err != nil {
		return nil, err
	}

	return games, nil
}

// FindByID は指定されたIDのゲームを1件取得します（詳細画面用）
func (r *GameRepository) FindByID(ctx context.Context, id int64) (*model.Game, error) {
	var game model.Game
	err := r.db.WithContext(ctx).Preload("Keywords").First(&game, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &game, nil
}

// Create は新しいゲームを登録します（多対多の紐付けもGORMが自動処理）
func (r *GameRepository) Create(ctx context.Context, g *model.Game) error {
	return r.db.WithContext(ctx).Create(g).Error
}

// Update は既存のゲーム情報を更新します
func (r *GameRepository) Update(ctx context.Context, g *model.Game) error {
	return r.db.WithContext(ctx).Save(g).Error
}

// Delete はゲームを削除します（中間テーブルのレコードはDBのON DELETE CASCADEにより自動連動削除されます）
func (r *GameRepository) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&model.Game{}, id).Error
}
