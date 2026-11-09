package repository

import "testing"

func TestCheckInPINComparison(t *testing.T) {
	pin := "0042"
	if !matchesCheckInPIN(&pin, "0042") {
		t.Fatal("leading zero PIN rejected")
	}
	for _, actual := range []string{"42", "0000", "0043", " 0042", "0042 ", "００４２", ""} {
		if matchesCheckInPIN(&pin, actual) {
			t.Errorf("incorrect PIN accepted: %q", actual)
		}
	}
	if matchesCheckInPIN(nil, pin) {
		t.Fatal("missing PIN accepted")
	}
}
