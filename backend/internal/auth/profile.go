package auth

import (
    "context"
    "errors"
    "net/http"

    "github.com/gin-gonic/gin"
    "github.com/jackc/pgx/v5"
    "github.com/nik003kk/DeskBooker/backend/internal/models"
)

type ProfileReader interface {
    GetUser(context.Context, string) (models.User, error)
}

// ProfileHandler is served only behind RequireJWT.
func ProfileHandler(store ProfileReader) gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Header("Cache-Control", "no-store")
        id := c.GetString("auth_user_id")
        if id == "" {
            responseError(c, http.StatusUnauthorized, "unauthorized", "Authentication required")
            return
        }
        user, err := store.GetUser(c.Request.Context(), id)
        if errors.Is(err, pgx.ErrNoRows) {
            responseError(c, http.StatusUnauthorized, "unauthorized", "Authentication required")
            return
        }
        if err != nil {
            responseError(c, http.StatusInternalServerError, "profile_unavailable", "Could not load profile")
            return
        }
        c.JSON(http.StatusOK, gin.H{"user": user.Public()})
    }
}
