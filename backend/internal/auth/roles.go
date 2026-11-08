package auth

import (
    "context"
    "errors"
    "net/http"

    "github.com/gin-gonic/gin"
    "github.com/jackc/pgx/v5"
    "github.com/nik003kk/DeskBooker/backend/internal/models"
)

// RoleLookup checks current DB role instead of trusting user-controlled data
// or stale JWT claims. RequireJWT must run earlier in the handler chain.
type RoleLookup interface {
    GetUser(context.Context, string) (models.User, error)
}

func RequireRoles(store RoleLookup, roles ...models.UserRole) gin.HandlerFunc {
    allowed := make(map[models.UserRole]bool, len(roles))
    for _, role := range roles { allowed[role] = true }
    return func(c *gin.Context) {
        id := c.GetString("auth_user_id")
        if id == "" {
            responseError(c, http.StatusUnauthorized, "unauthorized", "Authentication required")
            c.Abort()
            return
        }
        user, err := store.GetUser(c.Request.Context(), id)
        if errors.Is(err, pgx.ErrNoRows) {
            responseError(c, http.StatusUnauthorized, "unauthorized", "Authentication required")
            c.Abort()
            return
        }
        if err != nil {
            responseError(c, http.StatusInternalServerError, "authorization_failed", "Could not authorize request")
            c.Abort()
            return
        }
        if !allowed[user.Role] {
            responseError(c, http.StatusForbidden, "forbidden", "Insufficient permissions")
            c.Abort()
            return
        }
        c.Next()
    }
}
