package analytics

import (
	"context"
	"encoding/csv"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/nik003kk/DeskBooker/backend/internal/models"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestAnalyticsCSVMatchesAggregateFields(t *testing.T) {
	percent := 25.0
	report := models.Analytics{From: "2030-01-01", To: "2030-01-07", Timezone: "Europe/Ljubljana", GeneratedAt: time.Date(2030, 1, 7, 12, 0, 0, 0, time.UTC), TotalReservations: 4, ActiveReservations: 2, OccupancyRate: &percent, StatusCounts: map[string]int64{"EXPIRED": 1, "CHECKED_IN": 3}, NoShow: models.NoShowStatistics{Expired: 1, Attended: 3, Eligible: 4, Rate: &percent}, Daily: []models.OccupancyPoint{{Date: "2030-01-01", Weekday: 2, BookedMinutes: 120, CapacityMinutes: 480, OccupancyRate: &percent}}}
	raw, err := analyticsCSV(report)
	if err != nil {
		t.Fatal(err)
	}
	rows, err := csv.NewReader(strings.NewReader(strings.TrimPrefix(string(raw), "\uFEFF"))).ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	values := map[string]string{}
	for _, row := range rows[1:] {
		if len(row) != 6 {
			t.Fatal(row)
		}
		values[row[0]+"/"+row[1]+"/"+row[2]] = row[4]
	}
	for key, want := range map[string]string{"summary/total_reservations/": "4", "summary/no_show_rate/": "25", "status/EXPIRED/": "1", "daily/occupancy_rate/2030-01-01": "25", "metadata/timezone/": "Europe/Ljubljana"} {
		if values[key] != want {
			t.Fatalf("%s=%q want %q", key, values[key], want)
		}
	}
	report.NoShow.Rate = nil
	raw, err = analyticsCSV(report)
	if err != nil || strings.Contains(string(raw), "NaN") || strings.Contains(string(raw), "<nil>") {
		t.Fatal("invalid missing percentage")
	}
}

type csvStore struct{ err error }

func (s csvStore) ReservationAnalytics(_ context.Context, p models.AnalyticsRange) (models.Analytics, error) {
	return models.Analytics{From: p.From.Format(time.DateOnly), To: p.Until.AddDate(0, 0, -1).Format(time.DateOnly), StatusCounts: map[string]int64{}}, s.err
}
func TestCSVHandlerHeadersAndErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tc := range []struct {
		query  string
		err    error
		status int
	}{{"?from=2030-01-01&to=2030-01-07", nil, 200}, {"?from=bad", nil, 400}, {"", errors.New("private DB secret"), 500}} {
		router := gin.New()
		router.GET("/analytics/export.csv", CSVHandler(csvStore{tc.err}))
		r := httptest.NewRecorder()
		router.ServeHTTP(r, httptest.NewRequest("GET", "/analytics/export.csv"+tc.query, nil))
		if r.Code != tc.status || strings.Contains(r.Body.String(), "private DB") || r.Header().Get("Cache-Control") != "no-store" {
			t.Fatal(r.Code, r.Body.String())
		}
		if tc.status == 200 {
			if r.Header().Get("Content-Type") != "text/csv; charset=utf-8" || !strings.Contains(r.Header().Get("Content-Disposition"), "2030-01-01-2030-01-07.csv") {
				t.Fatal(r.Header())
			}
		} else if r.Header().Get("Content-Disposition") != "" {
			t.Fatal("error became CSV attachment")
		}
	}
}
