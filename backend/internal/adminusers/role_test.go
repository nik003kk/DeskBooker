package adminusers

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/nik003kk/DeskBooker/backend/internal/models"
	"github.com/nik003kk/DeskBooker/backend/internal/repository"
	"net/http/httptest"
	"strings"
	"testing"
)

const targetUser = "11111111-1111-4111-8111-111111111111"

type fakeChanger struct {
	calls         int
	actor, target string
	role          models.UserRole
	err           error
}

func (s *fakeChanger) ChangeUserRole(_ context.Context, actor, target string, role models.UserRole) (models.UserDTO, error) {
	s.calls++
	s.actor = actor
	s.target = target
	s.role = role
	return models.UserDTO{ID: target, Role: role}, s.err
}
func roleRequest(store RoleChanger, body string) *httptest.ResponseRecorder {
	router := gin.New()
	router.PATCH("/admin/users/:id/role", func(c *gin.Context) { c.Set("auth_user_id", "trusted-admin") }, ChangeRoleHandler(store))
	request := httptest.NewRequest("PATCH", "/admin/users/"+targetUser+"/role", strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	r := httptest.NewRecorder()
	router.ServeHTTP(r, request)
	return r
}
func TestChangeRoleValidatesPayloadAndUsesJWTActor(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tc := range []struct {
		input string
		role  models.UserRole
	}{{"EMPLOYEE", models.RoleEmployee}, {"OFFICE_MANAGER", models.RoleManager}, {"ADMIN", models.RoleAdmin}, {"manager", models.RoleManager}} {
		store := &fakeChanger{}
		r := roleRequest(store, `{"role":"`+tc.input+`"}`)
		if r.Code != 200 || store.actor != "trusted-admin" || store.target != targetUser || store.role != tc.role {
			t.Fatal(r.Code, r.Body.String())
		}
	}
	for _, body := range []string{`{}`, `{"role":"root"}`, `{"role":null}`, `{"role":42}`, `{"role":"ADMIN","actor_id":"attacker"}`, `{"role":"ADMIN","status":"active"}`, `{"role":"ADMIN"} {}`} {
		store := &fakeChanger{}
		r := roleRequest(store, body)
		if r.Code != 400 || store.calls != 0 {
			t.Fatal(body, r.Code)
		}
	}
}
func TestChangeRoleErrorMapping(t *testing.T) {
	for _, tc := range []struct {
		err  error
		code int
	}{{repository.ErrSelfRoleChange, 403}, {repository.ErrAdminRequired, 403}, {pgx.ErrNoRows, 404}, {errors.New("private DB password"), 500}} {
		r := roleRequest(&fakeChanger{err: tc.err}, `{"role":"ADMIN"}`)
		if r.Code != tc.code || strings.Contains(r.Body.String(), "private DB") {
			t.Fatal(r.Code, r.Body.String())
		}
	}
}
