package app

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestVisitorHashUsesVisitorIDBeforeIP(t *testing.T) {
	want := visitorHash("visitor-1", "192.0.2.1")
	got := visitorHash(" visitor-1 ", "198.51.100.1")
	if got != want {
		t.Fatalf("visitor IDが同じ場合はIPに依存しない: got %q want %q", got, want)
	}
}

func TestVisitorHashFallsBackToIP(t *testing.T) {
	if got := visitorHash("", "192.0.2.1"); got == "" {
		t.Fatal("IP fallback did not produce a hash")
	}
	if got := visitorHash("", "  "); got != "" {
		t.Fatalf("empty identity produced hash: %q", got)
	}
}

func TestNormalizePagePath(t *testing.T) {
	if got := normalizePagePath("app/games"); got != "" {
		t.Fatalf("先頭スラッシュのないパスを受理した: %q", got)
	}

	got := normalizePagePath("/" + strings.Repeat("あ", 200))
	if len([]rune(got)) != 191 {
		t.Fatalf("パスを文字単位で191文字に制限できていない: %d", len([]rune(got)))
	}
	if !strings.HasPrefix(got, "/") {
		t.Fatalf("正規化後のパスが不正: %q", got)
	}
	if !utf8.ValidString(got) {
		t.Fatal("truncated path is not valid UTF-8")
	}
}
