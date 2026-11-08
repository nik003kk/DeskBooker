// Package resources provides authenticated APIs for reservable resources.
package resources

import (
    "context"
    "encoding/json"
    "errors"
    "io"
    "mime"
    "net/http"
    "strings"

    "github.com/gin-gonic/gin"
    "github.com/nik003kk/DeskBooker/backend/internal/models"
)

type Creator interface {
    CreateResource(context.Context, models.CreateResourceInput) (models.Resource, error)
}

type createRequest struct {
    Name string `json:"name"`
    Type models.ResourceType `json:"type"`
    Floor *int `json:"floor"`
    Capacity int `json:"capacity"`
    IsActive *bool `json:"is_active"`
}

func CreateHandler(store Creator) gin.HandlerFunc {
    return func(c *gin.Context) {
        mediaType, _, err := mime.ParseMediaType(c.GetHeader("Content-Type"))
        if err != nil || mediaType != "application/json" {
            c.JSON(http.StatusUnsupportedMediaType, gin.H{"error":gin.H{"code":"unsupported_media_type","message":"Expected application/json"}})
            return
        }
        decoder:=json.NewDecoder(io.LimitReader(c.Request.Body,8193))
        decoder.DisallowUnknownFields()
        var request createRequest
        if decoder.Decode(&request)!=nil {
            c.JSON(400,gin.H{"error":gin.H{"code":"invalid_json","message":"Invalid resource JSON"}})
            return
        }
        var extra any
        if err:=decoder.Decode(&extra);!errors.Is(err,io.EOF){
            c.JSON(400,gin.H{"error":gin.H{"code":"invalid_json","message":"Only one JSON object allowed"}})
            return
        }
        if strings.TrimSpace(request.Name)=="" || len([]rune(request.Name))>150 ||
            request.Name!=strings.TrimSpace(request.Name) || request.Capacity<=0 ||
            (request.Type!=models.ResourceDesk && request.Type!=models.ResourceMeetingRoom &&
                request.Type!=models.ResourceParkingSpot) {
            c.JSON(400,gin.H{"error":gin.H{"code":"invalid_resource","message":"Invalid resource fields"}})
            return
        }
        active:=true
        if request.IsActive!=nil {active=*request.IsActive}
        resource,err:=store.CreateResource(c.Request.Context(),models.CreateResourceInput{
            Name:request.Name,Type:request.Type,Floor:request.Floor,
            Capacity:request.Capacity,IsActive:active,
        })
        if err!=nil {
            c.JSON(500,gin.H{"error":gin.H{"code":"create_resource_failed","message":"Could not create resource"}})
            return
        }
        c.JSON(http.StatusCreated,gin.H{"resource":resource.Public()})
    }
}
