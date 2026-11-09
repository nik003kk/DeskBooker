package reservations

import (
	"context"
	"errors"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/nik003kk/DeskBooker/backend/internal/models"
	"github.com/nik003kk/DeskBooker/backend/internal/repository"
)

type fakeCheckIn struct {
	calls     int
	user, pin string
	err       error
}

func (s *fakeCheckIn) CheckInReservation(_ context.Context, id, user, pin string) (models.Reservation, error) {
	s.calls++
	s.user = user
	s.pin = pin
	at := time.Now()
	return models.Reservation{ID: id, UserID: user, Status: models.ReservationCheckedIn, CheckedInAt: &at, CheckInCode: &pin}, s.err
}
func checkInRequest(store CheckInStore, body string, authenticated bool) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/reservations/:id/check-in", func(c *gin.Context) {
		if authenticated {
			c.Set("auth_user_id", "owner")
		}
	}, CheckInHandler(store))
	req := httptest.NewRequest("POST", "/reservations/"+resourceID+"/check-in", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, req)
	return response
}
func TestCheckInHandlerTrustsJWTAndNeverReturnsPIN(t *testing.T) {
	store := &fakeCheckIn{}
	r := checkInRequest(store, `{"pin":"0042"}`, true)
	if r.Code != 200 || store.user != "owner" || store.pin != "0042" || !strings.Contains(r.Body.String(), `"checked_in_at"`) || strings.Contains(r.Body.String(), "0042") {
		t.Fatalf("unexpected check-in: %d %s", r.Code, r.Body.String())
	}
	if r.Header().Get("Cache-Control") != "no-store" {
		t.Fatal("missing no-store")
	}
	r = checkInRequest(store, `{"pin":"0042"}`, false)
	if r.Code != 401 || store.calls != 1 {
		t.Fatal("unauthenticated request reached store")
	}
}
func TestCheckInHandlerMapsErrors(t *testing.T) {
	for _, tc := range []struct {
		err    error
		status int
		code   string
	}{
		{pgx.ErrNoRows, 404, "reservation_not_found"},
		{repository.ErrCheckInPIN, 422, "incorrect_pin"},
		{repository.ErrCheckInWindow, 409, "check_in_window"},
		{repository.ErrCheckInStatus, 409, "check_in_status"},
		{errors.New("private database details"), 500, "check_in_failed"},
	} {
		r := checkInRequest(&fakeCheckIn{err: tc.err}, `{"pin":"0042"}`, true)
		if r.Code != tc.status || !strings.Contains(r.Body.String(), tc.code) || strings.Contains(r.Body.String(), "private database") {
			t.Fatal(r.Code, r.Body.String())
		}
	}
}

type fakeDetail struct {
	owner string
	err   error
}

func (s fakeDetail) GetReservation(context.Context, string) (models.Reservation, error) {
	pin := "0042"
	return models.Reservation{UserID: s.owner, CheckInCode: &pin}, s.err
}
func (s fakeDetail) GetResource(context.Context, string) (models.Resource, error) {
	return models.Resource{Name: "Desk"}, nil
}
func TestReservationDetailHidesOtherOwnersAndPIN(t *testing.T) {
	for _, tc := range []struct {
		owner  string
		err    error
		status int
	}{{"owner", nil, 200}, {"someone-else", nil, 404}, {"", pgx.ErrNoRows, 404}} {
		router := gin.New()
		router.GET("/reservations/:id", func(c *gin.Context) { c.Set("auth_user_id", "owner") }, DetailHandler(fakeDetail{tc.owner, tc.err}))
		r := httptest.NewRecorder()
		router.ServeHTTP(r, httptest.NewRequest("GET", "/reservations/"+resourceID, nil))
		if r.Code != tc.status || strings.Contains(r.Body.String(), "0042") {
			t.Fatal(r.Code, r.Body.String())
		}
	}
}

func TestCheckInRejectsMalformedAndInjectedPINRequests(t *testing.T) {
	for _, body := range []string{`{}`, `{"pin":42}`, `{"pin":null}`, `{"pin":"123"}`, `{"pin":"12345"}`, `{"pin":"12a4"}`, `{"pin":"１２３４"}`, `{"pin":"0042","user_id":"other"}`, `{"pin":"0042"} {}`} {
		store := &fakeCheckIn{}
		r := checkInRequest(store, body, true)
		if r.Code != 400 || store.calls != 0 {
			t.Fatalf("invalid request reached store: %s: %d", body, r.Code)
		}
	}
}
