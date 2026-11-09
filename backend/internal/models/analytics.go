package models

import (
	"math"
	"time"
)

type AnalyticsRange struct{ From, Until time.Time }

type Analytics struct {
	Daily              []OccupancyPoint `json:"daily"`
	Weekdays           []OccupancyPoint `json:"weekdays"`
	OccupancyRate      *float64         `json:"occupancy_rate"`
	From               string           `json:"from"`
	To                 string           `json:"to"`
	Timezone           string           `json:"timezone"`
	GeneratedAt        time.Time        `json:"generated_at"`
	TotalReservations  int64            `json:"total_reservations"`
	ActiveReservations int64            `json:"active_reservations"`
	StatusCounts       map[string]int64 `json:"status_counts"`
}

type OccupancyPoint struct {
	Date            string   `json:"date,omitempty"`
	Weekday         int      `json:"weekday"`
	BookedMinutes   float64  `json:"booked_minutes"`
	CapacityMinutes float64  `json:"capacity_minutes"`
	OccupancyRate   *float64 `json:"occupancy_rate"`
}

// Percent is nil when no denominator exists; absent capacity/outcomes must
// not be presented as an observed zero percent.
func AnalyticsPercent(numerator, denominator float64) *float64 {
	if denominator <= 0 {
		return nil
	}
	value := math.Round(numerator/denominator*10000) / 100
	return &value
}
