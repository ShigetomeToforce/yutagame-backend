package admin

import (
	"testing"
	"yutagame-backend/application/usecase"
)

func TestParseDateStartUsesJapanLocation(t *testing.T) {
	got, ok := parseDateStart(" 2026-09-12 ")
	if !ok {
		t.Fatal("valid date was rejected")
	}
	if got.Format("2006-01-02 -0700") != "2026-09-12 +0900" {
		t.Fatalf("unexpected date or timezone: %s", got.Format("2006-01-02 -0700"))
	}
	if got.Location() != usecase.JapanLocation {
		t.Fatalf("unexpected location: %v", got.Location())
	}
	if _, ok := parseDateStart("2026-02-30"); ok {
		t.Fatal("invalid date was accepted")
	}
}

func TestResolvePeriodRange(t *testing.T) {
	from, to, target := resolvePeriodRange("daily", "2026-09-12", "")
	if target != "2026-09-12" || to.Sub(from).Hours() != 24 {
		t.Fatalf("unexpected daily range: %v %v %q", from, to, target)
	}

	from, to, target = resolvePeriodRange("monthly", "", "2024-02")
	if target != "2024-02" || from.Format("2006-01-02") != "2024-02-01" || to.Format("2006-01-02") != "2024-03-01" {
		t.Fatalf("unexpected monthly range: %v %v %q", from, to, target)
	}

	from, to, target = resolvePeriodRange("total", "", "")
	if !from.IsZero() || !to.IsZero() || target != "累計" {
		t.Fatalf("unexpected total range: %v %v %q", from, to, target)
	}
}

func TestNormalizeAnalyticsParameters(t *testing.T) {
	if got := normalizePeriod(" yearly "); got != "daily" {
		t.Fatalf("unsupported period = %q, want daily", got)
	}
	if got := normalizePeriod(" monthly "); got != "monthly" {
		t.Fatalf("monthly period = %q", got)
	}
	if got := normalizeBreakdownField("unknown"); got != "genreCode" {
		t.Fatalf("unsupported field = %q, want genreCode", got)
	}
	if got := normalizeBreakdownField(" searchWord "); got != "searchWord" {
		t.Fatalf("searchWord field = %q", got)
	}
}
