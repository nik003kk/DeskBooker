package auth

import (
    "context"
    "encoding/json"
    "errors"
    "io"
    "mime"
    "net/http"
    "strings"

    "github.com/gin-gonic/gin"
    "github.com/jackc/pgx/v5"
    "github.com/nik003kk/DeskBooker/backend/internal/models"
)
type ProfileUpdater interface {
    UpdateUserProfile(context.Context, string, models.UpdateUserProfileInput) (models.User, error)
}
type profilePatch struct {
    FirstName *string `json:"first_name"`
    LastName *string `json:"last_name"`
    Department json.RawMessage `json:"department"`
}

// UpdateProfileHandler disallows email, role, status and password edits.
// Department supports clearing with JSON null.
func UpdateProfileHandler(store ProfileUpdater) gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Header("Cache-Control", "no-store")
        id := c.GetString("auth_user_id")
        if id == "" { responseError(c, 401, "unauthorized", "Authentication required"); return }
        mediaType, _, mediaErr := mime.ParseMediaType(c.GetHeader("Content-Type"))
        if mediaErr != nil || mediaType != "application/json" {
            responseError(c, http.StatusUnsupportedMediaType, "unsupported_media_type", "Expected application/json")
            return
        }
        decoder := json.NewDecoder(io.LimitReader(c.Request.Body, maxRegistrationBody+1))
        decoder.DisallowUnknownFields()
        var req profilePatch
        if decoder.Decode(&req) != nil {
            responseError(c, 400, "invalid_json", "Invalid profile JSON"); return
        }
        var extra any
        if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
            responseError(c, 400, "invalid_json", "Only one JSON object is allowed"); return
        }
        input := models.UpdateUserProfileInput{FirstName:req.FirstName, LastName:req.LastName}
        if req.Department != nil {
            input.DepartmentSet = true
            if string(req.Department) != "null" {
                var department string
                if json.Unmarshal(req.Department, &department) != nil ||
                    strings.TrimSpace(department) == "" || department != strings.TrimSpace(department) ||
                    len([]rune(department)) > 100 {
                    responseError(c, 400, "invalid_department", "Invalid department"); return
                }
                input.Department = &department
            }
        }
        if input.FirstName == nil && input.LastName == nil && !input.DepartmentSet {
            responseError(c, 400, "empty_update", "Provide at least one editable field"); return
        }
        if input.FirstName != nil && !validName(*input.FirstName) ||
            input.LastName != nil && !validName(*input.LastName) {
            responseError(c, 400, "invalid_name", "Invalid name"); return
        }
        updated, err := store.UpdateUserProfile(c.Request.Context(), id, input)
        if errors.Is(err, pgx.ErrNoRows) {
            responseError(c, 401, "unauthorized", "Authentication required"); return
        }
        if err != nil {
            responseError(c, 500, "profile_update_failed", "Could not update profile"); return
        }
        c.JSON(http.StatusOK, gin.H{"user":updated.Public()})
    }
}
