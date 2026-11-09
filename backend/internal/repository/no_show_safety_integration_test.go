package repository

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nik003kk/DeskBooker/backend/internal/models"
)

// Only the test query replaces the clock expression; production always uses
// PostgreSQL's statement timestamp. This tests real SQL at exact boundaries.
func expireAt(ctx context.Context, db DB, now time.Time) (int64, error) {
	query := strings.ReplaceAll(expireNoShowsSQL, "statement_timestamp()", "$1::timestamptz")
	var count int64
	err := db.QueryRow(ctx, query, now).Scan(&count)
	return count, err
}

func auditCount(t *testing.T, pool *pgxpool.Pool, id string) int {
	t.Helper()
	var count int
	if err := pool.QueryRow(context.Background(), "SELECT count(*) FROM public.reservation_audit_log WHERE reservation_id=$1", id).Scan(&count); err != nil {
		t.Fatal(err)
	}
	return count
}

func testNoShowSafety(t *testing.T, pool *pgxpool.Pool, owner string) {
	ctx := context.Background()
	var now time.Time
	if err := pool.QueryRow(ctx, "SELECT clock_timestamp()").Scan(&now); err != nil {
		t.Fatal(err)
	}
	t.Run("exact inclusive deadline", func(t *testing.T) {
		booking := noShowFixture(t, pool, owner, now.Add(-15*time.Minute), "CONFIRMED")
		count, err := expireAt(ctx, pool, now)
		if err != nil || count != 0 {
			t.Fatalf("expired at inclusive check-in boundary: %d %v", count, err)
		}
		saved, err := New(pool).GetReservation(ctx, booking.ID)
		if err != nil || saved.Status != models.ReservationConfirmed || auditCount(t, pool, booking.ID) != 0 {
			t.Fatal("boundary changed reservation")
		}
		count, err = expireAt(ctx, pool, now.Add(time.Microsecond))
		if err != nil || count != 1 || auditCount(t, pool, booking.ID) != 1 {
			t.Fatalf("not expired after boundary: %d %v", count, err)
		}
	})
	t.Run("concurrent workers and repeated sweeps", func(t *testing.T) {
		fixtures := make([]models.Reservation, 4)
		for i := range fixtures {
			fixtures[i] = noShowFixture(t, pool, owner, now.Add(-20*time.Minute), "CONFIRMED")
		}
		counts := make(chan int64, 8)
		failures := make(chan error, 8)
		var wg sync.WaitGroup
		for i := 0; i < 8; i++ {
			wg.Add(1)
			go func() { defer wg.Done(); n, err := New(pool).ExpireNoShows(ctx); counts <- n; failures <- err }()
		}
		wg.Wait()
		close(counts)
		close(failures)
		for err := range failures {
			if err != nil {
				t.Fatal(err)
			}
		}
		var total int64
		for n := range counts {
			total += n
		}
		if total != 4 {
			t.Fatalf("expected four transitions across all workers, got %d", total)
		}
		for _, booking := range fixtures {
			saved, err := New(pool).GetReservation(ctx, booking.ID)
			if err != nil || saved.Status != models.ReservationExpired || auditCount(t, pool, booking.ID) != 1 {
				t.Fatalf("duplicate or missing transition: %v", err)
			}
		}
		for i := 0; i < 3; i++ {
			if n, err := New(pool).ExpireNoShows(ctx); err != nil || n != 0 {
				t.Fatalf("repeat processed rows: %d %v", n, err)
			}
		}
	})
	t.Run("audit failure rolls back expiration", func(t *testing.T) {
		booking := noShowFixture(t, pool, owner, now.Add(-20*time.Minute), "CONFIRMED")
		_, err := pool.Exec(ctx, `CREATE FUNCTION public.test_reject_no_show_audit() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RAISE EXCEPTION 'forced audit write failure'; END $$;
CREATE TRIGGER test_reject_no_show_audit BEFORE INSERT ON public.reservation_audit_log FOR EACH ROW EXECUTE FUNCTION public.test_reject_no_show_audit();`)
		if err != nil {
			t.Fatal(err)
		}
		defer pool.Exec(ctx, "DROP TRIGGER IF EXISTS test_reject_no_show_audit ON public.reservation_audit_log; DROP FUNCTION IF EXISTS public.test_reject_no_show_audit();")
		if _, err = New(pool).ExpireNoShows(ctx); err == nil {
			t.Fatal("audit failure must fail the whole sweep")
		}
		saved, err := New(pool).GetReservation(ctx, booking.ID)
		if err != nil || saved.Status != models.ReservationConfirmed || auditCount(t, pool, booking.ID) != 0 {
			t.Fatal("audit failure left partial expiration")
		}
		if _, err = pool.Exec(ctx, "DROP TRIGGER test_reject_no_show_audit ON public.reservation_audit_log; DROP FUNCTION public.test_reject_no_show_audit();"); err != nil {
			t.Fatal(err)
		}
		if n, err := New(pool).ExpireNoShows(ctx); err != nil || n != 1 || auditCount(t, pool, booking.ID) != 1 {
			t.Fatalf("retry after rollback: %d %v", n, err)
		}
	})
	t.Run("check-in wins while expiration skips locked row", func(t *testing.T) {
		booking := noShowFixture(t, pool, owner, now.Add(-14*time.Minute), "CONFIRMED")
		if _, err := pool.Exec(ctx, "UPDATE public.reservations SET check_in_code='7812' WHERE id=$1", booking.ID); err != nil {
			t.Fatal(err)
		}
		ready := make(chan struct{})
		release := make(chan struct{})
		defer close(release)
		starter := pausedCommitPool{pool: pool, ready: ready, release: release}
		result := make(chan error, 1)
		go func() {
			_, err := NewBookingStore(starter).CheckInReservation(ctx, booking.ID, owner, "7812")
			result <- err
		}()
		select {
		case <-ready:
		case err := <-result:
			t.Fatalf("check-in failed before commit: %v", err)
		case <-time.After(5 * time.Second):
			t.Fatal("check-in did not acquire lock")
		}
		if n, err := expireAt(ctx, pool, now.Add(2*time.Minute)); err != nil || n != 0 {
			t.Fatalf("locked check-in expired: %d %v", n, err)
		}
		release <- struct{}{}
		select {
		case err := <-result:
			if err != nil {
				t.Fatal(err)
			}
		case <-time.After(5 * time.Second):
			t.Fatal("check-in did not finish")
		}
		if n, err := expireAt(ctx, pool, now.Add(2*time.Minute)); err != nil || n != 0 {
			t.Fatalf("committed check-in expired: %d %v", n, err)
		}
		saved, err := New(pool).GetReservation(ctx, booking.ID)
		if err != nil || saved.Status != models.ReservationCheckedIn || saved.CheckedInAt == nil || auditCount(t, pool, booking.ID) != 0 {
			t.Fatal("check-in lost to expiration")
		}
	})
	t.Run("expiration wins before waiting check-in", func(t *testing.T) {
		booking := noShowFixture(t, pool, owner, now.Add(-20*time.Minute), "CONFIRMED")
		if _, err := pool.Exec(ctx, "UPDATE public.reservations SET check_in_code='7813' WHERE id=$1", booking.ID); err != nil {
			t.Fatal(err)
		}
		tx, err := pool.Begin(ctx)
		if err != nil {
			t.Fatal(err)
		}
		defer tx.Rollback(ctx)
		if n, err := New(tx).ExpireNoShows(ctx); err != nil || n != 1 {
			t.Fatalf("expiration transaction: %d %v", n, err)
		}
		attemptCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		ready := make(chan struct{})
		result := make(chan error, 1)
		go func() {
			_, err := NewBookingStore(announcedPool{pool: pool, ready: ready}).CheckInReservation(attemptCtx, booking.ID, owner, "7813")
			result <- err
		}()
		select {
		case <-ready:
		case <-attemptCtx.Done():
			t.Fatal("check-in did not start")
		}
		if err := tx.Commit(ctx); err != nil {
			t.Fatal(err)
		}
		select {
		case err := <-result:
			if !errors.Is(err, ErrCheckInStatus) {
				t.Fatalf("expired check-in returned %v", err)
			}
		case <-attemptCtx.Done():
			t.Fatal("check-in stayed blocked")
		}
		saved, err := New(pool).GetReservation(ctx, booking.ID)
		if err != nil || saved.Status != models.ReservationExpired || saved.CheckedInAt != nil || auditCount(t, pool, booking.ID) != 1 {
			t.Fatal("expired reservation revived")
		}
	})
}

type pausedCommitPool struct {
	pool    *pgxpool.Pool
	ready   chan struct{}
	release chan struct{}
}

func (p pausedCommitPool) Begin(ctx context.Context) (pgx.Tx, error) {
	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	return pausedCommitTx{Tx: tx, ready: p.ready, release: p.release}, nil
}

type pausedCommitTx struct {
	pgx.Tx
	ready   chan struct{}
	release chan struct{}
}

func (tx pausedCommitTx) Commit(ctx context.Context) error {
	close(tx.ready)
	select {
	case <-tx.release:
	case <-ctx.Done():
		return ctx.Err()
	}
	return tx.Tx.Commit(ctx)
}

type announcedPool struct {
	pool  *pgxpool.Pool
	ready chan struct{}
}

func (p announcedPool) Begin(ctx context.Context) (pgx.Tx, error) {
	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	return announcedTx{Tx: tx, ready: p.ready, once: &sync.Once{}}, nil
}

type announcedTx struct {
	pgx.Tx
	ready chan struct{}
	once  *sync.Once
}

func (tx announcedTx) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	if strings.Contains(sql, "FOR UPDATE") {
		tx.once.Do(func() { close(tx.ready) })
	}
	return tx.Tx.QueryRow(ctx, sql, args...)
}
