package admin

import "testing"

func TestParsePublishDateTime(t *testing.T) {
	got, err := parsePublishDateTime("2026-09-12T08:30")
	if err != nil {
		t.Fatalf("valid datetime returned error: %v", err)
	}
	if got == nil || got.Format("2006-01-02T15:04 -0700") != "2026-09-12T08:30 +0900" {
		t.Fatalf("unexpected datetime: %v", got)
	}

	empty, err := parsePublishDateTime("")
	if err != nil || empty != nil {
		t.Fatalf("empty datetime = %v, %v; want nil, nil", empty, err)
	}

	if _, err := parsePublishDateTime("2026/09/12 08:30"); err == nil {
		t.Fatal("invalid datetime was accepted")
	}
}
