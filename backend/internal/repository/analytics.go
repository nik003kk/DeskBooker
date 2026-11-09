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
), days AS (
 SELECT g.day::date AS day,extract(isodow FROM g.day)::int AS weekday,
 (g.day::date+time '09:00') AT TIME ZONE 'Europe/Ljubljana' AS opens,
 (g.day::date+time '17:00') AT TIME ZONE 'Europe/Ljubljana' AS closes
 FROM generate_series(($1::timestamptz AT TIME ZONE 'Europe/Ljubljana')::date::timestamp,
 (($2::timestamptz AT TIME ZONE 'Europe/Ljubljana')::date-1)::timestamp,interval '1 day') g(day)
 WHERE extract(isodow FROM g.day) BETWEEN 1 AND 5
), capacity AS (
 SELECT d.day,sum(extract(epoch FROM d.closes-greatest(d.opens,r.created_at))) / 60.0 AS minutes
 FROM days d JOIN public.resources r ON r.is_active AND r.created_at<d.closes
 GROUP BY d.day
), booked_ranges AS (
 SELECT d.day,r.id,range_agg(tstzrange(greatest(b.starts_at,d.opens,r.created_at),least(b.ends_at,d.closes),'[)')) AS spans
 FROM days d JOIN public.resources r ON r.is_active AND r.created_at<d.closes
 JOIN public.reservations b ON b.resource_id=r.id AND b.status IN ('CONFIRMED','CHECKED_IN','COMPLETED')
 AND b.starts_at<d.closes AND b.ends_at>greatest(d.opens,r.created_at)
 GROUP BY d.day,r.id
), booked AS (
 SELECT day,sum(extract(epoch FROM upper(span)-lower(span))) / 60.0 AS minutes
 FROM booked_ranges CROSS JOIN LATERAL unnest(spans) span GROUP BY day
), daily AS (
 SELECT d.day,d.weekday,coalesce(b.minutes,0) AS booked_minutes,coalesce(c.minutes,0) AS capacity_minutes
 FROM days d LEFT JOIN booked b USING(day) LEFT JOIN capacity c USING(day)
), weekdays AS (
 SELECT w.weekday,coalesce(sum(d.booked_minutes),0) AS booked_minutes,coalesce(sum(d.capacity_minutes),0) AS capacity_minutes
 FROM generate_series(1,5) w(weekday) LEFT JOIN daily d USING(weekday) GROUP BY w.weekday
)
SELECT jsonb_build_object(
 'generated_at',statement_timestamp(),'total_reservations',summary.total,
 'active_reservations',summary.active,'status_counts',status_counts.counts,
 'daily',(SELECT coalesce(jsonb_agg(jsonb_build_object('date',day,'weekday',weekday,'booked_minutes',booked_minutes,'capacity_minutes',capacity_minutes) ORDER BY day),'[]'::jsonb) FROM daily),
 'weekdays',(SELECT jsonb_agg(jsonb_build_object('weekday',weekday,'booked_minutes',booked_minutes,'capacity_minutes',capacity_minutes) ORDER BY weekday) FROM weekdays))
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
	var booked, capacity float64
	for i := range report.Daily {
		point := &report.Daily[i]
		point.OccupancyRate = models.AnalyticsPercent(point.BookedMinutes, point.CapacityMinutes)
		booked += point.BookedMinutes
		capacity += point.CapacityMinutes
	}
	for i := range report.Weekdays {
		point := &report.Weekdays[i]
		point.OccupancyRate = models.AnalyticsPercent(point.BookedMinutes, point.CapacityMinutes)
	}
	report.OccupancyRate = models.AnalyticsPercent(booked, capacity)
	report.Timezone = "Europe/Ljubljana"
	return report, nil
}
