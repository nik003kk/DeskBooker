package resources

import(
 "context"
 "net/http"
 "strings"
 "time"

 "github.com/gin-gonic/gin"
 "github.com/nik003kk/DeskBooker/backend/internal/models"
)

type Searcher interface {
 SearchResources(context.Context,models.ResourceSearch)([]models.Resource,error)
}

// SearchHandler accepts optional type, date/start/end and feature filters.
// Times are interpreted as Europe/Ljubljana local business time.
func SearchHandler(store Searcher)gin.HandlerFunc{
 return func(c *gin.Context){
  var filter models.ResourceSearch
  if raw:=c.Query("type");raw!="" {
   kind:=models.ResourceType(raw)
   if kind!=models.ResourceDesk && kind!=models.ResourceMeetingRoom && kind!=models.ResourceParkingSpot {
    c.JSON(400,gin.H{"error":gin.H{"code":"invalid_type","message":"Unsupported resource type"}});return
   }
   filter.Type=&kind
  }
  date,start,end:=c.Query("date"),c.Query("start"),c.Query("end")
  if date!="" || start!="" || end!="" {
   if date=="" || start=="" || end=="" {
    c.JSON(400,gin.H{"error":gin.H{"code":"invalid_time","message":"Provide date, start and end together"}});return
   }
   if len(date)!=10 || len(start)!=5 || len(end)!=5 {
    c.JSON(400,gin.H{"error":gin.H{"code":"invalid_time","message":"Use YYYY-MM-DD and HH:MM"}});return
   }
   location,err:=time.LoadLocation("Europe/Ljubljana")
   if err!=nil {c.JSON(500,gin.H{"error":gin.H{"code":"timezone_unavailable","message":"Time zone unavailable"}});return}
   from,errFrom:=time.ParseInLocation("2006-01-02 15:04",date+" "+start,location)
   to,errTo:=time.ParseInLocation("2006-01-02 15:04",date+" "+end,location)
   if errFrom!=nil || errTo!=nil || !to.After(from) || to.Sub(from)>24*time.Hour ||
      from.In(location).Format("2006-01-02 15:04")!=date+" "+start ||
      to.In(location).Format("2006-01-02 15:04")!=date+" "+end {
    c.JSON(400,gin.H{"error":gin.H{"code":"invalid_time","message":"Invalid reservation time window"}});return
   }
   filter.StartsAt=&from
   filter.EndsAt=&to
  }
  rawFeatures:=c.Query("features")
  if rawFeatures!="" {
   split:=strings.Split(rawFeatures,",")
   if len(split)>20 {
    c.JSON(400,gin.H{"error":gin.H{"code":"invalid_features","message":"Too many features"}});return
   }
   filter.Features=make([]string,0,len(split))
   for _,f:=range split {
    f=strings.TrimSpace(f)
    if f=="" {c.JSON(400,gin.H{"error":gin.H{"code":"invalid_features","message":"Empty feature"}});return}
    filter.Features=append(filter.Features,f)
   }
   if !validFeatureValues(filter.Features) {
    c.JSON(400,gin.H{"error":gin.H{"code":"invalid_features","message":"Invalid equipment filters"}});return
   }
  }
  found,err:=store.SearchResources(c.Request.Context(),filter)
  if err!=nil{
   c.JSON(500,gin.H{"error":gin.H{"code":"search_failed","message":"Could not search resources"}});return
  }
  result:=make([]models.ResourceDTO,0,len(found))
  for _,r:=range found {result=append(result,r.Public())}
  c.JSON(http.StatusOK,gin.H{"resources":result})
 }
}
