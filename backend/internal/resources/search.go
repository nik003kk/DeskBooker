package resources

import(
 "context"
 "net/http"
 "strings"

 "github.com/gin-gonic/gin"
 "github.com/nik003kk/DeskBooker/backend/internal/models"
)

type Searcher interface {
 SearchResources(context.Context,models.ResourceSearch)([]models.Resource,error)
}

// FeatureLister is optional for older search mocks, but implemented by
// the production repository. Search cards include actual stored equipment.
type FeatureLister interface {
 ListFeatures(context.Context,string)([]models.ResourceFeature,error)
}

type searchResourceResponse struct {
 models.ResourceDTO
 Features []string `json:"features"`
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
   from,to,err:=parseAvailabilityWindow(date,start,end)
   if err!=nil {
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
  result:=make([]searchResourceResponse,0,len(found))
  features,hasFeatures:=store.(FeatureLister)
  for _,r:=range found {
   item:=searchResourceResponse{ResourceDTO:r.Public(),Features:[]string{}}
   if hasFeatures {
    stored,err:=features.ListFeatures(c.Request.Context(),r.ID)
    if err!=nil {
     c.JSON(500,gin.H{"error":gin.H{"code":"search_failed","message":"Could not load equipment"}})
     return
    }
    for _,f:=range stored {item.Features=append(item.Features,f.Feature)}
   }
   result=append(result,item)
  }
  c.Header("Cache-Control","no-store")
  c.JSON(http.StatusOK,gin.H{"resources":result})
 }
}
