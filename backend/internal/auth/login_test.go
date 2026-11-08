package auth

import (
    "context"
    "encoding/json"
    "errors"
    "net/http"
    "net/http/httptest"
    "strings"
    "testing"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/jackc/pgx/v5"

    "github.com/nik003kk/DeskBooker/backend/internal/models"
)

type fakeUserFinder struct {
    gotEmail string
    calls int
    user models.User
    err error
}

func (f *fakeUserFinder) GetUserByEmail(_ context.Context, email string) (models.User, error) {
    f.calls++
    f.gotEmail = email
    if f.err != nil { return models.User{}, f.err }
    return f.user, nil
}

func loginTestRequest(t *testing.T, finder *fakeUserFinder, body, contentType string) *httptest.ResponseRecorder {
    t.Helper()
    gin.SetMode(gin.TestMode)
    router := gin.New()
    router.POST("/api/auth/login", LoginHandler(finder))
    request := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(body))
    if contentType != "" { request.Header.Set("Content-Type", contentType) }
    response := httptest.NewRecorder()
    router.ServeHTTP(response, request)
    return response
}

const goodLogin = `{"email":"Ana@Gmail.COM","password":"VeryStrongPass123!"}`

func validFinder(t *testing.T) *fakeUserFinder {
    t.Helper()
    hash, err := HashPassword("VeryStrongPass123!")
    if err != nil { t.Fatal(err) }
    return &fakeUserFinder{user: models.User{
        ID: "test-uuid", Email: "ana@gmail.com", PasswordHash: hash,
        FirstName: "Ana", LastName: "Novak", Role: models.RoleEmployee,
        CreatedAt: time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC),
    }}
}

func TestLoginSuccess(t *testing.T) {
    finder := validFinder(t)
    response := loginTestRequest(t, finder, goodLogin, "application/json")
    if response.Code != http.StatusOK { t.Fatalf("want 200 got %d: %s", response.Code, response.Body.String()) }
    if finder.calls != 1 || finder.gotEmail != "ana@gmail.com" {
        t.Fatalf("expected lower-case DB lookup, got %q calls=%d", finder.gotEmail, finder.calls)
    }
    if response.Header().Get("Cache-Control") != "no-store" { t.Fatal("missing no-store response header") }
    if response.Header().Get("Set-Cookie") != "" { t.Fatal("login must not silently create a session") }
    var data map[string]any
    if err := json.Unmarshal(response.Body.Bytes(), &data); err != nil { t.Fatal(err) }
    user := data["user"].(map[string]any)
    if user["email"] != "ana@gmail.com" || user["role"] != "employee" { t.Fatalf("wrong public user: %#v", user) }
    if strings.Contains(response.Body.String(), "password") ||
        strings.Contains(response.Body.String(), finder.user.PasswordHash) ||
        strings.Contains(response.Body.String(), "VeryStrongPass123!") {
        t.Fatal("login response leaked password material")
    }
    if _, ok := data["token"]; ok { t.Fatal("login must not issue an unimplemented token") }
}

func TestLoginWrongPasswordAndUnknownUserAreIndistinguishable(t *testing.T) {
    wrong := validFinder(t)
    wrongResponse := loginTestRequest(t, wrong,
        `{"email":"ana@gmail.com","password":"WrongPassword123!"}`, "application/json")
    missing := &fakeUserFinder{err: pgx.ErrNoRows}
    missingResponse := loginTestRequest(t, missing, goodLogin, "application/json")
    if wrongResponse.Code != http.StatusUnauthorized || missingResponse.Code != http.StatusUnauthorized {
        t.Fatalf("expected 401, got wrong=%d unknown=%d", wrongResponse.Code, missingResponse.Code)
    }
    if wrongResponse.Body.String() != missingResponse.Body.String() {
        t.Fatalf("unknown user and wrong password have different responses: %s vs %s",
            wrongResponse.Body.String(), missingResponse.Body.String())
    }
}

func TestLoginInputValidation(t *testing.T) {
    cases := []struct { name, body, contentType string; want int }{
        {"missing email", `{"password":"VeryStrongPass123!"}`, "application/json", 400},
        {"invalid email", `{"email":"not-email","password":"VeryStrongPass123!"}`, "application/json", 400},
        {"blank password", `{"email":"ana@gmail.com","password":""}`, "application/json", 400},
        {"oversized password", `{"email":"ana@gmail.com","password":"` + strings.Repeat("x", 73) + `"}`, "application/json", 400},
        {"unknown JSON field", `{"email":"ana@gmail.com","password":"VeryStrongPass123!","role":"admin"}`, "application/json", 400},
        {"malformed JSON", "{", "application/json", 400},
        {"multiple objects", goodLogin + goodLogin, "application/json", 400},
        {"JSON array", "[]", "application/json", 400},
        {"wrong content type", goodLogin, "text/plain", 415},
        {"fake JSON media type", goodLogin, "application/jsonp", 415},
    }
    for _, tc := range cases {
        t.Run(tc.name, func(t *testing.T) {
            finder := &fakeUserFinder{}
            response := loginTestRequest(t, finder, tc.body, tc.contentType)
            if response.Code != tc.want { t.Fatalf("want %d got %d: %s", tc.want, response.Code, response.Body.String()) }
            if finder.calls != 0 { t.Fatal("invalid input must not reach DB") }
        })
    }
}

func TestLoginDBErrorsAreNotLeaked(t *testing.T) {
    finder := &fakeUserFinder{err: errors.New("secret database address")}
    response := loginTestRequest(t, finder, goodLogin, "application/json")
    if response.Code != http.StatusInternalServerError || !strings.Contains(response.Body.String(), "login_failed") {
        t.Fatalf("expected generic 500, got %d: %s", response.Code, response.Body.String())
    }
    if strings.Contains(response.Body.String(), "secret database address") { t.Fatal("DB details leaked") }
}
