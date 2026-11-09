package auth

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/nik003kk/DeskBooker/backend/internal/models"
	"strings"
)

type RoleLookup interface {
	GetUser(context.Context, string) (models.User, error)
}

// RequireCurrentJWT loads a verified user once per request. Isolated routes
// without it still fall back to a live lookup; JWT role claims are never used.
func RequireRoles(store RoleLookup, roles ...models.UserRole) gin.HandlerFunc {
	allowed := make(map[models.UserRole]bool, len(roles))
	for _, role := range roles {
		allowed[role] = true
	}
	return func(c *gin.Context) {
		id := c.GetString("auth_user_id")
		if id == "" {
			responseError(c, 401, "unauthorized", "Authentication required")
			c.Abort()
			return
		}
		var user models.UserDTO
		if value, present := c.Get(currentUserKey); present {
			var valid bool
			user, valid = value.(models.UserDTO)
			if !valid || !strings.EqualFold(user.ID, id) {
				responseError(c, 401, "unauthorized", "Authentication required")
				c.Abort()
				return
			}
		} else {
			saved, err := store.GetUser(c.Request.Context(), id)
			if errors.Is(err, pgx.ErrNoRows) {
				responseError(c, 401, "unauthorized", "Authentication required")
				c.Abort()
				return
			}
			if err != nil {
				responseError(c, 500, "authorization_failed", "Could not authorize request")
				c.Abort()
				return
			}
			user = saved.Public()
		}
		if user.Status != models.UserActive || !allowed[user.Role] {
			responseError(c, 403, "forbidden", "Insufficient permissions")
			c.Abort()
			return
		}
		c.Next()
	}
}
