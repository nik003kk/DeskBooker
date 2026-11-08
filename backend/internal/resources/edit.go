package resources

import (
    "context"
    "encoding/json"
    "errors"
    "io"
    "mime"
    "net/http"
    "strings"
    "unicode"

    "github.com/gin-gonic/gin"
    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgconn"
    "github.com/nik003kk/DeskBooker/backend/internal/models"
)

type Editor interface {
    UpdateResourceWithFeatures(context.Context,string,models.ResourcePatch)(models.Resource,[]string,error)
}

type patchRequest struct {
    Name *string `json:"name"`
    Type *models.ResourceType `json:"type"`
    Floor json.RawMessage `json:"floor"`
    Capacity *int `json:"capacity"`
    IsActive *bool `json:"is_active"`
    Features *[]string `json:"features"`
}

func validFeatureValues(features []string) bool {
    if len(features)>20 { return false }
    seen:=map[string]bool{}
    for _,name:=range features {
        if name=="" || strings.TrimSpace(name)!=name || len([]rune(name))>100 { return false }
        for _,ch:=range name { if unicode.IsControl(ch) {return false} }
        lower:=strings.ToLower(name)
        if seen[lower] {return false}
        seen[lower]=true
    }
    return true
}

func EditHandler(store Editor) gin.HandlerFunc {
    return func(c *gin.Context) {
        id:=c.Param("id")
        if !uuidPattern.MatchString(id) {
            c.JSON(400,gin.H{"error":gin.H{"code":"invalid_resource_id","message":"Invalid resource ID"}});return
        }
        mediaType,_,err:=mime.ParseMediaType(c.GetHeader("Content-Type"))
        if err!=nil || mediaType!="application/json" {
            c.JSON(415,gin.H{"error":gin.H{"code":"unsupported_media_type","message":"Expected application/json"}});return
        }
        dec:=json.NewDecoder(io.LimitReader(c.Request.Body,8193))
        dec.DisallowUnknownFields()
        var request patchRequest
        if err:=dec.Decode(&request);err!=nil{
            c.JSON(400,gin.H{"error":gin.H{"code":"invalid_json","message":"Invalid resource JSON"}});return
        }
        var extra any
        if !errors.Is(dec.Decode(&extra),io.EOF) {
            c.JSON(400,gin.H{"error":gin.H{"code":"invalid_json","message":"Only one JSON document is allowed"}});return
        }
        patch:=models.ResourcePatch{Name:request.Name,Type:request.Type,Capacity:request.Capacity,IsActive:request.IsActive}
        if request.Floor!=nil {
            patch.FloorSet=true
            if string(request.Floor)!="null" {
                var n int
                if json.Unmarshal(request.Floor,&n)!=nil || n<0 {
                    c.JSON(400,gin.H{"error":gin.H{"code":"invalid_resource","message":"Invalid floor"}});return
                }
                patch.Floor=&n
            }
        }
        if request.Features!=nil {
            patch.FeaturesSet=true
            patch.Features=*request.Features
            if !validFeatureValues(patch.Features) {
                c.JSON(400,gin.H{"error":gin.H{"code":"invalid_features","message":"Invalid resource equipment"}});return
            }
        }
        if patch.Name==nil && patch.Type==nil && !patch.FloorSet &&
            patch.Capacity==nil && patch.IsActive==nil && !patch.FeaturesSet {
            c.JSON(400,gin.H{"error":gin.H{"code":"empty_update","message":"No fields to update"}});return
        }
        if patch.Name!=nil && (*patch.Name=="" || strings.TrimSpace(*patch.Name)!=*patch.Name || len([]rune(*patch.Name))>150) ||
            patch.Capacity!=nil && *patch.Capacity<=0 ||
            patch.Type!=nil && *patch.Type!=models.ResourceDesk && *patch.Type!=models.ResourceMeetingRoom && *patch.Type!=models.ResourceParkingSpot {
            c.JSON(400,gin.H{"error":gin.H{"code":"invalid_resource","message":"Invalid resource fields"}});return
        }
        resource,features,err:=store.UpdateResourceWithFeatures(c.Request.Context(),id,patch)
        if errors.Is(err,pgx.ErrNoRows){
            c.JSON(404,gin.H{"error":gin.H{"code":"resource_not_found","message":"Resource not found"}});return
        }
        var pgErr *pgconn.PgError
        if errors.As(err,&pgErr) && pgErr.Code=="23505" {
            c.JSON(409,gin.H{"error":gin.H{"code":"resource_exists","message":"Resource name already exists"}});return
        }
        if err!=nil{
            c.JSON(http.StatusInternalServerError,gin.H{"error":gin.H{"code":"resource_update_failed","message":"Could not update resource"}});return
        }
        c.Header("Cache-Control","no-store")
        c.JSON(200,gin.H{"resource":resource.Public(),"features":features})
    }
}
