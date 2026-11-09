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

type detailStore struct{ err error }
func(s detailStore)GetResource(context.Context,string)(models.Resource,error){
 if s.err!=nil{return models.Resource{},s.err}
 return models.Resource{ID:resourceUUID,Name:"Desk",Type:models.ResourceDesk,Capacity:1},nil
}
func(s detailStore)ListFeatures(context.Context,string)([]models.ResourceFeature,error){
 return []models.ResourceFeature{{ResourceID:resourceUUID,Feature:"Monitor"}},nil
}
func TestAdminDetailReturnsResourceAndFeatures(t *testing.T) {
 gin.SetMode(gin.TestMode)
 router:=gin.New()
 router.GET("/admin/:id",AdminDetailHandler(detailStore{}))
 response:=httptest.NewRecorder()
 router.ServeHTTP(response,httptest.NewRequest(http.MethodGet,"/admin/"+resourceUUID,nil))
 if response.Code!=200 || !strings.Contains(response.Body.String(),"Monitor") {
  t.Fatalf("status=%d body=%s",response.Code,response.Body.String())
 }
}
func TestAdminDetailHandlesUnknownResource(t *testing.T){
 gin.SetMode(gin.TestMode);router:=gin.New();router.GET("/admin/:id",AdminDetailHandler(detailStore{err:pgx.ErrNoRows}))
 response:=httptest.NewRecorder();router.ServeHTTP(response,httptest.NewRequest("GET","/admin/"+resourceUUID,nil))
 if response.Code!=404{t.Fatal(response.Code)}
}
func TestAdminDetailSanitizesErrors(t *testing.T){
 gin.SetMode(gin.TestMode);router:=gin.New();router.GET("/admin/:id",AdminDetailHandler(detailStore{err:errors.New("private DB")}))
 response:=httptest.NewRecorder();router.ServeHTTP(response,httptest.NewRequest("GET","/admin/"+resourceUUID,nil))
 if response.Code!=500 || strings.Contains(response.Body.String(),"private DB"){t.Fatal(response.Body.String())}
}
