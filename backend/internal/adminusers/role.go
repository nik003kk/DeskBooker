package adminusers

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/nik003kk/DeskBooker/backend/internal/models"
	"github.com/nik003kk/DeskBooker/backend/internal/repository"
	"io"
	"mime"
	"net/http"
	"strings"
	"time"
)

type RoleChanger interface {
	ChangeUserRole(context.Context, string, string, models.UserRole) (models.UserDTO, error)
}

func userUUID(id string) bool {
	if len(id) != 36 {
		return false
	}
	for i, c := range id {
		switch i {
		case 8, 13, 18, 23:
			if c != '-' {
				return false
			}
		default:
			if !(c >= '0' && c <= '9' || c >= 'a' && c <= 'f' || c >= 'A' && c <= 'F') {
				return false
			}
		}
	}
	return true
}
func ChangeRoleHandler(store RoleChanger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Cache-Control", "no-store")
		actor := c.GetString("auth_user_id")
		if actor == "" {
			failure(c, 401, "unauthorized", "Authentication required")
			return
		}
		target := strings.ToLower(c.Param("id"))
		if !userUUID(target) {
			failure(c, 400, "invalid_user", "Invalid user ID")
			return
		}
		media, _, err := mime.ParseMediaType(c.GetHeader("Content-Type"))
		if err != nil || media != "application/json" {
			failure(c, 415, "unsupported_media_type", "Expected application/json")
			return
		}
		var request struct {
			Role string `json:"role"`
		}
		dec := json.NewDecoder(http.MaxBytesReader(c.Writer, c.Request.Body, 1024))
		dec.DisallowUnknownFields()
		if err = dec.Decode(&request); err != nil {
			failure(c, 400, "invalid_json", "Invalid role JSON")
			return
		}
		var extra any
		if err = dec.Decode(&extra); !errors.Is(err, io.EOF) {
			failure(c, 400, "invalid_json", "Only one JSON object allowed")
			return
		}
		role, ok := models.ParseUserRole(request.Role)
		if !ok {
			failure(c, 400, "invalid_role", "Expected EMPLOYEE, OFFICE_MANAGER or ADMIN")
			return
		}
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()
		saved, err := store.ChangeUserRole(ctx, actor, target, role)
		switch {
		case errors.Is(err, repository.ErrSelfRoleChange):
			failure(c, 403, "self_role_protected", "Cannot change your own role")
		case errors.Is(err, repository.ErrAdminRequired):
			failure(c, 403, "forbidden", "Active administrator required")
		case errors.Is(err, pgx.ErrNoRows):
			failure(c, 404, "user_not_found", "User not found")
		case err != nil:
			failure(c, 500, "role_update_failed", "Could not change user role")
		default:
			c.JSON(200, gin.H{"user": saved})
		}
	}
}
