package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/nik003kk/DeskBooker/backend/internal/models"
)

// One SQL statement gives all aggregates the same MVCC snapshot.
const analyticsSQL = `
WITH scoped AS (
 SELECT status,ends_at FROM public.reservations WHERE starts_at >= $1 AND starts_at < $2
), summary AS (
 SELECT count(*) AS total,
 count(*) FILTER (WHERE status IN ('CONFIRMED','CHECKED_IN') AND ends_at > statement_timestamp()) AS active
 FROM scoped
), status_counts AS (
 SELECT coalesce(jsonb_object_agg(status,n),'{}'::jsonb) AS counts FROM (
 SELECT s.status,count(b.status) AS n
 FROM (VALUES ('PENDING'),('CONFIRMED'),('CHECKED_IN'),('COMPLETED'),('CANCELLED'),('EXPIRED')) s(status)
 LEFT JOIN scoped b ON b.status=s.status GROUP BY s.status
 ) t
)
SELECT jsonb_build_object(
 'generated_at',statement_timestamp(),'total_reservations',summary.total,
 'active_reservations',summary.active,'status_counts',status_counts.counts)
FROM summary CROSS JOIN status_counts`

func (s *Store) ReservationAnalytics(ctx context.Context, period models.AnalyticsRange) (models.Analytics, error) {
	var raw []byte
	if err := s.db.QueryRow(ctx, analyticsSQL, period.From, period.Until).Scan(&raw); err != nil {
		return models.Analytics{}, fmt.Errorf("read reservation analytics: %w", err)
	}
	var report models.Analytics
	if err := json.Unmarshal(raw, &report); err != nil {
		return models.Analytics{}, fmt.Errorf("decode reservation analytics: %w", err)
	}
	report.From = period.From.Format("2006-01-02")
	report.To = period.Until.AddDate(0, 0, -1).Format("2006-01-02")
	report.Timezone = "Europe/Ljubljana"
	return report, nil
}
