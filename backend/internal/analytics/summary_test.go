package analytics

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/nik003kk/DeskBooker/backend/internal/models"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestAnalyticsPeriodCalendarBoundaries(t *testing.T) {
	now := time.Date(2026, 11, 9, 0, 0, 0, 0, time.UTC)
	p, err := parseRange(url.Values{}, now)
	if err != nil || p.From.Format(time.DateOnly) != "2026-11-03" || p.Until.Format(time.DateOnly) != "2026-11-10" {
		t.Fatal(p, err)
	}
	for _, tc := range []struct {
		date  string
		hours time.Duration
	}{{"2030-03-31", 23 * time.Hour}, {"2030-10-27", 25 * time.Hour}} {
		p, err := parseRange(url.Values{"from": {tc.date}, "to": {tc.date}}, now)
		if err != nil || p.Until.Sub(p.From) != tc.hours {
			t.Fatalf("DST range %s: %v %v", tc.date, p, err)
		}
	}
	for _, query := range []string{"from=2030-01-01", "from=bad&to=bad", "from=2030-02-30&to=2030-03-01", "from=2030-01-02&to=2030-01-01", "from=2028-01-01&to=2029-01-01", "from=2030-01-01&from=2030-01-02&to=2030-01-02", "extra=x", "from=&to="} {
		q, _ := url.ParseQuery(query)
		if _, err := parseRange(q, now); err == nil {
			t.Errorf("accepted %s", query)
		}
	}
}

type fakeAnalytics struct {
	calls int
	err   error
}

func (s *fakeAnalytics) ReservationAnalytics(context.Context, models.AnalyticsRange) (models.Analytics, error) {
	s.calls++
	return models.Analytics{TotalReservations: 2, StatusCounts: map[string]int64{"CONFIRMED": 2}}, s.err
}
func TestSummaryHandlerValidationAndSanitizedFailure(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tc := range []struct {
		query         string
		err           error
		status, calls int
	}{{"", nil, 200, 1}, {"?from=bad", nil, 400, 0}, {"", errors.New("private connection secret"), 500, 1}} {
		store := &fakeAnalytics{err: tc.err}
		router := gin.New()
		router.GET("/analytics", SummaryHandler(store))
		r := httptest.NewRecorder()
		router.ServeHTTP(r, httptest.NewRequest("GET", "/analytics"+tc.query, nil))
		if r.Code != tc.status || store.calls != tc.calls || strings.Contains(r.Body.String(), "private connection") || r.Header().Get("Cache-Control") != "no-store" {
			t.Fatal(r.Code, r.Body.String())
		}
	}
}
