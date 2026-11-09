package adminusers

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/nik003kk/DeskBooker/backend/internal/models"
	"strconv"
	"time"
)

type Lister interface {
	ListAdminUsers(context.Context, int, int) (models.AdminUserPage, error)
}

func failure(c *gin.Context, status int, code, message string) {
	c.JSON(status, gin.H{"error": gin.H{"code": code, "message": message}})
}
func ListHandler(store Lister) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Cache-Control", "no-store")
		if c.GetString("auth_user_id") == "" {
			failure(c, 401, "unauthorized", "Authentication required")
			return
		}
		limit, offset := 25, 0
		for key, values := range c.Request.URL.Query() {
			if (key != "limit" && key != "offset") || len(values) != 1 {
				failure(c, 400, "invalid_pagination", "Invalid user pagination")
				return
			}
			n, err := strconv.Atoi(values[0])
			if err != nil {
				failure(c, 400, "invalid_pagination", "Invalid user pagination")
				return
			}
			if key == "limit" {
				limit = n
			} else {
				offset = n
			}
		}
		if limit < 1 || limit > 100 || offset < 0 || offset > 1000000 {
			failure(c, 400, "invalid_pagination", "Invalid user pagination")
			return
		}
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()
		page, err := store.ListAdminUsers(ctx, limit, offset)
		if err != nil {
			failure(c, 500, "users_unavailable", "Could not load users")
			return
		}
		c.JSON(200, page)
	}
}
