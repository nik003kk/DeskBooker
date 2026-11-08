package repository
import(
 "context"
 "errors"
 "strings"
 "testing"
 "time"
 "github.com/nik003kk/DeskBooker/backend/internal/models"
)
func TestSearchResourcesUsesBoundFilters(t *testing.T){
 db:=&stubDB{err:errors.New("fake DB")}
 kind:=models.ResourceDesk
 from:=time.Date(2026,11,9,9,0,0,0,time.UTC)
 to:=from.Add(8*time.Hour)
 _,err:=New(db).SearchResources(context.Background(),models.ResourceSearch{
  Type:&kind,StartsAt:&from,EndsAt:&to,Features:[]string{"Monitor","USB-C"},
 })
 if err==nil{t.Fatal("expected query failure")}
 if len(db.args)!=4 || db.args[0]!="DESK" ||
  !strings.Contains(db.sql,"b.status IN ('PENDING','CONFIRMED')") ||
  !strings.Contains(db.sql,"r.is_active=TRUE") ||
  strings.Contains(db.sql,"Monitor") ||
  !strings.Contains(db.sql,"b.starts_at < $3") ||
  !strings.Contains(db.sql,"b.ends_at > $2"){
  t.Fatalf("SQL not parameterized or missing window overlap: %q args=%#v",db.sql,db.args)
 }
}
