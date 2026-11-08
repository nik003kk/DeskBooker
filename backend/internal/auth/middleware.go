package auth

import (
    "net/http"
    "strings"

    "github.com/gin-gonic/gin"
)

// RequireJWT rejects absent, malformed, expired or incorrectly signed
// Authorization: Bearer tokens and stores the authenticated user ID in Gin.
func RequireJWT(tokens *JWTService) gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Header("Cache-Control", "no-store")
        authorization := c.GetHeader("Authorization")
        fields := strings.Fields(authorization)
        if len(fields) != 2 || fields[0] != "Bearer" {
            responseError(c, http.StatusUnauthorized, "unauthorized", "Valid bearer token required")
            c.Abort()
            return
        }
        claims, err := tokens.Verify(fields[1])
        if err != nil {
            responseError(c, http.StatusUnauthorized, "unauthorized", "Valid bearer token required")
            c.Abort()
            return
        }
        c.Set("auth_user_id", claims.Subject)
        c.Next()
    }
}
