package server

import "github.com/gin-gonic/gin"

func RegisterAdminUsers(router *gin.Engine, requireJWT, requireAdmin, list gin.HandlerFunc, changes ...gin.HandlerFunc) {
	group := router.Group("/api/admin/users", requireJWT, requireAdmin)
	group.GET("", list)
	if len(changes) > 0 && changes[0] != nil {
		group.PATCH("/:id/role", changes[0])
	}
}
