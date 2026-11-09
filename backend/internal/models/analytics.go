package models

import "time"

type AnalyticsRange struct{ From, Until time.Time }

type Analytics struct {
	From               string           `json:"from"`
	To                 string           `json:"to"`
	Timezone           string           `json:"timezone"`
	GeneratedAt        time.Time        `json:"generated_at"`
	TotalReservations  int64            `json:"total_reservations"`
	ActiveReservations int64            `json:"active_reservations"`
	StatusCounts       map[string]int64 `json:"status_counts"`
}
