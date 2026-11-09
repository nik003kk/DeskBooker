package server

import "github.com/gin-gonic/gin"

func RegisterAdminUsers(router *gin.Engine, requireJWT, requireAdmin, list gin.HandlerFunc) {
	group := router.Group("/api/admin/users", requireJWT, requireAdmin)
	group.GET("", list)
}
