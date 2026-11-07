package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// NewRouter sets up HTTP routes and middleware for DeskBooker.
// Database access, authentication and domain routes are added in later issues.
func NewRouter() *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	api := router.Group("/api")
	api.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "deskbooker-api",
		})
	})

	return router
}
