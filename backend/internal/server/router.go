package server

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// DatabasePinger allows the router to check database availability independently
// from the concrete PostgreSQL driver.
type DatabasePinger interface {
	Ping(context.Context) error
}

// NewRouter sets up HTTP routes and middleware for DeskBooker.
func NewRouter(db DatabasePinger) *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	api := router.Group("/api")
	api.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "deskbooker-api",
		})
	})

	api.GET("/health/db", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
		defer cancel()

		if err := db.Ping(ctx); err != nil {
			// Do not return internal DB connection errors or credentials to clients.
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unavailable"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok", "database": "connected"})
	})

	return router
}
