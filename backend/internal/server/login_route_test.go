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

type loginStore struct {
    user models.User
    calls int
}

func (s *loginStore) GetUserByEmail(_ context.Context, _ string) (models.User, error) {
    s.calls++
    return s.user, nil
}

func TestLoginRouteIsWired(t *testing.T) {
    hash, err := auth.HashPassword("VeryStrongPass123!")
    if err != nil { t.Fatal(err) }
    store := &loginStore{user: models.User{ID: "test-id", Email: "ana@gmail.com",
        PasswordHash: hash, Role: models.RoleEmployee}}
    router := NewRouterWithAuth(fakeDatabase{}, mustOrigins(t, ""),
        nil, auth.LoginHandler(store))
    request := httptest.NewRequest(http.MethodPost, "/api/auth/login",
        strings.NewReader(`{"email":"ana@gmail.com","password":"VeryStrongPass123!"}`))
    request.Header.Set("Content-Type", "application/json")
    response := httptest.NewRecorder()
    router.ServeHTTP(response, request)
    if response.Code != http.StatusOK || store.calls != 1 {
        t.Fatalf("want HTTP 200, 1 lookup; got %d, %d: %s", response.Code, store.calls, response.Body.String())
    }
}
