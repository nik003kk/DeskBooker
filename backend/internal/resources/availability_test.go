package resources

import (
 "context"
 "errors"
 "net/http/httptest"
 "strings"
 "testing"
 "time"

 "github.com/gin-gonic/gin"
 "github.com/jackc/pgx/v5"
)
type availabilityStore struct{result bool;called int;err error}
func(s *availabilityStore)IsResourceAvailable(context.Context,string,time.Time,time.Time)(bool,error){
 s.called++;return s.result,s.err
}
func availabilityHTTP(t *testing.T,s *availabilityStore,query string)*httptest.ResponseRecorder{
 t.Helper();gin.SetMode(gin.TestMode);r:=gin.New()
 r.GET("/resources/:id/availability",AvailabilityHandler(s))
 w:=httptest.NewRecorder()
 r.ServeHTTP(w,httptest.NewRequest("GET","/resources/"+resourceUUID+"/availability"+query,nil))
 return w
}
func TestAvailabilityResponse(t *testing.T){
 for _,tc:=range []struct{available bool;expected string}{
  {true,`"available":true`},{false,`"available":false`},
 }{
  store:=&availabilityStore{result:tc.available}
  resp:=availabilityHTTP(t,store,"?date=2026-11-09&start=09:00&end=17:00")
  if resp.Code!=200 || store.called!=1 || !strings.Contains(resp.Body.String(),tc.expected){
   t.Fatalf("%d %s",resp.Code,resp.Body.String())
  }
 }
}
func TestAvailabilityRejectsInvalidWindow(t *testing.T){
 for _,query:=range []string{"","?date=2026-11-09","?date=2026-11-09&start=09:00&end=09:00",
 "?date=2026-02-30&start=09:00&end=17:00","?date=2026-11-09&start=17:00&end=09:00",
 "?date=2026-03-29&start=02:30&end=04:00"}{
  store:=&availabilityStore{}
  w:=availabilityHTTP(t,store,query)
  if w.Code!=400 || store.called!=0{t.Fatalf("%s: %d",query,w.Code)}
 }
}
func TestAvailabilityErrorsAreSanitized(t *testing.T){
 store:=&availabilityStore{err:pgx.ErrNoRows}
 w:=availabilityHTTP(t,store,"?date=2026-11-09&start=09:00&end=10:00")
 if w.Code!=404{t.Fatal(w.Code)}
 store.err=errors.New("secret db")
 w=availabilityHTTP(t,store,"?date=2026-11-09&start=09:00&end=10:00")
 if w.Code!=500 || strings.Contains(w.Body.String(),"secret db"){t.Fatal(w.Body.String())}
}
