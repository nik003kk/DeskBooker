package resources

import (
    "context"
    "encoding/json"
    "errors"
    "io"
    "mime"
    "net/http"
    "regexp"
    "strings"
    "unicode"

    "github.com/gin-gonic/gin"
    "github.com/jackc/pgx/v5"
    "github.com/nik003kk/DeskBooker/backend/internal/models"
)

type FeatureWriter interface {
    AddFeatures(context.Context, string, []string) ([]models.ResourceFeature, error)
}
var uuidPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
type featureRequest struct {
    Features []string `json:"features"`
}

func AddFeaturesHandler(store FeatureWriter) gin.HandlerFunc {
    return func(c *gin.Context) {
        resourceID:=c.Param("id")
        if !uuidPattern.MatchString(resourceID) {
            c.JSON(400,gin.H{"error":gin.H{"code":"invalid_resource_id","message":"Invalid resource ID"}});return
        }
        mediaType,_,err:=mime.ParseMediaType(c.GetHeader("Content-Type"))
        if err!=nil || mediaType!="application/json" {
            c.JSON(415,gin.H{"error":gin.H{"code":"unsupported_media_type","message":"Expected application/json"}});return
        }
        decoder:=json.NewDecoder(io.LimitReader(c.Request.Body,8193))
        decoder.DisallowUnknownFields()
        var req featureRequest
        if decoder.Decode(&req)!=nil {
            c.JSON(400,gin.H{"error":gin.H{"code":"invalid_json","message":"Invalid JSON"}});return
        }
        var extra any
        if err:=decoder.Decode(&extra); !errors.Is(err,io.EOF) {
            c.JSON(400,gin.H{"error":gin.H{"code":"invalid_json","message":"Only one object allowed"}});return
        }
        if len(req.Features)==0 || len(req.Features)>20 {
            c.JSON(400,gin.H{"error":gin.H{"code":"invalid_features","message":"Provide 1 to 20 features"}});return
        }
        seen:=map[string]bool{}
        for _,name:=range req.Features {
            if name=="" || strings.TrimSpace(name)!=name || len([]rune(name))>100 {
                c.JSON(400,gin.H{"error":gin.H{"code":"invalid_features","message":"Invalid feature name"}});return
            }
            for _,char:=range name {
                if unicode.IsControl(char) {
                    c.JSON(400,gin.H{"error":gin.H{"code":"invalid_features","message":"Invalid feature name"}});return
                }
            }
            folded:=strings.ToLower(name)
            if seen[folded] {
                c.JSON(400,gin.H{"error":gin.H{"code":"duplicate_feature","message":"Duplicate feature"}});return
            }
            seen[folded]=true
        }
        features,err:=store.AddFeatures(c.Request.Context(),resourceID,req.Features)
        if errors.Is(err,pgx.ErrNoRows) {
            c.JSON(404,gin.H{"error":gin.H{"code":"resource_not_found","message":"Resource not found"}});return
        }
        if err!=nil {
            c.JSON(500,gin.H{"error":gin.H{"code":"features_failed","message":"Could not save features"}});return
        }
        c.JSON(http.StatusCreated,gin.H{"features":features})
    }
}
