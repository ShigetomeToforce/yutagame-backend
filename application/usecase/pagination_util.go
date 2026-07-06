package usecase

import "math"

// PaginationParam ページング計算の結果を保持する構造体
type PaginationParam struct {
	Offset     int // DB（GORM）のOffsetにそのまま渡す値
	TotalPages int // フロントに返却する総ページ数
	ActivePage int // 補正された実際のページ番号
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
