package auth

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/nik003kk/DeskBooker/backend/internal/models"
	"net/http/httptest"
	"strings"
	"testing"
)

type liveAccount struct {
	user  models.User
	err   error
	calls int
}

func (s *liveAccount) GetUser(context.Context, string) (models.User, error) {
	s.calls++
	return s.user, s.err
}
func TestCurrentJWTReloadsRightsOnEveryProtectedRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tokens, err := NewJWTService("live-user-rights-test-secret-at-least-32-bytes")
	if err != nil {
		t.Fatal(err)
	}
	token, _, _ := tokens.Issue("current-user")
	store := &liveAccount{user: models.User{ID: "current-user", Role: models.RoleAdmin, Status: models.UserActive, PasswordHash: "sensitive-hash"}}
	router := gin.New()
	router.GET("/admin", RequireCurrentJWT(tokens, store), RequireAdministrator(store), MeHandler(store))
	router.GET("/private", RequireCurrentJWT(tokens, store), MeHandler(store))
	call := func(path string, want int) {
		t.Helper()
		store.calls = 0
		request := httptest.NewRequest("GET", path, nil)
		request.Header.Set("Authorization", "Bearer "+token)
		r := httptest.NewRecorder()
		router.ServeHTTP(r, request)
		if r.Code != want || store.calls != 1 || strings.Contains(r.Body.String(), "sensitive-hash") || strings.Contains(r.Body.String(), "private secret") {
			t.Fatalf("%s: %d calls=%d %s", path, r.Code, store.calls, r.Body.String())
		}
	}
	call("/admin", 200)
	store.user.Role = models.RoleEmployee
	call("/admin", 403)
	call("/private", 200)
	store.user.Role = models.RoleAdmin
	call("/admin", 200)
	store.user.Status = models.UserSuspended
	call("/private", 403)
	store.err = pgx.ErrNoRows
	call("/private", 401)
	store.err = errors.New("private secret")
	call("/private", 500)
}
func TestCurrentJWTDoesNotLookupInvalidTokensAndFailsClosedWithoutStore(t *testing.T) {
	tokens, _ := NewJWTService("live-user-rights-test-secret-at-least-32-bytes")
	store := &liveAccount{}
	router := gin.New()
	router.GET("/private", RequireCurrentJWT(tokens, store), func(c *gin.Context) { t.Error("invalid token reached handler") })
	r := httptest.NewRecorder()
	router.ServeHTTP(r, httptest.NewRequest("GET", "/private", nil))
	if r.Code != 401 || store.calls != 0 {
		t.Fatal(r.Code, store.calls)
	}
	router = gin.New()
	router.GET("/private", RequireCurrentJWT(tokens, nil))
	r = httptest.NewRecorder()
	router.ServeHTTP(r, httptest.NewRequest("GET", "/private", nil))
	if r.Code != 503 {
		t.Fatal("nil account store granted access")
	}
}
