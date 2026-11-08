package resources
import(
 "net/http/httptest"
 "strings"
 "testing"
 "github.com/gin-gonic/gin"
)
func activationHTTP(t *testing.T,store *patchStore,body string)*httptest.ResponseRecorder{
 t.Helper();gin.SetMode(gin.TestMode);router:=gin.New();router.PATCH("/resources/:id/activation",ActivationHandler(store))
 req:=httptest.NewRequest("PATCH","/resources/"+resourceUUID+"/activation",strings.NewReader(body));req.Header.Set("Content-Type","application/json")
 w:=httptest.NewRecorder();router.ServeHTTP(w,req);return w
}
func TestDeactivateRetainsResourceAndOnlyChangesActiveFlag(t *testing.T){
 store:=&patchStore{}
 response:=activationHTTP(t,store,`{"is_active":false}`)
 if response.Code!=200 || store.called!=1 || store.patch.IsActive==nil || *store.patch.IsActive ||
  store.patch.Name!=nil || store.patch.FeaturesSet {
  t.Fatalf("invalid deactivation: %d %+v",response.Code,store.patch)
 }
}
func TestActivationRejectsInvalidPayload(t *testing.T){
 for _,body:=range []string{`{}`,`{"is_active":null}`,`{"is_active":"false"}`,`{"is_active":false,"role":"admin"}`}{
  store:=&patchStore{};resp:=activationHTTP(t,store,body)
  if resp.Code!=400 || store.called!=0 {t.Fatalf("%s: %d",body,resp.Code)}
 }
}
