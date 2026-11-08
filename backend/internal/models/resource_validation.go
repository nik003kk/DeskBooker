package models

import (
    "errors"
    "strings"
    "unicode"
)

var ErrInvalidResource = errors.New("invalid resource name, floor, type or capacity")

// ValidateResourceInput is shared by HTTP handlers and repository callers.
// The database also enforces core invariants.
func ValidateResourceInput(in CreateResourceInput) error {
    if in.Name == "" || strings.TrimSpace(in.Name) != in.Name ||
        len([]rune(in.Name)) > 150 || in.Capacity <= 0 ||
        (in.Floor != nil && *in.Floor < 0) {
        return ErrInvalidResource
    }
    for _, ch := range in.Name {
        if unicode.IsControl(ch) { return ErrInvalidResource }
    }
    switch in.Type {
    case ResourceDesk, ResourceMeetingRoom, ResourceParkingSpot:
        return nil
    default:
        return ErrInvalidResource
    }
}
