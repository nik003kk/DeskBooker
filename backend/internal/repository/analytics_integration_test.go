package repository

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nik003kk/DeskBooker/backend/internal/models"
	"testing"
	"time"
)

func testAnalytics(t *testing.T, pool *pgxpool.Pool, owner string) {
	ctx := context.Background()
	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)
	if _, err = tx.Exec(ctx, "UPDATE public.resources SET is_active=false"); err != nil {
		t.Fatal(err)
	}
	var first, second string
	for i, dest := range []*string{&first, &second} {
		if err = tx.QueryRow(ctx, "INSERT INTO public.resources(name,type,created_at) VALUES($1,'DESK','2020-01-01') RETURNING id", []string{"Analytics A", "Analytics B"}[i]).Scan(dest); err != nil {
			t.Fatal(err)
		}
	}
	loc, _ := time.LoadLocation("Europe/Ljubljana")
	from := time.Date(2030, 1, 7, 0, 0, 0, 0, loc)
	period := models.AnalyticsRange{From: from, Until: from.AddDate(0, 0, 3)}
	for _, tc := range []struct {
		status         string
		hour, duration int
	}{{"PENDING", 9, 1}, {"CONFIRMED", 10, 2}, {"CHECKED_IN", 12, 2}, {"COMPLETED", 14, 1}, {"CANCELLED", 11, 2}, {"EXPIRED", 10, 2}} {
		start := from.Add(time.Duration(tc.hour) * time.Hour)
		var checked any
		if tc.status == "CHECKED_IN" {
			checked = start
		}
		if _, err = tx.Exec(ctx, "INSERT INTO public.reservations(user_id,resource_id,starts_at,ends_at,status,checked_in_at) VALUES($1,$2,$3,$4,$5,$6)", owner, first, start, start.Add(time.Duration(tc.duration)*time.Hour), tc.status, checked); err != nil {
			t.Fatal(err)
		}
	}
	for _, start := range []time.Time{from.Add(-30 * time.Minute), period.Until} {
		if _, err = tx.Exec(ctx, "INSERT INTO public.reservations(user_id,resource_id,starts_at,ends_at,status) VALUES($1,$2,$3,$4,'COMPLETED')", owner, first, start, start.Add(10*time.Hour)); err != nil {
			t.Fatal(err)
		}
	}
	report, err := New(tx).ReservationAnalytics(ctx, period)
	if err != nil || report.TotalReservations != 6 || report.ActiveReservations != 2 || report.GeneratedAt.IsZero() || len(report.StatusCounts) != 6 {
		t.Fatalf("analytics summary %#v %v", report, err)
	}
	for status, count := range report.StatusCounts {
		if count != 1 {
			t.Fatalf("status %s=%d", status, count)
		}
	}
	empty, err := New(tx).ReservationAnalytics(ctx, models.AnalyticsRange{From: from.AddDate(-1, 0, 0), Until: from.AddDate(-1, 0, 1)})
	if err != nil || empty.TotalReservations != 0 || empty.ActiveReservations != 0 {
		t.Fatalf("empty range %#v %v", empty, err)
	}
	if report.NoShow.Eligible != 3 || report.NoShow.Expired != 1 || report.NoShow.Rate == nil || *report.NoShow.Rate != 33.33 || empty.NoShow.Rate != nil {
		t.Fatalf("no-show report %#v empty %#v", report.NoShow, empty.NoShow)
	}
	if len(report.Daily) != 3 || len(report.Weekdays) != 5 || report.OccupancyRate == nil || *report.OccupancyRate != 11.46 {
		t.Fatalf("occupancy report %#v", report)
	}
	monday := report.Daily[0]
	if monday.Date != "2030-01-07" || monday.BookedMinutes != 330 || monday.CapacityMinutes != 960 || monday.OccupancyRate == nil || *monday.OccupancyRate != 34.38 {
		t.Fatalf("Monday clipping/occupancy: %#v", monday)
	}
	if report.Daily[1].BookedMinutes != 0 || report.Weekdays[3].OccupancyRate != nil {
		t.Fatal("empty weekday must be zero, absent weekday must be nil")
	}
	// Completed history can overlap another booking; occupancy is a union of
	// resource time ranges, never a double count of overlapping history.
	if _, err = tx.Exec(ctx, "INSERT INTO public.reservations(user_id,resource_id,starts_at,ends_at,status) VALUES($1,$2,$3,$4,'COMPLETED')", owner, first, from.Add(11*time.Hour), from.Add(13*time.Hour)); err != nil {
		t.Fatal(err)
	}
	overlap, err := New(tx).ReservationAnalytics(ctx, period)
	if err != nil || overlap.Daily[0].BookedMinutes != 330 {
		t.Fatalf("overlapping history double-counted: %#v %v", overlap, err)
	}
	if _, err = tx.Exec(ctx, "UPDATE public.resources SET is_active=false WHERE id=$1", second); err != nil {
		t.Fatal(err)
	}
	reduced, err := New(tx).ReservationAnalytics(ctx, period)
	if err != nil || reduced.Daily[0].CapacityMinutes != 480 || *reduced.Daily[0].OccupancyRate != 68.75 {
		t.Fatalf("inactive capacity: %#v %v", reduced, err)
	}
	if _, err = tx.Exec(ctx, "UPDATE public.resources SET is_active=false WHERE id=$1", first); err != nil {
		t.Fatal(err)
	}
	none, err := New(tx).ReservationAnalytics(ctx, period)
	if err != nil || none.OccupancyRate != nil {
		t.Fatalf("zero capacity should be unknown: %#v %v", none, err)
	}

}
