package reservations

import(
 "context"
 "errors"
 "net/http"
 "net/http/httptest"
 "strings"
 "testing"
 "time"
 "github.com/gin-gonic/gin"
 "github.com/jackc/pgx/v5/pgconn"
 "github.com/nik003kk/DeskBooker/backend/internal/models"
)
const resourceID="11111111-1111-4111-8111-111111111111"
type fakeCreator struct{called int;input models.CreateReservationInput;err error}
func(s *fakeCreator)CreateConfirmedReservation(_ context.Context,in models.CreateReservationInput)(models.Reservation,error){
 s.called++;s.input=in
 if s.err!=nil{return models.Reservation{},s.err}
 return models.Reservation{ID:"reservation",ResourceID:in.ResourceID,UserID:in.UserID,
  StartsAt:in.StartsAt,EndsAt:in.EndsAt,Status:models.ReservationConfirmed},nil
}
func createTest(t *testing.T,store *fakeCreator,body string,authorized bool)*httptest.ResponseRecorder{
 t.Helper();gin.SetMode(gin.TestMode)
 router:=gin.New()
 router.POST("/api/reservations",func(c *gin.Context){
  if authorized{c.Set("auth_user_id","trusted-user")}
  c.Next()
 },CreateHandler(store))
 request:=httptest.NewRequest(http.MethodPost,"/api/reservations",strings.NewReader(body))
 request.Header.Set("Content-Type","application/json")
 response:=httptest.NewRecorder();router.ServeHTTP(response,request);return response
}
func TestCreateReservationUsesTrustedJWTSubject(t *testing.T){
 store:=&fakeCreator{}
 body:=`{"resource_id":"`+resourceID+`","starts_at":"2026-11-10T09:00:00+01:00","ends_at":"2026-11-10T10:00:00+01:00"}`
 response:=createTest(t,store,body,true)
 if response.Code!=201||store.called!=1||store.input.UserID!="trusted-user"||
  !strings.Contains(response.Body.String(),`"status":"CONFIRMED"`) {t.Fatalf("%d %s",response.Code,response.Body.String())}
 if noAuth:=createTest(t,store,body,false);noAuth.Code!=401||store.called!=1{t.Fatalf("unauthorized %d",noAuth.Code)}
}
func TestCreateReservationRejectsInjectedFields(t *testing.T){
 for _,body:=range []string{
  `{}`,`{"resource_id":"bad","starts_at":"2026-11-10T09:00:00Z","ends_at":"2026-11-10T10:00:00Z"}`,
  `{"resource_id":"`+resourceID+`","starts_at":"2026-11-10T10:00:00Z","ends_at":"2026-11-10T09:00:00Z"}`,
  `{"resource_id":"`+resourceID+`","starts_at":"2026-11-10T09:00:00Z","ends_at":"2026-11-10T10:00:00Z","user_id":"attacker"}`,
 }{
  store:=&fakeCreator{};r:=createTest(t,store,body,true)
  if r.Code!=400||store.called!=0{t.Fatalf("%s returned %d",body,r.Code)}
 }
}
func TestOldAndExcessivelyLongReservationRejected(t *testing.T){
 for _,body:=range []string{
  `{"resource_id":"`+resourceID+`","starts_at":"2020-01-01T09:00:00Z","ends_at":"2020-01-01T10:00:00Z"}`,
  `{"resource_id":"`+resourceID+`","starts_at":"2030-01-01T09:00:00Z","ends_at":"2030-01-03T09:00:00Z"}`,
  `{"resource_id":"`+resourceID+`","starts_at":"2030-01-01T09:00:01Z","ends_at":"2030-01-01T10:00:00Z"}`,
 }{
  store:=&fakeCreator{}
  response:=createTest(t,store,body,true)
  if response.Code!=400||store.called!=0{t.Fatalf("%s: status %d",body,response.Code)}
 }
 _=time.Now()
}
func TestConflictAndDBErrorMapping(t *testing.T){
 body:=`{"resource_id":"`+resourceID+`","starts_at":"2026-11-10T09:00:00Z","ends_at":"2026-11-10T10:00:00Z"}`
 for _,tc:=range []struct{err error;status int}{
  {&pgconn.PgError{Code:"23P01"},409},{errors.New("sensitive error"),500},
 }{
  r:=createTest(t,&fakeCreator{err:tc.err},body,true)
  if r.Code!=tc.status||strings.Contains(r.Body.String(),"sensitive error"){t.Fatal(r.Code,r.Body.String())}
 }
}
