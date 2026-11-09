package resources

import (
 "encoding/json"
 "errors"
 "io"
 "mime"
 "net/http"

 "github.com/gin-gonic/gin"
 "github.com/jackc/pgx/v5"
 "github.com/nik003kk/DeskBooker/backend/internal/models"
)

type activationRequest struct{ IsActive *bool `json:"is_active"` }

// ActivationHandler never deletes resource rows, preserving reservation history.
func ActivationHandler(store Editor) gin.HandlerFunc {
 return func(c *gin.Context){
  if !uuidPattern.MatchString(c.Param("id")) {
   c.JSON(400,gin.H{"error":gin.H{"code":"invalid_resource_id","message":"Invalid resource ID"}});return
  }
  mediaType,_,err:=mime.ParseMediaType(c.GetHeader("Content-Type"))
  if err!=nil || mediaType!="application/json" {
   c.JSON(415,gin.H{"error":gin.H{"code":"unsupported_media_type","message":"Expected application/json"}});return
  }
  decoder:=json.NewDecoder(io.LimitReader(c.Request.Body,8193))
  decoder.DisallowUnknownFields()
  var req activationRequest
  if err:=decoder.Decode(&req);err!=nil {
   c.JSON(400,gin.H{"error":gin.H{"code":"invalid_json","message":"Invalid JSON"}});return
  }
  var extra any
  if !errors.Is(decoder.Decode(&extra),io.EOF) || req.IsActive==nil {
   c.JSON(400,gin.H{"error":gin.H{"code":"invalid_activation","message":"is_active must be a boolean"}});return
  }
  item,_,err:=store.UpdateResourceWithFeatures(c.Request.Context(),c.Param("id"),models.ResourcePatch{IsActive:req.IsActive})
  if errors.Is(err,pgx.ErrNoRows) {
   c.JSON(404,gin.H{"error":gin.H{"code":"resource_not_found","message":"Resource not found"}});return
  }
  if err!=nil {
   c.JSON(http.StatusInternalServerError,gin.H{"error":gin.H{"code":"activation_failed","message":"Could not change resource status"}});return
  }
  c.Header("Cache-Control","no-store")
  c.JSON(200,gin.H{"resource":item.Public()})
 }
}
