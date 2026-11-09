package auth

import (
    "net/http"
    "net/http/httptest"
    "strings"
    "testing"

    "github.com/gin-gonic/gin"
    "github.com/nik003kk/DeskBooker/backend/internal/models"
)
func TestProfileRequiresTokenAndHidesPasswordHash(t *testing.T) {
    gin.SetMode(gin.TestMode)
    tokens, _ := NewJWTService(testJWTSecret)
    token, _, err := tokens.Issue("user-123")
    if err != nil {t.Fatal(err)}
    db := &jwtUserDB{user: models.User{ID:"user-123", Email:"ana@gmail.com",
        FirstName:"Ana", LastName:"Novak", Role:models.RoleEmployee,
        Status:models.UserActive, PasswordHash:"HIDDEN-HASH"}}
    router := gin.New()
    router.GET("/api/users/me", RequireJWT(tokens), ProfileHandler(db))
    noToken := httptest.NewRecorder()
    router.ServeHTTP(noToken, httptest.NewRequest(http.MethodGet,"/api/users/me",nil))
    if noToken.Code!=401 || db.lookups!=0 {t.Fatal("unauthorized profile request reached DB")}
    request:=httptest.NewRequest(http.MethodGet,"/api/users/me",nil)
    request.Header.Set("Authorization","Bearer "+token)
    response:=httptest.NewRecorder()
    router.ServeHTTP(response,request)
    if response.Code!=200 || !strings.Contains(response.Body.String(), `"status":"active"`) ||
        strings.Contains(response.Body.String(),"HIDDEN-HASH") {
        t.Fatalf("wrong profile: %d %s",response.Code,response.Body.String())
    }
}
