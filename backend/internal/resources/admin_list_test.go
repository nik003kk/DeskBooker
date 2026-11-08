package resources
import (
 "context"
 "errors"
 "net/http"
 "net/http/httptest"
 "testing"
 "strings"
 "github.com/gin-gonic/gin"
 "github.com/nik003kk/DeskBooker/backend/internal/models"
)
type testLister struct{err error}
func(s testLister)ListAllResources(context.Context)([]models.Resource,error){
 if s.err!=nil{return nil,s.err}
 return []models.Resource{{ID:"a",Name:"Active",IsActive:true},{ID:"b",Name:"Inactive",IsActive:false}},nil
}
func TestAdminListIncludesInactive(t *testing.T){
 gin.SetMode(gin.TestMode);r:=gin.New();r.GET("/admin",AdminListHandler(testLister{}))
 w:=httptest.NewRecorder();r.ServeHTTP(w,httptest.NewRequest("GET","/admin",nil))
 if w.Code!=200{t.Fatal(w.Code)}
 if !strings.Contains(w.Body.String(),`"is_active":true` ) || !strings.Contains(w.Body.String(),`"is_active":false`){t.Fatal(w.Body.String())}
}
func TestAdminListDBError(t *testing.T){
 gin.SetMode(gin.TestMode);r:=gin.New();r.GET("/admin",AdminListHandler(testLister{err:errors.New("private error")}))
 w:=httptest.NewRecorder();r.ServeHTTP(w,httptest.NewRequest("GET","/admin",nil))
 if w.Code!=500{t.Fatal(w.Code)}
}
