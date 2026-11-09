package repository

import(
 "context"
 "errors"
 "strings"
 "testing"
 "time"
)
func TestAvailabilitySQLExcludesInactiveAndBookedResources(t *testing.T){
 db:=&stubDB{err:errors.New("no row")}
 from:=time.Date(2026,11,9,9,0,0,0,time.UTC)
 _,err:=New(db).IsResourceAvailable(context.Background(),"resource-id",from,from.Add(time.Hour))
 if err==nil{t.Fatal("expected query error")}
 for _,fragment:=range []string{
  "r.is_active","b.status IN ('PENDING','CONFIRMED','CHECKED_IN')",
  "b.starts_at < $3","b.ends_at > $2","r.id=$1",
 }{
  if !strings.Contains(db.sql,fragment){t.Fatalf("missing %s in %s",fragment,db.sql)}
 }
 if len(db.args)!=3 || db.args[0]!="resource-id"{t.Fatalf("unexpected args %#v",db.args)}
}

