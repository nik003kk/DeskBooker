package adminusers

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/nik003kk/DeskBooker/backend/internal/auth"
	"github.com/nik003kk/DeskBooker/backend/internal/models"
	"net/http/httptest"
	"strings"
	"testing"
)

type listStore struct {
	role             models.UserRole
	status           models.UserStatus
	userErr, listErr error
	calls            int
	limit, offset    int
}

func (s *listStore) GetUser(context.Context, string) (models.User, error) {
	return models.User{ID: "admin", Role: s.role, Status: s.status}, s.userErr
}
func (s *listStore) ListAdminUsers(_ context.Context, limit, offset int) (models.AdminUserPage, error) {
	s.calls++
	s.limit = limit
	s.offset = offset
	return models.AdminUserPage{Users: []models.UserDTO{}, Limit: limit, Offset: offset}, s.listErr
}
func TestAdminListAuthorizationAndPagination(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tokens, err := auth.NewJWTService("admin-users-list-secret-at-least-32-bytes")
	if err != nil {
		t.Fatal(err)
	}
	token, _, _ := tokens.Issue("admin")
	for _, tc := range []struct {
		name, role, status, query, token string
		userErr, listErr                 error
		code, calls                      int
	}{
		{name: "admin", role: "admin", status: "active", token: token, code: 200, calls: 1},
		{name: "manager", role: "manager", status: "active", token: token, code: 403},
		{name: "employee", role: "employee", status: "active", token: token, code: 403},
		{name: "inactive", role: "admin", status: "inactive", token: token, code: 403},
		{name: "deleted", token: token, userErr: pgx.ErrNoRows, code: 401},
		{name: "no token", code: 401},
		{name: "bad limit", role: "admin", status: "active", token: token, query: "?limit=101", code: 400},
		{name: "negative offset", role: "admin", status: "active", token: token, query: "?offset=-1", code: 400},
		{name: "duplicate", role: "admin", status: "active", token: token, query: "?limit=1&limit=2", code: 400},
		{name: "page", role: "admin", status: "active", token: token, query: "?limit=10&offset=20", code: 200, calls: 1},
		{name: "database failure", role: "admin", status: "active", token: token, listErr: errors.New("private database secret"), code: 500, calls: 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			store := &listStore{role: models.UserRole(tc.role), status: models.UserStatus(tc.status), userErr: tc.userErr, listErr: tc.listErr}
			router := gin.New()
			router.GET("/admin/users", auth.RequireJWT(tokens), auth.RequireAdministrator(store), ListHandler(store))
			request := httptest.NewRequest("GET", "/admin/users"+tc.query, nil)
			if tc.token != "" {
				request.Header.Set("Authorization", "Bearer "+tc.token)
			}
			r := httptest.NewRecorder()
			router.ServeHTTP(r, request)
			if r.Code != tc.code || store.calls != tc.calls || strings.Contains(r.Body.String(), "private database") {
				t.Fatal(r.Code, r.Body.String(), store.calls)
			}
			if tc.name == "page" && (store.limit != 10 || store.offset != 20) {
				t.Fatal("pagination not passed")
			}
			if tc.code == 200 && (!strings.Contains(r.Body.String(), `"users":[]`) || r.Header().Get("Cache-Control") != "no-store") {
				t.Fatal(r.Body.String(), r.Header())
			}
		})
	}
}
