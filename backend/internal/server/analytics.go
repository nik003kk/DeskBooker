package server

import "github.com/gin-gonic/gin"

// Analytics is manager-only from its first release. Every registered endpoint
// shares both JWT validation and a live role/status lookup from PostgreSQL.
func RegisterAnalytics(router *gin.Engine, requireJWT, requireManager, summary gin.HandlerFunc, exports ...gin.HandlerFunc) {
	group := router.Group("/api/analytics", requireJWT, requireManager)
	group.GET("", summary)
	if len(exports) == 1 && exports[0] != nil {
		group.GET("/export.csv", exports[0])
	}
}
