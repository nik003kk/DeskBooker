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
}
