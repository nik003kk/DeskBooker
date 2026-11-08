package auth

import (
    "context"
    "errors"
    "net/http"
    "net/http/httptest"
    "strings"
    "testing"

    "github.com/gin-gonic/gin"
    "github.com/nik003kk/DeskBooker/backend/internal/models"
)
type testUpdater struct { calls int; input models.UpdateUserProfileInput; err error }
func (s *testUpdater) UpdateUserProfile(_ context.Context, _ string, in models.UpdateUserProfileInput) (models.User,error) {
    s.calls++;s.input=in
    if s.err!=nil{return models.User{},s.err}
    return models.User{ID:"test",Email:"ana@gmail.com",FirstName:"Ana",
        LastName:"Novak",Role:models.RoleEmployee,Status:models.UserActive},nil
}
func TestProfilePatchValidatesAndPreservesEmail(t *testing.T) {
    gin.SetMode(gin.TestMode)
    for _, tc:=range []struct { body string; code int; calls int }{
        {`{"first_name":"Ana","department":"Engineering"}`, 200,1},
        {`{"department":null}`,200,1},
        {`{"email":"attacker@example.com"}`,400,0},
        {`{"role":"admin"}`,400,0},
        {`{"status":"active"}`,400,0},
        {`{"first_name":""}`,400,0},
        {`{"last_name":" "}`,400,0},
        {`{"department":" "}`,400,0},
        {`{}`,400,0},
        {`{"first_name":"Ana"}{"last_name":"Novak"}`,400,0},
    } {
        store:=&testUpdater{}
        router:=gin.New()
        router.PATCH("/profile",func(c *gin.Context){c.Set("auth_user_id","uuid")},UpdateProfileHandler(store))
        request:=httptest.NewRequest(http.MethodPatch,"/profile",strings.NewReader(tc.body))
        request.Header.Set("Content-Type","application/json")
        response:=httptest.NewRecorder()
        router.ServeHTTP(response,request)
        if response.Code!=tc.code || store.calls!=tc.calls {
            t.Fatalf("body %s returned %d/%d: %s",tc.body,response.Code,store.calls,response.Body.String())
        }
        if tc.code==200 && strings.Contains(response.Body.String(),"password_hash"){t.Fatal("hash leaked")}
    }
}
func TestProfilePatchDatabaseErrorsAreSanitized(t *testing.T) {
    store:=&testUpdater{err:errors.New("private db info")}
    gin.SetMode(gin.TestMode)
    router:=gin.New()
    router.PATCH("/profile",func(c *gin.Context){c.Set("auth_user_id","uuid")},UpdateProfileHandler(store))
    request:=httptest.NewRequest("PATCH","/profile",strings.NewReader(`{"last_name":"Other"}`))
    request.Header.Set("Content-Type","application/json")
    response:=httptest.NewRecorder()
    router.ServeHTTP(response,request)
    if response.Code!=500 || strings.Contains(response.Body.String(),"private db info"){t.Fatal(response.Body.String())}
}
