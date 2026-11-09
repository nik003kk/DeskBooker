package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/nik003kk/DeskBooker/backend/internal/models"
)

var (
	ErrCheckInPIN    = errors.New("incorrect check-in PIN")
	ErrCheckInWindow = errors.New("outside check-in window")
	ErrCheckInStatus = errors.New("reservation cannot be checked in")
)

// The owner predicate hides other users' reservations. The row lock serializes
// repeated check-ins and future cancellation/expiration updates.
func (s *BookingStore) CheckInReservation(ctx context.Context, id, userID, pin string) (models.Reservation, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return models.Reservation{}, fmt.Errorf("begin check-in: %w", err)
	}
	defer tx.Rollback(ctx)
	saved, err := scanReservation(tx.QueryRow(ctx,
		"SELECT "+reservationColumns+" FROM public.reservations WHERE id=$1 AND user_id=$2 FOR UPDATE", id, userID))
	if err != nil {
		return models.Reservation{}, fmt.Errorf("find owned reservation: %w", err)
	}
	if saved.CheckInCode == nil || *saved.CheckInCode != pin {
		return models.Reservation{}, ErrCheckInPIN
	}
	// A retried successful request returns the original timestamp unchanged.
	if saved.Status == models.ReservationCheckedIn {
		return saved, nil
	}
	if saved.Status != models.ReservationConfirmed {
		return models.Reservation{}, ErrCheckInStatus
	}
	var now time.Time
	// clock_timestamp is evaluated after acquiring the row lock, not at BEGIN.
	if err = tx.QueryRow(ctx, "SELECT clock_timestamp()").Scan(&now); err != nil {
		return models.Reservation{}, fmt.Errorf("read check-in clock: %w", err)
	}
	if now.Before(saved.StartsAt.Add(-10*time.Minute)) || now.After(saved.StartsAt.Add(15*time.Minute)) || !now.Before(saved.EndsAt) {
		return models.Reservation{}, ErrCheckInWindow
	}
	saved, err = scanReservation(tx.QueryRow(ctx,
		"UPDATE public.reservations SET status='CHECKED_IN', checked_in_at=$3 WHERE id=$1 AND user_id=$2 AND status='CONFIRMED' RETURNING "+reservationColumns,
		id, userID, now))
	if err != nil {
		return models.Reservation{}, fmt.Errorf("update check-in: %w", err)
	}
	if err = tx.Commit(ctx); err != nil {
		return models.Reservation{}, fmt.Errorf("commit check-in: %w", err)
	}
	return saved, nil
}
