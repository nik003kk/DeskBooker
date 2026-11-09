package repository

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nik003kk/DeskBooker/backend/internal/models"
)

// Run only against an empty disposable test database. CI provisions PostgreSQL.
func TestCheckInIntegration(t *testing.T) {
	url := os.Getenv("CHECK_IN_TEST_DATABASE_URL")
	if url == "" {
		t.Skip("CHECK_IN_TEST_DATABASE_URL is not configured")
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	paths, err := filepath.Glob("../../migrations/*.up.sql")
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range paths {
		sql, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if _, err = pool.Exec(ctx, string(sql)); err != nil {
			t.Fatalf("migration %s: %v", path, err)
		}
	}
	var owner, other, resource, id string
	for i, dest := range []*string{&owner, &other} {
		err = pool.QueryRow(ctx, "INSERT INTO public.users(email,password_hash,first_name,last_name) VALUES($1,'hash','Test','User') RETURNING id", []string{"owner@example.test", "other@example.test"}[i]).Scan(dest)
		if err != nil {
			t.Fatal(err)
		}
	}
	err = pool.QueryRow(ctx, "INSERT INTO public.resources(name,type) VALUES('Check-in desk','DESK') RETURNING id").Scan(&resource)
	if err != nil {
		t.Fatal(err)
	}
	err = pool.QueryRow(ctx, "INSERT INTO public.reservations(user_id,resource_id,starts_at,ends_at,status,check_in_code) VALUES($1,$2,clock_timestamp(),clock_timestamp()+interval '1 hour','CONFIRMED','0042') RETURNING id", owner, resource).Scan(&id)
	if err != nil {
		t.Fatal(err)
	}
	store := NewBookingStore(pool)
	if _, err = store.CheckInReservation(ctx, id, other, "0042"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("other owner: %v", err)
	}
	if _, err = store.CheckInReservation(ctx, id, owner, "9999"); !errors.Is(err, ErrCheckInPIN) {
		t.Fatalf("incorrect PIN: %v", err)
	}
	saved, err := New(pool).GetReservation(ctx, id)
	if err != nil || saved.Status != models.ReservationConfirmed || saved.CheckedInAt != nil {
		t.Fatalf("failed attempt mutated reservation: %#v %v", saved, err)
	}
	var wg sync.WaitGroup
	results := make(chan models.Reservation, 2)
	failures := make(chan error, 2)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			r, e := store.CheckInReservation(ctx, id, owner, "0042")
			results <- r
			failures <- e
		}()
	}
	wg.Wait()
	close(results)
	close(failures)
	for err := range failures {
		if err != nil {
			t.Fatal(err)
		}
	}
	var first *time.Time
	for r := range results {
		if r.Status != models.ReservationCheckedIn || r.CheckedInAt == nil {
			t.Fatalf("missing check-in: %#v", r)
		}
		if first == nil {
			first = r.CheckedInAt
		} else if !first.Equal(*r.CheckedInAt) {
			t.Fatal("retry changed checked_in_at")
		}
	}
	available, err := New(pool).IsResourceAvailable(ctx, resource, saved.StartsAt, saved.EndsAt)
	if err != nil || available {
		t.Fatalf("checked-in resource available: %v %v", available, err)
	}
	matches, err := New(pool).SearchResources(ctx, models.ResourceSearch{StartsAt: &saved.StartsAt, EndsAt: &saved.EndsAt})
	if err != nil || len(matches) != 0 {
		t.Fatalf("checked-in resource in search: %v %v", matches, err)
	}
	_, err = pool.Exec(ctx, "INSERT INTO public.reservations(user_id,resource_id,starts_at,ends_at) VALUES($1,$2,$3,$4)", other, resource, saved.StartsAt, saved.EndsAt)
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != "23P01" {
		t.Fatalf("checked-in booking failed to block overlap: %v", err)
	}
	// A later reservation is valid but cannot be checked in yet.
	var futureID string
	err = pool.QueryRow(ctx, "INSERT INTO public.reservations(user_id,resource_id,starts_at,ends_at,status,check_in_code) VALUES($1,$2,clock_timestamp()+interval '2 hours',clock_timestamp()+interval '3 hours','CONFIRMED','0043') RETURNING id", owner, resource).Scan(&futureID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = store.CheckInReservation(ctx, futureID, owner, "0043"); !errors.Is(err, ErrCheckInWindow) {
		t.Fatalf("future check-in: %v", err)
	}
	// Independent resource avoids overlap while testing late and terminal states.
	var lateResource, lateID string
	err = pool.QueryRow(ctx, "INSERT INTO public.resources(name,type) VALUES('Late check-in desk','DESK') RETURNING id").Scan(&lateResource)
	if err != nil {
		t.Fatal(err)
	}
	err = pool.QueryRow(ctx, "INSERT INTO public.reservations(user_id,resource_id,starts_at,ends_at,status,check_in_code) VALUES($1,$2,clock_timestamp()-interval '20 minutes',clock_timestamp()+interval '40 minutes','CONFIRMED','0044') RETURNING id", owner, lateResource).Scan(&lateID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = store.CheckInReservation(ctx, lateID, owner, "0044"); !errors.Is(err, ErrCheckInWindow) {
		t.Fatalf("late check-in: %v", err)
	}
	for _, status := range []string{"CANCELLED", "COMPLETED", "PENDING"} {
		if _, err = pool.Exec(ctx, "UPDATE public.reservations SET status=$2 WHERE id=$1", lateID, status); err != nil {
			t.Fatal(err)
		}
		if _, err = store.CheckInReservation(ctx, lateID, owner, "0044"); !errors.Is(err, ErrCheckInStatus) {
			t.Fatalf("status %s: %v", status, err)
		}
	}

	t.Run("no-show expiration", func(t *testing.T) { testNoShowExpiration(t, pool, owner) })

	t.Run("no-show concurrency and rollback", func(t *testing.T) { testNoShowSafety(t, pool, owner) })

	t.Run("analytics", func(t *testing.T) { testAnalytics(t, pool, owner) })

	t.Run("user administration", func(t *testing.T) { testUserAdministration(t, pool) })

}
