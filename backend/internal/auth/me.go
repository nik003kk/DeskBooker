package auth

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/nik003kk/DeskBooker/backend/internal/models"
)

type UserByIDFinder interface {
	GetUser(context.Context, string) (models.User, error)
}

// MeHandler reloads the user from PostgreSQL: a deleted user cannot keep
// accessing the application merely because an earlier token is still valid.
func MeHandler(store UserByIDFinder) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Cache-Control", "no-store")
		id := c.GetString("auth_user_id")
		if id == "" {
			responseError(c, http.StatusUnauthorized, "unauthorized", "Valid bearer token required")
			return
		}
		if value, present := c.Get(currentUserKey); present {
			if current, valid := value.(models.UserDTO); valid && current.ID == id {
				c.JSON(http.StatusOK, gin.H{"user": current})
				return
			}
		}
		user, err := store.GetUser(c.Request.Context(), id)
		if errors.Is(err, pgx.ErrNoRows) {
			responseError(c, http.StatusUnauthorized, "unauthorized", "Valid bearer token required")
			return
		}
		if err != nil {
			responseError(c, http.StatusInternalServerError, "profile_unavailable", "Could not load user")
			return
		}
		c.JSON(http.StatusOK, gin.H{"user": user.Public()})
	}
}
