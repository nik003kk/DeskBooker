package adminusers

import (
	"github.com/gin-gonic/gin"
	"strings"
)

// Account deletion is outside the MVP. Explicitly protect a self-delete
// request and reject all other deletions without modifying any account.
func ProtectDeletionHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Cache-Control", "no-store")
		actor := c.GetString("auth_user_id")
		if actor == "" {
			failure(c, 401, "unauthorized", "Authentication required")
			return
		}
		target := c.Param("id")
		if !userUUID(target) {
			failure(c, 400, "invalid_user", "Invalid user ID")
			return
		}
		if strings.EqualFold(actor, target) {
			failure(c, 403, "self_account_protected", "Cannot delete your own account")
			return
		}
		failure(c, 403, "account_deletion_disabled", "Account deletion is disabled")
	}
}
