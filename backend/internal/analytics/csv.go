package analytics

import (
	"bytes"
	"encoding/csv"
	"github.com/gin-gonic/gin"
	"github.com/nik003kk/DeskBooker/backend/internal/models"
	"net/http"
	"strconv"
	"time"
)

func CSVHandler(store Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		report, ok := loadReport(c, store)
		if !ok {
			return
		}
		content, err := analyticsCSV(report)
		if err != nil {
			c.JSON(500, gin.H{"error": gin.H{"code": "csv_failed", "message": "Could not export analytics"}})
			return
		}
		c.Header("Content-Disposition", `attachment; filename="deskbooker-analytics-`+report.From+`-`+report.To+`.csv"`)
		c.Header("X-Content-Type-Options", "nosniff")
		c.Data(http.StatusOK, "text/csv; charset=utf-8", content)
	}
}
func analyticsCSV(report models.Analytics) ([]byte, error) {
	rows := [][]string{{"section", "metric", "date", "weekday", "value", "unit"}}
	add := func(section, metric, date, weekday, value, unit string) {
		rows = append(rows, []string{section, metric, date, weekday, value, unit})
	}
	integer := func(n int64) string { return strconv.FormatInt(n, 10) }
	number := func(n float64) string { return strconv.FormatFloat(n, 'f', -1, 64) }
	percentage := func(n *float64) string {
		if n == nil {
			return ""
		}
		return number(*n)
	}
	for _, pair := range [][2]string{{"from", report.From}, {"to", report.To}, {"timezone", report.Timezone}, {"generated_at", report.GeneratedAt.Format(time.RFC3339Nano)}} {
		add("metadata", pair[0], "", "", pair[1], "text")
	}
	add("summary", "total_reservations", "", "", integer(report.TotalReservations), "count")
	add("summary", "active_reservations", "", "", integer(report.ActiveReservations), "count")
	add("summary", "occupancy_rate", "", "", percentage(report.OccupancyRate), "percent")
	add("summary", "no_show_rate", "", "", percentage(report.NoShow.Rate), "percent")
	add("summary", "no_show_expired", "", "", integer(report.NoShow.Expired), "count")
	add("summary", "no_show_attended", "", "", integer(report.NoShow.Attended), "count")
	add("summary", "no_show_eligible", "", "", integer(report.NoShow.Eligible), "count")
	for _, status := range []string{"PENDING", "CONFIRMED", "CHECKED_IN", "COMPLETED", "CANCELLED", "EXPIRED"} {
		add("status", status, "", "", integer(report.StatusCounts[status]), "count")
	}
	for _, series := range []struct {
		name   string
		points []models.OccupancyPoint
	}{{"daily", report.Daily}, {"weekday", report.Weekdays}} {
		for _, point := range series.points {
			day := strconv.Itoa(point.Weekday)
			add(series.name, "booked_minutes", point.Date, day, number(point.BookedMinutes), "minutes")
			add(series.name, "capacity_minutes", point.Date, day, number(point.CapacityMinutes), "minutes")
			add(series.name, "occupancy_rate", point.Date, day, percentage(point.OccupancyRate), "percent")
		}
	}
	var buffer bytes.Buffer
	buffer.WriteString("\uFEFF") // UTF-8 BOM for spreadsheet applications.
	writer := csv.NewWriter(&buffer)
	writer.UseCRLF = true
	if err := writer.WriteAll(rows); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}
