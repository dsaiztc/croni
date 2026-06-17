package cron

import (
	"testing"
)

func TestExpandSimple(t *testing.T) {
	intervals, err := Expand("0 * * * *")
	if err != nil {
		t.Fatal(err)
	}
	if len(intervals) != 1 {
		t.Fatalf("expected 1 interval, got %d", len(intervals))
	}
	if *intervals[0].Minute != 0 {
		t.Errorf("expected minute 0, got %d", *intervals[0].Minute)
	}
	if intervals[0].Hour != nil {
		t.Error("expected nil hour")
	}
}

func TestExpandEvery15Min(t *testing.T) {
	intervals, err := Expand("*/15 * * * *")
	if err != nil {
		t.Fatal(err)
	}
	if len(intervals) != 4 {
		t.Fatalf("expected 4 intervals, got %d", len(intervals))
	}
	expected := []int{0, 15, 30, 45}
	for i, e := range expected {
		if *intervals[i].Minute != e {
			t.Errorf("interval %d: expected minute %d, got %d", i, e, *intervals[i].Minute)
		}
	}
}

func TestExpandWorkHours(t *testing.T) {
	intervals, err := Expand("*/15 9-17 * * 1-5")
	if err != nil {
		t.Fatal(err)
	}
	if len(intervals) != 180 {
		t.Fatalf("expected 180 intervals, got %d", len(intervals))
	}
}

func TestExpandDayAndWeekdayOR(t *testing.T) {
	intervals, err := Expand("0 9 1 * 5")
	if err != nil {
		t.Fatal(err)
	}
	if len(intervals) != 2 {
		t.Fatalf("expected 2 intervals (OR logic), got %d", len(intervals))
	}
	if intervals[0].Day == nil || intervals[0].Weekday != nil {
		t.Error("first interval should have Day set, no Weekday")
	}
	if intervals[1].Weekday == nil || intervals[1].Day != nil {
		t.Error("second interval should have Weekday set, no Day")
	}
}

func TestExpandWildcard(t *testing.T) {
	_, err := Expand("* * * * *")
	if err != nil {
		t.Fatal(err)
	}
}

func TestExpandInvalid(t *testing.T) {
	cases := []string{
		"bad",
		"* * *",
		"60 * * * *",
		"* 25 * * *",
	}
	for _, c := range cases {
		_, err := Expand(c)
		if err == nil {
			t.Errorf("expected error for %q", c)
		}
	}
}

func TestExpandList(t *testing.T) {
	intervals, err := Expand("0,30 9,18 * * *")
	if err != nil {
		t.Fatal(err)
	}
	if len(intervals) != 4 {
		t.Fatalf("expected 4 intervals, got %d", len(intervals))
	}
}

func TestExpandSpecialExpressions(t *testing.T) {
	cases := []struct {
		expr     string
		expected int
	}{
		{"@hourly", 1},
		{"@daily", 1},
		{"@midnight", 1},
		{"@weekly", 1},
		{"@monthly", 1},
		{"@yearly", 1},
		{"@annually", 1},
	}
	for _, c := range cases {
		intervals, err := Expand(c.expr)
		if err != nil {
			t.Fatalf("%s: %v", c.expr, err)
		}
		if len(intervals) != c.expected {
			t.Errorf("%s: expected %d intervals, got %d", c.expr, c.expected, len(intervals))
		}
	}
}

func TestExpandDaily(t *testing.T) {
	intervals, err := Expand("@daily")
	if err != nil {
		t.Fatal(err)
	}
	if *intervals[0].Minute != 0 || *intervals[0].Hour != 0 {
		t.Errorf("@daily should be minute=0 hour=0, got minute=%d hour=%d",
			*intervals[0].Minute, *intervals[0].Hour)
	}
}

func TestExpandNamedDays(t *testing.T) {
	intervals, err := Expand("0 9 * * MON-FRI")
	if err != nil {
		t.Fatal(err)
	}
	if len(intervals) != 5 {
		t.Fatalf("expected 5 intervals for MON-FRI, got %d", len(intervals))
	}
}

func TestExpandNamedMonths(t *testing.T) {
	intervals, err := Expand("0 0 1 JAN,JUN *")
	if err != nil {
		t.Fatal(err)
	}
	if len(intervals) != 2 {
		t.Fatalf("expected 2 intervals for JAN,JUN, got %d", len(intervals))
	}
}

func TestExpandNamedMixedCase(t *testing.T) {
	intervals, err := Expand("0 9 * * mon,Wed,FRI")
	if err != nil {
		t.Fatal(err)
	}
	if len(intervals) != 3 {
		t.Fatalf("expected 3 intervals, got %d", len(intervals))
	}
}
