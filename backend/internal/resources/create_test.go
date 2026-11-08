package resources

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
type createStore struct {
    called int
    input models.CreateResourceInput
    err error
}
func (s *createStore) CreateResource(_ context.Context,in models.CreateResourceInput)(models.Resource,error){
    s.called++;s.input=in
    if s.err!=nil{return models.Resource{},s.err}
    return models.Resource{ID:"uuid",Name:in.Name,Type:in.Type,
        Capacity:in.Capacity,IsActive:in.IsActive},nil
}
func createRequestTest(t *testing.T,store *createStore,payload string) *httptest.ResponseRecorder {
    t.Helper()
    gin.SetMode(gin.TestMode)
    router:=gin.New()
    router.POST("/api/resources",CreateHandler(store))
    request:=httptest.NewRequest(http.MethodPost,"/api/resources",strings.NewReader(payload))
    request.Header.Set("Content-Type","application/json")
    response:=httptest.NewRecorder()
    router.ServeHTTP(response,request)
    return response
}
func TestCreateResourceVariants(t *testing.T){
    for _,kind:=range []string{"DESK","MEETING_ROOM","PARKING_SPOT"} {
        store:=&createStore{}
        response:=createRequestTest(t,store,`{"name":"Resource","type":"`+kind+`","capacity":2}`)
        if response.Code!=201 || store.called!=1 || !store.input.IsActive ||
            string(store.input.Type)!=kind {t.Fatalf("%s: %d %s",kind,response.Code,response.Body.String())}
    }
}
func TestCreateResourceInvalid(t *testing.T){
    for _,body:=range []string{
        `{"name":"","type":"DESK","capacity":1}`,
        `{"name":"Room","type":"INVALID","capacity":1}`,
        `{"name":"Room","type":"MEETING_ROOM","capacity":0}`,
        `{"name":"Room","type":"DESK","capacity":1,"role":"admin"}`,
        `{}`,
    } {
        store:=&createStore{}
        response:=createRequestTest(t,store,body)
        if response.Code!=400 || store.called!=0 {t.Fatalf("%s: %d",body,response.Code)}
    }
}
func TestCreateResourceSanitizesDatabaseErrors(t *testing.T){
    store:=&createStore{err:errors.New("secret db error")}
    response:=createRequestTest(t,store,`{"name":"Room","type":"DESK","capacity":1}`)
    if response.Code!=500 || strings.Contains(response.Body.String(),"secret db error"){t.Fatalf("%d %s",response.Code,response.Body.String())}
}
