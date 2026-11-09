package repository

import (
	"context"
	"fmt"
)

// SKIP LOCKED lets multiple instances cooperate without waiting for check-in.
// A single statement atomically selects, rechecks, and expires at most 500 rows.
// Strictly older than 15 minutes preserves the inclusive check-in deadline.
const expireNoShowsSQL = `
WITH candidates AS (
    SELECT id FROM public.reservations
    WHERE status='CONFIRMED' AND checked_in_at IS NULL
      AND starts_at < statement_timestamp() - interval '15 minutes'
    ORDER BY starts_at, id
    LIMIT 500
    FOR UPDATE SKIP LOCKED
), expired AS (
    UPDATE public.reservations r SET status='EXPIRED'
    FROM candidates c
    WHERE r.id=c.id AND r.status='CONFIRMED' AND r.checked_in_at IS NULL
    RETURNING r.id
)
SELECT count(*) FROM expired`

func (s *Store) ExpireNoShows(ctx context.Context) (int64, error) {
	var count int64
	if err := s.db.QueryRow(ctx, expireNoShowsSQL).Scan(&count); err != nil {
		return 0, fmt.Errorf("expire no-shows: %w", err)
	}
	return count, nil
}
