package auth

import (
    "bytes"
    "context"
    "encoding/json"
    "errors"
    "net/http"
    "net/http/httptest"
    "strings"
    "testing"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/jackc/pgx/v5/pgconn"
    "golang.org/x/crypto/bcrypt"

    "github.com/nik003kk/DeskBooker/backend/internal/models"
)

type mockCreator struct {
    calls int
    input models.CreateUserInput
    err error
}
func (m *mockCreator) CreateUser(_ context.Context, in models.CreateUserInput) (models.User, error) {
    m.calls++
    m.input = in
    if m.err != nil { return models.User{}, m.err }
    return models.User{ID: "test-id", Email: in.Email, PasswordHash: in.PasswordHash,
        FirstName: in.FirstName, LastName: in.LastName, Role: models.RoleEmployee,
        CreatedAt: time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC)}, nil
}

func send(t *testing.T, creator *mockCreator, domain, body string, contentType string) *httptest.ResponseRecorder {
    t.Helper()
    gin.SetMode(gin.TestMode)
    router := gin.New()
    router.POST("/api/auth/register", RegisterHandler(creator, domain))
    response := httptest.NewRecorder()
    request := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBufferString(body))
    if contentType != "" { request.Header.Set("Content-Type", contentType) }
    router.ServeHTTP(response, request)
    return response
}

const goodRequest = `{"first_name":"Ana","last_name":"Novak","email":"Ana@Example.COM","password":"VeryStrongPass123!"}`

func TestRegistrationSuccessHashesPasswordAndHidesHash(t *testing.T) {
    creator := &mockCreator{}
    response := send(t, creator, "example.com", goodRequest, "application/json")
    if response.Code != http.StatusCreated { t.Fatalf("expected 201, got %d: %s", response.Code, response.Body.String()) }
    if creator.calls != 1 || creator.input.Email != "ana@example.com" {
        t.Fatalf("expected one create with normalized company email; got %+v", creator.input)
    }
    if creator.input.PasswordHash == "VeryStrongPass123!" ||
        bcrypt.CompareHashAndPassword([]byte(creator.input.PasswordHash), []byte("VeryStrongPass123!")) != nil {
        t.Fatal("password was not stored as a valid bcrypt hash")
    }
    body := response.Body.String()
    if strings.Contains(body, "password") || strings.Contains(body, "VeryStrongPass123!") ||
        strings.Contains(body, creator.input.PasswordHash) {
        t.Fatalf("response leaked password material: %s", body)
    }
    var result map[string]any
    if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil { t.Fatal(err) }
    user := result["user"].(map[string]any)
    if user["email"] != "ana@example.com" || user["role"] != "employee" { t.Fatalf("unexpected user response: %v", user) }
}

func TestRegistrationValidation(t *testing.T) {
    cases := []struct{
        name, domain, body, contentType string
        code int
    }{
        {"missing domain", "", goodRequest, "application/json", 503},
        {"public email", "example.com", `{"first_name":"Ana","last_name":"Novak","email":"ana@gmail.com","password":"VeryStrongPass123!"}`, "application/json", 400},
        {"invalid email", "example.com", `{"first_name":"Ana","last_name":"Novak","email":"not-email","password":"VeryStrongPass123!"}`, "application/json", 400},
        {"missing name", "example.com", `{"first_name":"","last_name":"Novak","email":"ana@example.com","password":"VeryStrongPass123!"}`, "application/json", 400},
        {"blank last name", "example.com", `{"first_name":"Ana","last_name":"  ","email":"ana@example.com","password":"VeryStrongPass123!"}`, "application/json", 400},
        {"short password", "example.com", `{"first_name":"Ana","last_name":"Novak","email":"ana@example.com","password":"123"}`, "application/json", 400},
        {"unknown field", "example.com", `{"first_name":"Ana","last_name":"Novak","email":"ana@example.com","password":"VeryStrongPass123!","role":"admin"}`, "application/json", 400},
        {"two objects", "example.com", goodRequest + goodRequest, "application/json", 400},
        {"invalid json", "example.com", "{", "application/json", 400},
        {"wrong content type", "example.com", goodRequest, "text/plain", 415},
    }
    for _, tc := range cases {
        t.Run(tc.name, func(t *testing.T) {
            store := &mockCreator{}
            response := send(t, store, tc.domain, tc.body, tc.contentType)
            if response.Code != tc.code { t.Fatalf("expected %d got %d: %s", tc.code, response.Code, response.Body.String()) }
            if store.calls != 0 { t.Fatalf("invalid request reached DB: %d calls", store.calls) }
        })
    }
}

func TestRegistrationDuplicateEmailAndDatabaseError(t *testing.T) {
    for _, tc := range []struct{ name string; err error; code int; token string }{
        {"duplicate", &pgconn.PgError{Code: "23505", ConstraintName: "users_email_lower_unique_idx"}, 409, "email_exists"},
        {"db offline", errors.New("connection leaked-secret"), 500, "registration_failed"},
    } {
        t.Run(tc.name, func(t *testing.T) {
            creator := &mockCreator{err: tc.err}
            response := send(t, creator, "example.com", goodRequest, "application/json")
            if response.Code != tc.code || !strings.Contains(response.Body.String(), tc.token) {
                t.Fatalf("unexpected response: %d %s", response.Code, response.Body.String())
            }
            if strings.Contains(response.Body.String(), "leaked-secret") { t.Fatal("DB details leaked to client") }
        })
    }
}
