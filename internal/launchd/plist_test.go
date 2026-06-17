package launchd

import (
	"strings"
	"testing"
	"time"
)

func TestParseAt_RelativeIsIdempotentWhenStoredAsRFC3339(t *testing.T) {
	// This test documents the correct usage pattern: resolve the relative
	// expression once, store the RFC3339 result, and re-parse that.
	// If the raw relative string "20m" is re-parsed later, time.Now() shifts
	// the result on every call — this test shows the fix works correctly.

	before := time.Now()
	resolved, err := ParseAt("20m")
	after := time.Now()
	if err != nil {
		t.Fatalf("ParseAt(20m): %v", err)
	}

	if resolved.Before(before.Add(20*time.Minute)) || resolved.After(after.Add(20*time.Minute+time.Second)) {
		t.Errorf("ParseAt(20m) = %v, want ~now+20m", resolved)
	}

	// Round-trip: store as RFC3339, re-parse — result must be stable.
	stored := resolved.Format(time.RFC3339)
	reparsed, err := ParseAt(stored)
	if err != nil {
		t.Fatalf("ParseAt(RFC3339): %v", err)
	}
	// Allow 1s rounding from RFC3339 truncation.
	if reparsed.Sub(resolved).Abs() > time.Second {
		t.Errorf("round-trip drift: got %v, want %v (diff=%v)", reparsed, resolved, reparsed.Sub(resolved))
	}
}

func TestParseAt_RelativeFormats(t *testing.T) {
	before := time.Now()
	tests := []struct {
		expr   string
		wantIn time.Duration
	}{
		{"30s", 30 * time.Second},
		{"5m", 5 * time.Minute},
		{"2h", 2 * time.Hour},
	}
	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			got, err := ParseAt(tt.expr)
			if err != nil {
				t.Fatalf("ParseAt(%q): %v", tt.expr, err)
			}
			wantMin := before.Add(tt.wantIn)
			wantMax := wantMin.Add(2 * time.Second)
			if got.Before(wantMin) || got.After(wantMax) {
				t.Errorf("ParseAt(%q) = %v, want in [%v, %v]", tt.expr, got, wantMin, wantMax)
			}
		})
	}
}

func TestParseAt_AbsoluteFormats(t *testing.T) {
	tests := []struct {
		expr string
		want string // date portion
	}{
		{"2030-01-15T09:30:00", "2030-01-15"},
		{"2030-01-15T09:30", "2030-01-15"},
		{"2030-01-15 09:30:05", "2030-01-15"},
	}
	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			got, err := ParseAt(tt.expr)
			if err != nil {
				t.Fatalf("ParseAt(%q): %v", tt.expr, err)
			}
			if !strings.HasPrefix(got.Format("2006-01-02"), tt.want) {
				t.Errorf("ParseAt(%q) date = %v, want %v", tt.expr, got.Format("2006-01-02"), tt.want)
			}
		})
	}
}

func TestParseAt_InvalidExpressions(t *testing.T) {
	invalids := []string{"", "xyz", "0m", "-5m", "abc", "5x"}
	for _, expr := range invalids {
		t.Run(expr, func(t *testing.T) {
			_, err := ParseAt(expr)
			if err == nil {
				t.Errorf("ParseAt(%q) expected error, got nil", expr)
			}
		})
	}
}
