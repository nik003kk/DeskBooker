package models

import "testing"

func TestCheckInPINFormat(t *testing.T) {
	for _, pin := range []string{"0000", "0042", "9999"} {
		if !ValidCheckInPIN(pin) {
			t.Errorf("valid PIN rejected: %q", pin)
		}
	}
	for _, pin := range []string{"", "123", "12345", " 123", "123 ", "12a4", "１２３４", "١٢٣٤", "12\n4"} {
		if ValidCheckInPIN(pin) {
			t.Errorf("invalid PIN accepted: %q", pin)
		}
	}
}
