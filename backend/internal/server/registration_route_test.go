package server

import (
    "context"
    "net/http"
    "net/http/httptest"
    "strings"
    "testing"
    "time"

    "github.com/nik003kk/DeskBooker/backend/internal/auth"
    "github.com/nik003kk/DeskBooker/backend/internal/models"
)

type registrationStore struct { calls int }

func (s *registrationStore) CreateUser(_ context.Context, in models.CreateUserInput) (models.User, error) {
    s.calls++
    return models.User{
        ID: "user-uuid", Email: in.Email, PasswordHash: in.PasswordHash,
        FirstName: in.FirstName, LastName: in.LastName, Role: models.RoleEmployee,
        CreatedAt: time.Now().UTC(),
    }, nil
}

func TestRegistrationRouteIsWired(t *testing.T) {
    store := &registrationStore{}
    router := NewRouterWithRegistration(
        fakeDatabase{}, mustOrigins(t, ""),
        auth.RegisterHandler(store, "example.com"),
    )
    request := httptest.NewRequest(http.MethodPost, "/api/auth/register",
        strings.NewReader(`{"first_name":"Ana","last_name":"Novak","email":"ana@example.com","password":"VeryStrongPass123!"}`))
    request.Header.Set("Content-Type", "application/json")
    response := httptest.NewRecorder()
    router.ServeHTTP(response, request)
    if response.Code != http.StatusCreated || store.calls != 1 {
        t.Fatalf("expected HTTP 201, 1 store call; got %d, %d: %s", response.Code, store.calls, response.Body.String())
    }
    if strings.Contains(response.Body.String(), "password") {
        t.Fatalf("password leaked in HTTP response: %s", response.Body.String())
    }
}
