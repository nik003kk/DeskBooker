package models

// ValidCheckInPIN preserves leading zeros and accepts only four ASCII digits.
func ValidCheckInPIN(pin string) bool {
	if len(pin) != 4 {
		return false
	}
	for _, digit := range pin {
		if digit < '0' || digit > '9' {
			return false
		}
	}
	return true
}
