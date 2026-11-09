package models

import "testing"

func TestNoShowUsesKnownAttendanceOutcomes(t *testing.T) {
	counts := map[string]int64{"EXPIRED": 1, "CHECKED_IN": 2, "COMPLETED": 1, "CONFIRMED": 100, "PENDING": 100, "CANCELLED": 50}
	result := ReservationNoShow(counts)
	if result.Expired != 1 || result.Attended != 3 || result.Eligible != 4 || result.Rate == nil || *result.Rate != 25 {
		t.Fatal(result)
	}
	counts["CHECKED_IN"] = 0
	counts["COMPLETED"] = 3
	if *ReservationNoShow(counts).Rate != 25 {
		t.Fatal("completion changed no-show rate")
	}
	if ReservationNoShow(nil).Rate != nil {
		t.Fatal("empty outcomes should be unknown")
	}
	if *ReservationNoShow(map[string]int64{"EXPIRED": 3}).Rate != 100 {
		t.Fatal("all no-show")
	}
	if *ReservationNoShow(map[string]int64{"COMPLETED": 3}).Rate != 0 {
		t.Fatal("all attended")
	}
}
