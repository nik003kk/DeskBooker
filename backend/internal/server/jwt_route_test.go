package server

import (
    "context"
    "net/http"
    "net/http/httptest"
    "strings"
    "testing"

    "github.com/nik003kk/DeskBooker/backend/internal/auth"
    "github.com/nik003kk/DeskBooker/backend/internal/models"
)

type profileStore struct {
    called int
}

func (s *profileStore) GetUser(_ context.Context, id string) (models.User, error) {
    s.called++
    return models.User{ID:id, Email:"ana@gmail.com", Role:models.RoleEmployee}, nil
}

func TestJWTProtectedRouteInMainRouter(t *testing.T) {
    tokens, err := auth.NewJWTService("local-test-key-must-be-at-least-32-bytes-long")
    if err != nil { t.Fatal(err) }
    store := &profileStore{}
    router := NewRouterWithJWT(fakeDatabase{}, mustOrigins(t, ""),
        nil, nil, auth.RequireJWT(tokens), auth.MeHandler(store))
    noToken := httptest.NewRecorder()
    router.ServeHTTP(noToken, httptest.NewRequest(http.MethodGet, "/api/auth/me", nil))
    if noToken.Code != http.StatusUnauthorized || store.called != 0 {
        t.Fatalf("unauthenticated request accepted: %d, DB calls: %d", noToken.Code, store.called)
    }
    token, _, err := tokens.Issue("test-user-id")
    if err != nil {t.Fatal(err)}
    request := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
    request.Header.Set("Authorization", "Bearer "+token)
    response := httptest.NewRecorder()
    router.ServeHTTP(response,request)
    if response.Code!=http.StatusOK || store.called!=1 || !strings.Contains(response.Body.String(),"ana@gmail.com") {
        t.Fatalf("authenticated request failed: status=%d body=%s calls=%d",response.Code,response.Body.String(),store.called)
    }
}
