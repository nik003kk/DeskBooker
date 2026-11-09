package repository

import (
	"crypto/subtle"
	"github.com/nik003kk/DeskBooker/backend/internal/models"
)

func matchesCheckInPIN(expected *string, actual string) bool {
	if expected == nil || !models.ValidCheckInPIN(*expected) || !models.ValidCheckInPIN(actual) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(*expected), []byte(actual)) == 1
}
