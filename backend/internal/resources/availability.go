package resources

import (
 "context"
 "errors"
 "net/http"
 "time"

 "github.com/gin-gonic/gin"
 "github.com/jackc/pgx/v5"
)

type AvailabilityChecker interface {
 IsResourceAvailable(context.Context,string,time.Time,time.Time)(bool,error)
}

var ErrInvalidWindow = errors.New("invalid reservation window")

// parseAvailabilityWindow parses office-local date/time and rejects
// impossible dates, missing fields and nonexistent DST transitions.
func parseAvailabilityWindow(date,start,end string)(time.Time,time.Time,error){
 if len(date)!=10 || len(start)!=5 || len(end)!=5 {
  return time.Time{},time.Time{},ErrInvalidWindow
 }
 loc,err:=time.LoadLocation("Europe/Ljubljana")
 if err!=nil {return time.Time{},time.Time{},err}
 from,e1:=time.ParseInLocation("2006-01-02 15:04",date+" "+start,loc)
 to,e2:=time.ParseInLocation("2006-01-02 15:04",date+" "+end,loc)
 if e1!=nil || e2!=nil || !to.After(from) || to.Sub(from)>24*time.Hour ||
  from.In(loc).Format("2006-01-02 15:04")!=date+" "+start ||
  to.In(loc).Format("2006-01-02 15:04")!=date+" "+end {
  return time.Time{},time.Time{},ErrInvalidWindow
 }
 return from,to,nil
}

func AvailabilityHandler(store AvailabilityChecker)gin.HandlerFunc {
 return func(c *gin.Context){
  id:=c.Param("id")
  if !uuidPattern.MatchString(id) {
   c.JSON(400,gin.H{"error":gin.H{"code":"invalid_resource_id","message":"Invalid resource ID"}});return
  }
  from,to,err:=parseAvailabilityWindow(c.Query("date"),c.Query("start"),c.Query("end"))
  if err!=nil {
   c.JSON(400,gin.H{"error":gin.H{"code":"invalid_time","message":"Use valid date, start and end"}});return
  }
  available,err:=store.IsResourceAvailable(c.Request.Context(),id,from,to)
  if errors.Is(err,pgx.ErrNoRows){
   c.JSON(404,gin.H{"error":gin.H{"code":"resource_not_found","message":"Resource not found"}});return
  }
  if err!=nil{
   c.JSON(http.StatusInternalServerError,gin.H{"error":gin.H{"code":"availability_failed","message":"Could not check resource availability"}});return
  }
  c.Header("Cache-Control","no-store")
  c.JSON(http.StatusOK,gin.H{"resource_id":id,"available":available,"starts_at":from.UTC(),"ends_at":to.UTC()})
 }
}
