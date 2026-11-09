package repository

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nik003kk/DeskBooker/backend/internal/models"
)

func noShowFixture(t *testing.T, pool *pgxpool.Pool, owner string, start time.Time, status string) models.Reservation {
	t.Helper()
	ctx := context.Background()
	var resource, id string
	err := pool.QueryRow(ctx, "INSERT INTO public.resources(name,type) VALUES('No-show ' || gen_random_uuid()::text,'DESK') RETURNING id").Scan(&resource)
	if err != nil {
		t.Fatal(err)
	}
	var checkedAt any
	if status == "CHECKED_IN" {
		checkedAt = start
	}
	err = pool.QueryRow(ctx, "INSERT INTO public.reservations(user_id,resource_id,starts_at,ends_at,status,checked_in_at) VALUES($1,$2,$3,$4,$5,$6) RETURNING id", owner, resource, start, start.Add(2*time.Hour), status, checkedAt).Scan(&id)
	if err != nil {
		t.Fatal(err)
	}
	saved, err := New(pool).GetReservation(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	return saved
}

func testNoShowExpiration(t *testing.T, pool *pgxpool.Pool, owner string) {
	t.Helper()
	ctx := context.Background()
	var now time.Time
	if err := pool.QueryRow(ctx, "SELECT clock_timestamp()").Scan(&now); err != nil {
		t.Fatal(err)
	}
	due := noShowFixture(t, pool, owner, now.Add(-16*time.Minute), "CONFIRMED")
	protected := []models.Reservation{
		noShowFixture(t, pool, owner, now.Add(20*time.Minute), "CONFIRMED"),
		noShowFixture(t, pool, owner, now.Add(-16*time.Minute), "CHECKED_IN"),
		noShowFixture(t, pool, owner, now.Add(-16*time.Minute), "CANCELLED"),
		noShowFixture(t, pool, owner, now.Add(-16*time.Minute), "COMPLETED"),
		noShowFixture(t, pool, owner, now.Add(-16*time.Minute), "PENDING"),
	}
	store := New(pool)
	count, err := store.ExpireNoShows(ctx)
	if err != nil || count != 1 {
		t.Fatalf("first expiration count=%d err=%v", count, err)
	}
	saved, err := store.GetReservation(ctx, due.ID)
	if err != nil || saved.Status != models.ReservationExpired || saved.CheckedInAt != nil {
		t.Fatalf("unexpected expiration: %#v %v", saved, err)
	}
	count, err = store.ExpireNoShows(ctx)
	if err != nil || count != 0 {
		t.Fatalf("repeat expiration count=%d err=%v", count, err)
	}
	for _, original := range protected {
		saved, err := store.GetReservation(ctx, original.ID)
		if err != nil || saved.Status != original.Status {
			t.Fatalf("protected reservation changed: %s %v", original.Status, err)
		}
	}
	available, err := store.IsResourceAvailable(ctx, due.ResourceID, due.StartsAt, due.EndsAt)
	if err != nil || !available {
		t.Fatalf("expired reservation blocks availability: %v %v", available, err)
	}
	matches, err := store.SearchResources(ctx, models.ResourceSearch{StartsAt: &due.StartsAt, EndsAt: &due.EndsAt})
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, resource := range matches {
		if resource.ID == due.ResourceID {
			found = true
		}
	}
	if !found {
		t.Fatal("released resource missing from search")
	}
	start := now.Add(time.Minute).Truncate(time.Minute)
	end := start.Add(30 * time.Minute)
	rebooked, err := NewBookingStore(pool).CreateConfirmedReservation(ctx, models.CreateReservationInput{UserID: owner, ResourceID: due.ResourceID, StartsAt: start, EndsAt: end})
	if err != nil || rebooked.Status != models.ReservationConfirmed {
		t.Fatalf("rebooking released resource: %v", err)
	}
	available, err = store.IsResourceAvailable(ctx, due.ResourceID, start, end)
	if err != nil || available {
		t.Fatalf("new confirmed booking does not block resource: %v %v", available, err)
	}
	history, err := store.GetReservation(ctx, due.ID)
	if err != nil || history.Status != models.ReservationExpired {
		t.Fatalf("rebooking changed history: %v", err)
	}

	var events int
	var event, previous, next, message string
	var occurred time.Time
	err = pool.QueryRow(ctx, "SELECT count(*) FROM public.reservation_audit_log WHERE reservation_id=$1", due.ID).Scan(&events)
	if err != nil || events != 1 {
		t.Fatalf("expected one audit event, got %d: %v", events, err)
	}
	err = pool.QueryRow(ctx, "SELECT event_type,previous_status,new_status,occurred_at,message FROM public.reservation_audit_log WHERE reservation_id=$1", due.ID).Scan(&event, &previous, &next, &occurred, &message)
	if err != nil || event != "NO_SHOW_EXPIRED" || previous != "CONFIRMED" || next != "EXPIRED" || occurred.IsZero() || message != "Rezervacija potekla zaradi odsotnosti check-in-a" {
		t.Fatalf("invalid audit event: %s %s %s %v", event, previous, next, err)
	}
	for _, original := range protected {
		if err := pool.QueryRow(ctx, "SELECT count(*) FROM public.reservation_audit_log WHERE reservation_id=$1", original.ID).Scan(&events); err != nil || events != 0 {
			t.Fatalf("audit created for unchanged reservation: %v", err)
		}
	}

}
