package repository

import "context"

// CountNoShows supplies a read-only sweep while expiration is introduced in
// DB-064. Only CONFIRMED bookings without attendance qualify after 15 minutes.
func (s *Store) CountNoShows(ctx context.Context) (int64, error) {
	var count int64
	err := s.db.QueryRow(ctx, "SELECT count(*) FROM public.reservations WHERE status='CONFIRMED' AND checked_in_at IS NULL AND starts_at < statement_timestamp() - interval '15 minutes'").Scan(&count)
	return count, err
}
