package app

import "testing"

func TestParsePositiveInt(t *testing.T) {
	tests := []struct {
		raw      string
		fallback int
		want     int
	}{
		{raw: "12", fallback: 1, want: 12},
		{raw: "", fallback: 20, want: 20},
		{raw: "abc", fallback: 20, want: 20},
		{raw: "0", fallback: 20, want: 20},
		{raw: "-1", fallback: 20, want: 20},
	}

	for _, test := range tests {
		if got := parsePositiveInt(test.raw, test.fallback); got != test.want {
			t.Errorf("parsePositiveInt(%q, %d) = %d, want %d", test.raw, test.fallback, got, test.want)
		}
	}
}

func TestParseLimit(t *testing.T) {
	if got := parseLimit("101", 20); got != 100 {
		t.Fatalf("parseLimit over maximum = %d, want 100", got)
	}
	if got := parseLimit("30", 20); got != 30 {
		t.Fatalf("parseLimit valid value = %d, want 30", got)
	}
	if got := parseLimit("invalid", 20); got != 20 {
		t.Fatalf("parseLimit invalid value = %d, want 20", got)
	}
}
