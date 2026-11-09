package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/nik003kk/DeskBooker/backend/internal/models"
)

// Product OFFICE_MANAGER corresponds to the persisted manager role.
// RequireRoles reloads both role and account status for every API request.
func RequireAnalyticsManager(store RoleLookup) gin.HandlerFunc {
	return RequireRoles(store, models.RoleManager, models.RoleAdmin)
}
