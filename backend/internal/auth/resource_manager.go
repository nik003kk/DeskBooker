package auth

import (
    "github.com/gin-gonic/gin"
    "github.com/nik003kk/DeskBooker/backend/internal/models"
)

// RequireResourceManager maps product OFFICE_MANAGER / ADMIN permissions
// to persisted roles "manager" / "admin". Only active accounts may act.
func RequireResourceManager(store RoleLookup) gin.HandlerFunc {
    return RequireRoles(store,models.RoleManager,models.RoleAdmin)
}
