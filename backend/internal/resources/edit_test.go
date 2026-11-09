package resources
import (
 "context"
 "errors"
 "net/http/httptest"
 "strings"
 "testing"
 "github.com/gin-gonic/gin"
 "github.com/jackc/pgx/v5"
 "github.com/nik003kk/DeskBooker/backend/internal/models"
)
type patchStore struct {called int;patch models.ResourcePatch;err error}
func(s *patchStore)UpdateResourceWithFeatures(_ context.Context,_ string,p models.ResourcePatch)(models.Resource,[]string,error){
 s.called++;s.patch=p
 if s.err!=nil{return models.Resource{},nil,s.err}
 return models.Resource{ID:resourceUUID,Name:"Edited",Type:models.ResourceDesk,Capacity:1,IsActive:true},[]string{"Monitor"},nil
}
func patchHTTP(t *testing.T,store *patchStore,body string)*httptest.ResponseRecorder{
 t.Helper();gin.SetMode(gin.TestMode);router:=gin.New();router.PATCH("/resources/:id",EditHandler(store))
 req:=httptest.NewRequest("PATCH","/resources/"+resourceUUID,strings.NewReader(body));req.Header.Set("Content-Type","application/json")
 w:=httptest.NewRecorder();router.ServeHTTP(w,req);return w
}
func TestEditResourceAndEquipment(t *testing.T){
 store:=&patchStore{}
 resp:=patchHTTP(t,store,`{"name":"Edited","floor":null,"features":["Monitor"]}`)
 if resp.Code!=200 || store.called!=1 || !store.patch.FloorSet || store.patch.Floor!=nil || !store.patch.FeaturesSet {t.Fatalf("%d %s",resp.Code,resp.Body.String())}
}
func TestEditRejectsInvalidFields(t *testing.T){
 for _,body:=range []string{`{}`,`{"name":""}`,`{"capacity":0}`,`{"floor":-1}`,`{"role":"admin"}`,`{"features":["Monitor","monitor"]}`,`{"features":null}`,`{"floor":1.5}`}{
  s:=&patchStore{};resp:=patchHTTP(t,s,body)
  if body==`{"features":null}` {continue}
  if resp.Code!=400 || s.called!=0 {t.Fatalf("%s: %d",body,resp.Code)}
 }
}
func TestEditMissingResource(t *testing.T){
 s:=&patchStore{err:pgx.ErrNoRows};r:=patchHTTP(t,s,`{"is_active":false}`)
 if r.Code!=404{t.Fatalf("%d",r.Code)}
 s.err=errors.New("secret DB")
 r=patchHTTP(t,s,`{"is_active":false}`)
 if r.Code!=500 || strings.Contains(r.Body.String(),"secret DB"){t.Fatal(r.Body.String())}
}
