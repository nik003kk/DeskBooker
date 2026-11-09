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
  !strings.Contains(db.sql,"b.status IN ('PENDING','CONFIRMED','CHECKED_IN')") ||
  !strings.Contains(db.sql,"r.is_active=TRUE") ||
  strings.Contains(db.sql,"Monitor") ||
  !strings.Contains(db.sql,"b.starts_at < $3") ||
  !strings.Contains(db.sql,"b.ends_at > $2"){
  t.Fatalf("SQL not parameterized or missing window overlap: %q args=%#v",db.sql,db.args)
 }
}

func TestSearchResourcesRequiresAllEquipmentFilters(t *testing.T) {
 db:=&stubDB{err:errors.New("no database")}
 _,_ = New(db).SearchResources(context.Background(),models.ResourceSearch{
  Features:[]string{"Monitor","USB-C"},
 })
 if !strings.Contains(db.sql,"FROM unnest($4::text[]) AS wanted(feature)") ||
  !strings.Contains(db.sql,"WHERE NOT EXISTS (SELECT 1 FROM public.resource_features f") ||
  !strings.Contains(db.sql,"LOWER(f.feature)=wanted.feature") ||
  strings.Contains(db.sql,"LOWER(f.feature)=ANY") {
  t.Fatalf("must match every feature, got query: %s",db.sql)
 }
 values,ok:=db.args[3].([]string)
 if !ok || len(values)!=2 || values[0]!="monitor" || values[1]!="usb-c" {
  t.Fatalf("unexpected normalized filter values: %#v",db.args[3])
 }
}
func TestSearchResourcesNoEquipmentFilterStillMatches(t *testing.T) {
 db:=&stubDB{err:errors.New("offline")}
 _,_ = New(db).SearchResources(context.Background(),models.ResourceSearch{})
 if !strings.Contains(db.sql,"unnest($4::text[])") {t.Fatal("empty filter must be treated as vacuous all-of match")}
 if values,ok:=db.args[3].([]string);!ok || len(values)!=0 {t.Fatalf("unexpected empty filters: %#v",db.args[3])}
}

