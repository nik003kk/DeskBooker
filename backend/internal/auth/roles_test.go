package auth

import (
    "errors"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/gin-gonic/gin"
    "github.com/jackc/pgx/v5"
    "github.com/nik003kk/DeskBooker/backend/internal/models"
)

func TestRequireRolesEnforces401And403(t *testing.T) {
    gin.SetMode(gin.TestMode)
    tokens, err := NewJWTService(testJWTSecret)
    if err != nil { t.Fatal(err) }
    token, _, err := tokens.Issue("user-123")
    if err != nil { t.Fatal(err) }
    db := &jwtUserDB{user:models.User{ID:"user-123", Role:models.RoleEmployee,Status:models.UserActive}}
    router := gin.New()
    router.GET("/private", RequireJWT(tokens), RequireRoles(db, models.RoleManager, models.RoleAdmin),
        func(c *gin.Context) { c.Status(http.StatusOK) })
    check := func(bearer string, want int) {
        t.Helper()
        req := httptest.NewRequest(http.MethodGet, "/private", nil)
        if bearer != "" { req.Header.Set("Authorization", bearer) }
        response := httptest.NewRecorder()
        router.ServeHTTP(response, req)
        if response.Code != want { t.Fatalf("want %d got %d: %s", want, response.Code, response.Body.String()) }
    }
    check("", http.StatusUnauthorized)
    check("Bearer "+token, http.StatusForbidden)
    db.user.Role = models.RoleManager
    check("Bearer "+token, http.StatusOK)
    db.user.Role = models.RoleAdmin
    check("Bearer "+token, http.StatusOK)
    db.err = pgx.ErrNoRows
    check("Bearer "+token, http.StatusUnauthorized)
    db.err = errors.New("internal-secret")
    check("Bearer "+token, http.StatusInternalServerError)
}
