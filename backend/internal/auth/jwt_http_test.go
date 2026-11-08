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

type jwtUserDB struct {
    user models.User
    err error
    lookups int
}
func (s *jwtUserDB) GetUser(_ context.Context, id string) (models.User,error) {
    s.lookups++
    if s.err != nil { return models.User{},s.err }
    if id != s.user.ID { return models.User{},pgx.ErrNoRows }
    return s.user,nil
}

func TestJWTLoginAndProtectedProfile(t *testing.T) {
    service, err := NewJWTService(testJWTSecret)
    if err != nil { t.Fatal(err) }
    hash, err := HashPassword("VeryStrongPass123!")
    if err != nil { t.Fatal(err) }
    user := models.User{ID: "user-123", Email: "ana@gmail.com", PasswordHash: hash,
        FirstName: "Ana", LastName: "Novak", Role: models.RoleEmployee,
        CreatedAt: time.Now().UTC()}
    finder := &fakeUserFinder{user:user}
    db := &jwtUserDB{user:user}
    gin.SetMode(gin.TestMode)
    router := gin.New()
    router.POST("/api/auth/login", LoginHandlerWithJWT(finder, service))
    router.GET("/api/auth/me", RequireJWT(service), MeHandler(db))

    loginReq := httptest.NewRequest(http.MethodPost, "/api/auth/login",
        strings.NewReader(`{"email":"ana@gmail.com","password":"VeryStrongPass123!"}`))
    loginReq.Header.Set("Content-Type","application/json")
    loginResp := httptest.NewRecorder()
    router.ServeHTTP(loginResp,loginReq)
    if loginResp.Code!=http.StatusOK { t.Fatalf("login: %d %s",loginResp.Code,loginResp.Body.String()) }
    var data struct {
        AccessToken string `json:"access_token"`
        TokenType string `json:"token_type"`
        ExpiresAt time.Time `json:"expires_at"`
    }
    if err:=json.Unmarshal(loginResp.Body.Bytes(),&data);err!=nil { t.Fatal(err) }
    if data.AccessToken=="" || data.TokenType!="Bearer" || data.ExpiresAt.IsZero() { t.Fatalf("missing token: %+v",data) }
    if strings.Contains(loginResp.Body.String(),hash) || strings.Contains(loginResp.Body.String(),"VeryStrongPass123!") {
        t.Fatal("sensitive data leaked")
    }
    meReq:=httptest.NewRequest(http.MethodGet,"/api/auth/me",nil)
    meReq.Header.Set("Authorization","Bearer "+data.AccessToken)
    meResp:=httptest.NewRecorder()
    router.ServeHTTP(meResp,meReq)
    if meResp.Code!=http.StatusOK || db.lookups!=1 || !strings.Contains(meResp.Body.String(),"ana@gmail.com") {
        t.Fatalf("me: %d %s lookups=%d",meResp.Code,meResp.Body.String(),db.lookups)
    }

    for _, authHeader := range []string{"", "Basic abc", "Bearer garbage", "Bearer "+data.AccessToken+"extra"} {
        req:=httptest.NewRequest(http.MethodGet,"/api/auth/me",nil)
        if authHeader!="" {req.Header.Set("Authorization",authHeader)}
        resp:=httptest.NewRecorder()
        router.ServeHTTP(resp,req)
        if resp.Code!=http.StatusUnauthorized { t.Fatalf("authorization %q accepted: %d", authHeader, resp.Code) }
    }
    if db.lookups!=1 { t.Fatal("rejected JWT must not query database") }

    db.err=pgx.ErrNoRows
    deletedReq:=httptest.NewRequest(http.MethodGet,"/api/auth/me",nil)
    deletedReq.Header.Set("Authorization","Bearer "+data.AccessToken)
    deletedResp:=httptest.NewRecorder()
    router.ServeHTTP(deletedResp,deletedReq)
    if deletedResp.Code!=http.StatusUnauthorized { t.Fatalf("deleted user returned %d",deletedResp.Code) }

    db.err=errors.New("internal db secret")
    failedReq:=httptest.NewRequest(http.MethodGet,"/api/auth/me",nil)
    failedReq.Header.Set("Authorization","Bearer "+data.AccessToken)
    failedResp:=httptest.NewRecorder()
    router.ServeHTTP(failedResp,failedReq)
    if failedResp.Code!=http.StatusInternalServerError ||
        strings.Contains(failedResp.Body.String(),"internal db secret") {
        t.Fatalf("DB error not sanitized: %d %s",failedResp.Code,failedResp.Body.String())
    }
}

func TestJWTExpiredTokenRejectedByMiddleware(t *testing.T) {
    service,_:=NewJWTService(testJWTSecret)
    fixed:=time.Date(2030,1,1,0,0,0,0,time.UTC)
    service.now=func()time.Time{return fixed}
    token,_,err:=service.Issue("user-123")
    if err!=nil {t.Fatal(err)}
    service.now=func()time.Time{return fixed.Add(tokenTTL)}
    gin.SetMode(gin.TestMode)
    router:=gin.New()
    router.GET("/me",RequireJWT(service),func(c *gin.Context){c.Status(http.StatusOK)})
    req:=httptest.NewRequest(http.MethodGet,"/me",nil)
    req.Header.Set("Authorization","Bearer "+token)
    resp:=httptest.NewRecorder()
    router.ServeHTTP(resp,req)
    if resp.Code!=http.StatusUnauthorized {t.Fatalf("expired token accepted: %d",resp.Code)}
}
