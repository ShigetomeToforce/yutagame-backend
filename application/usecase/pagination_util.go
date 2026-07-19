package usecase

import (
	"context"
	"math"

	"gorm.io/gorm"
)

// PaginationParam ページング計算の結果を保持する構造体
type PaginationParam struct {
	Offset     int // DB（GORM）のOffsetにそのまま渡す値
	TotalPages int // フロントに返却する総ページ数
	ActivePage int // 補正された実際のページ番号
}

// ExecutePaginatedSearch はUseCase層における「件数カウント ➔ ページング計算 ➔ データ取得」の一連の流れを共通化する汎用関数
func ExecutePaginatedSearch[T any](
	ctx context.Context,
	page, limit int,
	whereQuery func(*gorm.DB) *gorm.DB, // 各機能ごとに組み立てたWhere句（nilでもOK）
	countFn func(ctx context.Context, whereQueries ...func(*gorm.DB) *gorm.DB) (int64, error), // 各RepoのCountAll
	findFn func(ctx context.Context, limit, offset int, whereQueries ...func(*gorm.DB) *gorm.DB) ([]T, error), // 各Repo.FindAllWithPagination
) ([]T, int64, int, error) {

	// 1. 可変長引数に渡すためのスライスを用意（nilチェック付き）
	var queries []func(*gorm.DB) *gorm.DB
	if whereQuery != nil {
		queries = append(queries, whereQuery)
	}

	// 2. 全件数（検索時はヒット件数）の取得
	totalCount, err := countFn(ctx, queries...)
	if err != nil {
		return nil, 0, 0, err
	}

	// 3. ページングの計算
	p := CalculatePagination(totalCount, page, limit)

	// 4. 該当ページのデータ取得
	data, err := findFn(ctx, limit, p.Offset, queries...)
	if err != nil {
		return nil, 0, 0, err
	}

	return data, totalCount, p.TotalPages, nil
}

// CalculatePagination 総件数、要求ページ、リミットから、各種ページングパラメータを安全に計算する汎用関数
func CalculatePagination(totalCount int64, page, limit int) PaginationParam {
	// 1. 総ページ数の計算（切り上げ）
	totalPages := int(math.Ceil(float64(totalCount) / float64(limit)))
	if totalPages < 1 {
		totalPages = 1
	}

	// 2. 要求ページが最大ページを超えていた場合の安全な最終ページ補正
	activePage := page
	if activePage > totalPages {
		activePage = totalPages
	}
	if activePage < 1 {
		activePage = 1
	}

	// 3. リポジトリ用の OFFSET の計算
	offset := (activePage - 1) * limit

	return PaginationParam{
		Offset:     offset,
		TotalPages: totalPages,
		ActivePage: activePage,
	}
}
