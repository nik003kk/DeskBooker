package repository

import (
	"testing"
	"time"
)

func TestCheckInWindowExactBoundaries(t *testing.T) {
	start := time.Date(2030, 1, 10, 9, 0, 0, 0, time.UTC)
	end := start.Add(time.Hour)
	for _, tc := range []struct {
		name   string
		offset time.Duration
		want   bool
	}{
		{"before opening", -10*time.Minute - time.Nanosecond, false},
		{"opening", -10 * time.Minute, true},
		{"start", 0, true},
		{"last instant", 15 * time.Minute, true},
		{"after closing", 15*time.Minute + time.Nanosecond, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := checkInWindowOpen(start.Add(tc.offset), start, end); got != tc.want {
				t.Fatalf("got %v want %v", got, tc.want)
			}
		})
	}
	if checkInWindowOpen(start.Add(5*time.Minute), start, start.Add(5*time.Minute)) {
		t.Fatal("ended short reservation accepted")
	}
	if checkInWindowOpen(start, start, start) || checkInWindowOpen(start, time.Time{}, end) {
		t.Fatal("invalid window accepted")
	}
}

func TestCheckInWindowUsesUTCInstantsAcrossDST(t *testing.T) {
	start, err := time.Parse(time.RFC3339, "2030-10-27T02:00:00+01:00")
	if err != nil {
		t.Fatal(err)
	}
	now, err := time.Parse(time.RFC3339, "2030-10-27T02:55:00+02:00")
	if err != nil {
		t.Fatal(err)
	}
	if !checkInWindowOpen(now, start, start.Add(time.Hour)) {
		t.Fatal("same instant with different offset rejected")
	}
}
