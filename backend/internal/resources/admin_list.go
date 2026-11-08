package resources

import (
  "context"
  "net/http"

  "github.com/gin-gonic/gin"
  "github.com/nik003kk/DeskBooker/backend/internal/models"
)

type AdminLister interface {
  ListAllResources(context.Context) ([]models.Resource,error)
}
func AdminListHandler(store AdminLister) gin.HandlerFunc {
  return func(c *gin.Context) {
    items,err:=store.ListAllResources(c.Request.Context())
    if err!=nil {
      c.JSON(http.StatusInternalServerError,gin.H{"error":gin.H{"code":"list_resources_failed","message":"Could not list resources"}})
      return
    }
    result:=make([]models.ResourceDTO,0,len(items))
    for _,item:=range items {result=append(result,item.Public())}
    c.Header("Cache-Control","no-store")
    c.JSON(http.StatusOK,gin.H{"resources":result})
  }
}
