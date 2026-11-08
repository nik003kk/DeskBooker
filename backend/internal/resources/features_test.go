package resources

import (
    "context"
    "errors"
    "net/http"
    "net/http/httptest"
    "strings"
    "testing"

    "github.com/gin-gonic/gin"
    "github.com/jackc/pgx/v5"
    "github.com/nik003kk/DeskBooker/backend/internal/models"
)
const resourceUUID = "11111111-1111-4111-8111-111111111111"
type featureStore struct { calls int; input []string; err error }
func (s *featureStore) AddFeatures(_ context.Context, _ string, items []string)([]models.ResourceFeature,error){
    s.calls++;s.input=items
    if s.err!=nil{return nil,s.err}
    return []models.ResourceFeature{{ResourceID:resourceUUID,Feature:items[0]}},nil
}
func testFeatureRequest(t *testing.T, store *featureStore,body string) *httptest.ResponseRecorder {
    t.Helper()
    gin.SetMode(gin.TestMode)
    r:=gin.New()
    r.POST("/api/resources/:id/features",AddFeaturesHandler(store))
    request:=httptest.NewRequest(http.MethodPost,"/api/resources/"+resourceUUID+"/features",strings.NewReader(body))
    request.Header.Set("Content-Type","application/json")
    response:=httptest.NewRecorder()
    r.ServeHTTP(response,request)
    return response
}
func TestAddResourceFeatures(t *testing.T){
    store:=&featureStore{}
    response:=testFeatureRequest(t,store,`{"features":["Monitor","USB-C"]}`)
    if response.Code!=201 || store.calls!=1 || len(store.input)!=2 {
        t.Fatalf("status %d, calls %d",response.Code,store.calls)
    }
}
func TestAddFeaturesValidation(t *testing.T){
    for _,body:=range []string{
        `{}`,
        `{"features":[]}`,
        `{"features":[""]}`,
        `{"features":["  Monitor"]}`,
        `{"features":["Monitor","monitor"]}`,
        `{"features":["Mouse"],"role":"admin"}`,
    } {
        store:=&featureStore{}
        resp:=testFeatureRequest(t,store,body)
        if resp.Code!=400 || store.calls!=0 {t.Fatalf("%s: status %d",body,resp.Code)}
    }
}
func TestFeatureResourceNotFoundAndDBError(t *testing.T){
    for _,tc:=range []struct{err error; want int}{
        {pgx.ErrNoRows,404},{errors.New("private DB secret"),500},
    }{
        resp:=testFeatureRequest(t,&featureStore{err:tc.err},`{"features":["Monitor"]}`)
        if resp.Code!=tc.want || strings.Contains(resp.Body.String(),"private DB secret") {
            t.Fatalf("%d %s",resp.Code,resp.Body.String())
        }
    }
}
