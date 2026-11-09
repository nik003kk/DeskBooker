package auth

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/nik003kk/DeskBooker/backend/internal/models"
	"net/http"
	"strings"
)

const currentUserKey = "auth_current_user"

// RequireJWT is signature-only for isolated JWT/route tests. Production routes
// use RequireCurrentJWT so account existence, status and role are fresh.
func RequireJWT(tokens *JWTService) gin.HandlerFunc { return requireJWT(tokens, nil) }
func RequireCurrentJWT(tokens *JWTService, store UserByIDFinder) gin.HandlerFunc {
	if store == nil {
		return func(c *gin.Context) {
			responseError(c, 503, "authentication_unavailable", "Authentication unavailable")
			c.Abort()
		}
	}
	return requireJWT(tokens, store)
}
func requireJWT(tokens *JWTService, store UserByIDFinder) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Cache-Control", "no-store")
		fields := strings.Fields(c.GetHeader("Authorization"))
		if len(fields) != 2 || fields[0] != "Bearer" {
			responseError(c, http.StatusUnauthorized, "unauthorized", "Valid bearer token required")
			c.Abort()
			return
		}
		claims, err := tokens.Verify(fields[1])
		if err != nil {
			responseError(c, 401, "unauthorized", "Valid bearer token required")
			c.Abort()
			return
		}
		c.Set("auth_user_id", claims.Subject)
		if store != nil {
			user, err := store.GetUser(c.Request.Context(), claims.Subject)
			if errors.Is(err, pgx.ErrNoRows) {
				responseError(c, 401, "unauthorized", "Authentication required")
				c.Abort()
				return
			}
			if err != nil {
				responseError(c, 500, "authorization_failed", "Could not verify account")
				c.Abort()
				return
			}
			if !strings.EqualFold(user.ID, claims.Subject) {
				responseError(c, 401, "unauthorized", "Authentication required")
				c.Abort()
				return
			}
			if user.Status != models.UserActive || (user.Role != models.RoleEmployee && user.Role != models.RoleManager && user.Role != models.RoleAdmin) {
				responseError(c, 403, "forbidden", "Active account required")
				c.Abort()
				return
			}
			c.Set("auth_user_id", user.ID)
			c.Set(currentUserKey, user.Public())
		}
		c.Next()
	}
}
