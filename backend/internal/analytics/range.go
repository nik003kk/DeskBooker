package analytics

import (
	"errors"
	"github.com/nik003kk/DeskBooker/backend/internal/models"
	"net/url"
	"time"
)

var errRange = errors.New("expected from/to dates in YYYY-MM-DD, in order, at most 366 days")

// Date limits are inclusive in Europe/Ljubljana. AddDate preserves local
// midnight through DST; duration in hours is not a calendar-day count.
func parseRange(query url.Values, now time.Time) (models.AnalyticsRange, error) {
	for key, values := range query {
		if (key != "from" && key != "to") || len(values) != 1 || values[0] == "" {
			return models.AnalyticsRange{}, errRange
		}
	}
	loc, err := time.LoadLocation("Europe/Ljubljana")
	if err != nil {
		return models.AnalyticsRange{}, err
	}
	from, to := query.Get("from"), query.Get("to")
	if from == "" && to == "" {
		today := now.In(loc)
		to = today.Format(time.DateOnly)
		from = today.AddDate(0, 0, -6).Format(time.DateOnly)
	}
	start, err := time.ParseInLocation(time.DateOnly, from, loc)
	if err != nil || start.Format(time.DateOnly) != from {
		return models.AnalyticsRange{}, errRange
	}
	last, err := time.ParseInLocation(time.DateOnly, to, loc)
	if err != nil || last.Format(time.DateOnly) != to || last.Before(start) {
		return models.AnalyticsRange{}, errRange
	}
	civilFrom, _ := time.Parse(time.DateOnly, from)
	civilTo, _ := time.Parse(time.DateOnly, to)
	if civilTo.Sub(civilFrom) > 365*24*time.Hour {
		return models.AnalyticsRange{}, errRange
	}
	return models.AnalyticsRange{From: start, Until: last.AddDate(0, 0, 1)}, nil
}
