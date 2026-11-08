package resources
import(
 "context"
 "errors"
 "net/http"
 "net/http/httptest"
 "strings"
 "testing"

 "github.com/gin-gonic/gin"
 "github.com/nik003kk/DeskBooker/backend/internal/models"
)
type searchStore struct{called int;filter models.ResourceSearch;err error; featureErr error}
func(s *searchStore)ListFeatures(_ context.Context,_ string)([]models.ResourceFeature,error){
 if s.featureErr!=nil{return nil,s.featureErr}
 return []models.ResourceFeature{{ResourceID:"resource1",Feature:"Monitor"},{ResourceID:"resource1",Feature:"USB-C"}},nil
}
func(s *searchStore)SearchResources(_ context.Context,filter models.ResourceSearch)([]models.Resource,error){
 s.called++;s.filter=filter
 if s.err!=nil{return nil,s.err}
 return []models.Resource{{ID:"resource1",Name:"Desk",Type:models.ResourceDesk,Capacity:1,IsActive:true}},nil
}
func searchHTTP(t *testing.T,s *searchStore,url string)*httptest.ResponseRecorder{
 t.Helper();gin.SetMode(gin.TestMode);r:=gin.New();r.GET("/resources",SearchHandler(s))
 request:=httptest.NewRequest(http.MethodGet,url,nil);w:=httptest.NewRecorder();r.ServeHTTP(w,request);return w
}
func TestSearchParsesFiltersAndReturnsResources(t *testing.T){
 s:=&searchStore{}
 response:=searchHTTP(t,s,"/resources?date=2026-11-09&start=09:00&end=17:00&type=DESK&features=Monitor,USB-C")
 if response.Code!=200 || s.called!=1 || s.filter.StartsAt==nil || s.filter.EndsAt==nil ||
   s.filter.Type==nil || *s.filter.Type!=models.ResourceDesk || len(s.filter.Features)!=2 || !strings.Contains(response.Body.String(),`"features":["Monitor","USB-C"]`) {
  t.Fatalf("status=%d calls=%d body=%s",response.Code,s.called,response.Body.String())
 }
}
func TestSearchRejectsInvalidQuery(t *testing.T){
 cases:=[]string{
  "/resources?date=2026-11-09","/resources?start=09:00",
  "/resources?date=2026-11-09&start=18:00&end=09:00",
  "/resources?date=2026-11-09&start=02:70&end=10:00",
  "/resources?date=bad&start=09:00&end=17:00",
  "/resources?type=INVALID","/resources?features=Monitor,monitor",
 }
 for _,url:=range cases {
  s:=&searchStore{};w:=searchHTTP(t,s,url)
  if w.Code!=400 || s.called!=0 {t.Fatalf("%s: code=%d",url,w.Code)}
 }
}
func TestSearchSanitizesStoreFailure(t *testing.T){
 s:=&searchStore{err:errors.New("private db secret")}
 w:=searchHTTP(t,s,"/resources")
 if w.Code!=500 || strings.Contains(w.Body.String(),"private db secret"){t.Fatal(w.Body.String())}
}

func TestSearchFeatureLoadErrorIsSanitized(t *testing.T) {
 s:=&searchStore{featureErr:errors.New("private feature DB error")}
 response:=searchHTTP(t,s,"/resources")
 if response.Code!=500 || strings.Contains(response.Body.String(),"private feature DB error"){
  t.Fatalf("unexpected error response: %d %s",response.Code,response.Body.String())
 }
}
