package server

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/nik003kk/DeskBooker/backend/internal/analytics"
	"github.com/nik003kk/DeskBooker/backend/internal/auth"
	"github.com/nik003kk/DeskBooker/backend/internal/models"
	"net/http/httptest"
	"strings"
	"testing"
)

type analyticsAccessStore struct {
	user  models.User
	err   error
	calls int
}

func (s *analyticsAccessStore) GetUser(context.Context, string) (models.User, error) {
	return s.user, s.err
}
func (s *analyticsAccessStore) ReservationAnalytics(_ context.Context, p models.AnalyticsRange) (models.Analytics, error) {
	s.calls++
	return models.Analytics{From: p.From.Format("2006-01-02"), To: p.Until.AddDate(0, 0, -1).Format("2006-01-02"), StatusCounts: map[string]int64{}}, nil
}
func TestAnalyticsAndCSVRequireLiveActiveManagerRole(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tokens, err := auth.NewJWTService("analytics-access-test-secret-at-least-32-bytes")
	if err != nil {
		t.Fatal(err)
	}
	token, _, err := tokens.Issue("analytics-user")
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"/api/analytics", "/api/analytics/export.csv"} {
		for _, tc := range []struct {
			name, role, status string
			lookupErr          error
			authorization      string
			code               int
		}{
			{"admin", "admin", "active", nil, "Bearer " + token, 200},
			{"manager", "manager", "active", nil, "Bearer " + token, 200},
			{"employee", "employee", "active", nil, "Bearer " + token, 403},
			{"inactive manager", "manager", "inactive", nil, "Bearer " + token, 403},
			{"suspended admin", "admin", "suspended", nil, "Bearer " + token, 403},
			{"unknown role", "OFFICE_MANAGER", "active", nil, "Bearer " + token, 403},
			{"deleted user", "", "", pgx.ErrNoRows, "Bearer " + token, 401},
			{"lookup failure", "", "", errors.New("private database secret"), "Bearer " + token, 500},
			{"no token", "admin", "active", nil, "", 401},
			{"invalid token", "admin", "active", nil, "Bearer invalid", 401},
		} {
			t.Run(path+"/"+tc.name, func(t *testing.T) {
				store := &analyticsAccessStore{user: models.User{ID: "analytics-user", Role: models.UserRole(tc.role), Status: models.UserStatus(tc.status)}, err: tc.lookupErr}
				router := gin.New()
				RegisterAnalytics(router, auth.RequireJWT(tokens), auth.RequireAnalyticsManager(store), analytics.SummaryHandler(store), analytics.CSVHandler(store))
				request := httptest.NewRequest("GET", path+"?from=2030-01-01&to=2030-01-07", nil)
				request.Header.Set("Authorization", tc.authorization)
				response := httptest.NewRecorder()
				router.ServeHTTP(response, request)
				if response.Code != tc.code || strings.Contains(response.Body.String(), "private database") {
					t.Fatal(response.Code, response.Body.String())
				}
				if tc.code == 200 && store.calls != 1 || tc.code != 200 && store.calls != 0 {
					t.Fatalf("unauthorized aggregate query: %d", store.calls)
				}
				if tc.code != 200 && response.Header().Get("Content-Disposition") != "" {
					t.Fatal("unauthorized CSV download")
				}
			})
		}
	}
	store := &analyticsAccessStore{user: models.User{Role: models.RoleManager, Status: models.UserActive}}
	router := gin.New()
	RegisterAnalytics(router, auth.RequireJWT(tokens), auth.RequireAnalyticsManager(store), analytics.SummaryHandler(store), analytics.CSVHandler(store))
	for _, code := range []int{200, 403} {
		r := httptest.NewRecorder()
		request := httptest.NewRequest("GET", "/api/analytics", nil)
		request.Header.Set("Authorization", "Bearer "+token)
		router.ServeHTTP(r, request)
		if r.Code != code {
			t.Fatal("stale role accepted", r.Code)
		}
		store.user.Role = models.RoleEmployee
	}
	if store.calls != 1 {
		t.Fatal("role downgrade still queried analytics")
	}
}
