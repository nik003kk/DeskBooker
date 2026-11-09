package reservations

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/nik003kk/DeskBooker/backend/internal/models"
	"github.com/nik003kk/DeskBooker/backend/internal/repository"
)

type CheckInStore interface {
	CheckInReservation(context.Context, string, string, string) (models.Reservation, error)
}

func CheckInHandler(store CheckInStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Cache-Control", "no-store")
		userID := c.GetString("auth_user_id")
		if userID == "" {
			fail(c, 401, "unauthorized", "Authentication required")
			return
		}
		if !validUUID(c.Param("id")) {
			fail(c, 400, "invalid_reservation", "Invalid reservation ID")
			return
		}
		media, _, err := mime.ParseMediaType(c.GetHeader("Content-Type"))
		if err != nil || media != "application/json" {
			fail(c, 415, "unsupported_media_type", "Expected application/json")
			return
		}
		var request struct {
			PIN string `json:"pin"`
		}
		dec := json.NewDecoder(http.MaxBytesReader(c.Writer, c.Request.Body, 1024))
		dec.DisallowUnknownFields()
		if err = dec.Decode(&request); err != nil {
			fail(c, 400, "invalid_json", "Invalid check-in JSON")
			return
		}
		var extra any
		if err = dec.Decode(&extra); !errors.Is(err, io.EOF) {
			fail(c, 400, "invalid_json", "Only one JSON object allowed")
			return
		}
		if !models.ValidCheckInPIN(request.PIN) {
			fail(c, 400, "invalid_pin", "Expected a four-digit PIN")
			return
		}
		saved, err := store.CheckInReservation(c.Request.Context(), c.Param("id"), userID, request.PIN)
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			fail(c, 404, "reservation_not_found", "Reservation not found")
		case errors.Is(err, repository.ErrCheckInPIN):
			fail(c, 422, "incorrect_pin", "Napačna PIN koda")
		case errors.Is(err, repository.ErrCheckInWindow):
			fail(c, 409, "check_in_window", "Check-in is outside the allowed time window")
		case errors.Is(err, repository.ErrCheckInStatus):
			fail(c, 409, "check_in_status", "Reservation cannot be checked in")
		case err != nil:
			fail(c, 500, "check_in_failed", "Could not check in")
		default:
			c.JSON(http.StatusOK, gin.H{"reservation": saved.Public()})
		}
	}
}

// This single-reservation endpoint makes a saved check-in URL reloadable.
// The full reservation list and cancellation remain separate work items.
type DetailStore interface {
	GetReservation(context.Context, string) (models.Reservation, error)
	GetResource(context.Context, string) (models.Resource, error)
}

func DetailHandler(store DetailStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Cache-Control", "no-store")
		userID := c.GetString("auth_user_id")
		if userID == "" {
			fail(c, 401, "unauthorized", "Authentication required")
			return
		}
		if !validUUID(c.Param("id")) {
			fail(c, 400, "invalid_reservation", "Invalid reservation ID")
			return
		}
		saved, err := store.GetReservation(c.Request.Context(), c.Param("id"))
		if errors.Is(err, pgx.ErrNoRows) || (err == nil && saved.UserID != userID) {
			fail(c, 404, "reservation_not_found", "Reservation not found")
			return
		}
		if err != nil {
			fail(c, 500, "reservation_failed", "Could not load reservation")
			return
		}
		resource, err := store.GetResource(c.Request.Context(), saved.ResourceID)
		if err != nil {
			fail(c, 500, "reservation_failed", "Could not load reservation")
			return
		}
		c.JSON(http.StatusOK, gin.H{"reservation": saved.Public(), "resource": resource.Public()})
	}
}
