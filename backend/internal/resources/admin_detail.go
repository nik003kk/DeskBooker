package resources

import (
 "context"
 "errors"
 "net/http"
 "github.com/gin-gonic/gin"
 "github.com/jackc/pgx/v5"
 "github.com/nik003kk/DeskBooker/backend/internal/models"
)
type AdminReader interface {
 GetResource(context.Context,string)(models.Resource,error)
 ListFeatures(context.Context,string)([]models.ResourceFeature,error)
}
func AdminDetailHandler(store AdminReader)gin.HandlerFunc{
 return func(c *gin.Context){
  id:=c.Param("id")
  if !uuidPattern.MatchString(id){c.JSON(400,gin.H{"error":gin.H{"code":"invalid_resource_id","message":"Invalid resource ID"}});return}
  item,err:=store.GetResource(c.Request.Context(),id)
  if errors.Is(err,pgx.ErrNoRows){c.JSON(404,gin.H{"error":gin.H{"code":"resource_not_found","message":"Resource not found"}});return}
  if err!=nil{c.JSON(500,gin.H{"error":gin.H{"code":"resource_unavailable","message":"Could not load resource"}});return}
  features,err:=store.ListFeatures(c.Request.Context(),id)
  if err!=nil{c.JSON(500,gin.H{"error":gin.H{"code":"resource_unavailable","message":"Could not load resource"}});return}
  result:=make([]string,0,len(features))
  for _,feature:=range features{result=append(result,feature.Feature)}
  c.Header("Cache-Control","no-store")
  c.JSON(http.StatusOK,gin.H{"resource":item.Public(),"features":result})
 }
}
