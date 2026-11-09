package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/nik003kk/DeskBooker/backend/internal/models"
)

func RequireAdministrator(store RoleLookup) gin.HandlerFunc {
	return RequireRoles(store, models.RoleAdmin)
}
