package reservations

import(
 "errors"
 "testing"
 "time"
)
func TestReservationTimeValidation(t *testing.T){
 now:=time.Date(2026,11,8,8,0,0,0,time.UTC)
 start:=now.Add(2*time.Hour)
 for _,tc:=range []struct{name string;from,to time.Time;valid bool}{
  {"valid",start,start.Add(time.Hour),true},
  {"boundary 24h",start,start.Add(24*time.Hour),true},
  {"reversed",start,start.Add(-time.Hour),false},
  {"equal",start,start,false},
  {"past",now.Add(-time.Minute),start,false},
  {"more than 24h",start,start.Add(25*time.Hour),false},
  {"seconds",start.Add(time.Second),start.Add(time.Hour),false},
  {"zero",time.Time{},start,false},
 }{
  t.Run(tc.name,func(t *testing.T){
   err:=ValidateWindow(now,tc.from,tc.to)
   if (err==nil)!=tc.valid{t.Fatalf("got %v, want valid=%v",err,tc.valid)}
   if !tc.valid&&!errors.Is(err,ErrInvalidReservationWindow){t.Fatal(err)}
  })
 }
}
