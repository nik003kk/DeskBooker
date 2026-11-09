package repository

import "time"

const (
	checkInBeforeStart = 10 * time.Minute
	checkInAfterStart  = 15 * time.Minute
)

// Both edges of the 25-minute window are inclusive. A short reservation
// closes check-in at its end. All comparisons use instants, including DST.
func checkInWindowOpen(now, start, end time.Time) bool {
	return !start.IsZero() && end.After(start) &&
		!now.Before(start.Add(-checkInBeforeStart)) &&
		!now.After(start.Add(checkInAfterStart)) && now.Before(end)
}
