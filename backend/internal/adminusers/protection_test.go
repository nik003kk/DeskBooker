package adminusers

import (
	"github.com/gin-gonic/gin"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAccountDeletionNeverDeletesUsers(t *testing.T) {
	const actor = "aabbccdd-1111-4111-8111-111111111111"
	for _, tc := range []struct {
		id   string
		code int
	}{{actor, 403}, {strings.ToUpper(actor), 403}, {targetUser, 403}, {"bad", 400}} {
		router := gin.New()
		router.DELETE("/users/:id", func(c *gin.Context) { c.Set("auth_user_id", actor) }, ProtectDeletionHandler())
		r := httptest.NewRecorder()
		router.ServeHTTP(r, httptest.NewRequest("DELETE", "/users/"+tc.id, nil))
		if r.Code != tc.code {
			t.Fatal(r.Code, r.Body.String())
		}
	}
}
func TestOversizedRoleRequestRejectedBeforeMutation(t *testing.T) {
	store := &fakeChanger{}
	r := roleRequest(store, `{"role":"ADMIN"}`+strings.Repeat(" ", 1100)+`{}`)
	if r.Code != 400 || store.calls != 0 {
		t.Fatal("oversized JSON reached role mutation", r.Code)
	}
}
