package analytics

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/nik003kk/DeskBooker/backend/internal/models"
	"net/http"
	"time"
)

type Store interface {
	ReservationAnalytics(context.Context, models.AnalyticsRange) (models.Analytics, error)
}

func loadReport(c *gin.Context, store Store) (models.Analytics, bool) {
	c.Header("Cache-Control", "no-store")
	period, err := parseRange(c.Request.URL.Query(), time.Now())
	if err != nil {
		c.JSON(400, gin.H{"error": gin.H{"code": "invalid_period", "message": "Expected from/to dates in order, at most 366 days"}})
		return models.Analytics{}, false
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	report, err := store.ReservationAnalytics(ctx, period)
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "analytics_failed", "message": "Could not load analytics"}})
		return models.Analytics{}, false
	}
	return report, true
}
func SummaryHandler(store Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		if report, ok := loadReport(c, store); ok {
			c.JSON(http.StatusOK, gin.H{"analytics": report})
		}
	}
}
