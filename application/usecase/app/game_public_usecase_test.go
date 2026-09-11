package app

import "testing"

func TestGameSearchOrder(t *testing.T) {
	tests := []struct {
		key  string
		want string
	}{
		{key: "release_desc", want: "games.release_date desc, games.id desc"},
		{key: "kana_asc", want: "games.kana asc, games.id asc"},
		{key: "price_asc", want: "games.list_price asc, games.id asc"},
		{key: "price_desc", want: "games.list_price desc, games.id asc"},
		{key: "rank_asc", want: "CASE WHEN game_ranking_active_entries.display_rank IS NULL THEN 1 ELSE 0 END asc, game_ranking_active_entries.display_rank asc, games.id asc"},
		{key: "rank_desc", want: "CASE WHEN game_ranking_active_entries.display_rank IS NULL THEN 1 ELSE 0 END asc, game_ranking_active_entries.display_rank desc, games.id asc"},
		{key: "unknown", want: "games.release_date asc, games.id asc"},
	}

	for _, test := range tests {
		t.Run(test.key, func(t *testing.T) {
			if got := gameSearchOrder(test.key); got != test.want {
				t.Fatalf("gameSearchOrder(%q) = %q, want %q", test.key, got, test.want)
			}
		})
	}
}
