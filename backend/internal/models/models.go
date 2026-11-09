// Package models defines the canonical PostgreSQL domain entities.
package models

import "time"

type UserRole string

const (
	RoleEmployee UserRole = "employee"
	RoleManager  UserRole = "manager"
	RoleAdmin    UserRole = "admin"
)

type UserStatus string

const (
	UserActive    UserStatus = "active"
	UserInactive  UserStatus = "inactive"
	UserSuspended UserStatus = "suspended"
)

type ResourceType string

const (
	ResourceDesk        ResourceType = "DESK"
	ResourceMeetingRoom ResourceType = "MEETING_ROOM"
	ResourceParkingSpot ResourceType = "PARKING_SPOT"
)

type ReservationStatus string

const (
	ReservationPending   ReservationStatus = "PENDING"
	ReservationConfirmed ReservationStatus = "CONFIRMED"
	ReservationCheckedIn ReservationStatus = "CHECKED_IN"
	ReservationCancelled ReservationStatus = "CANCELLED"
	ReservationCompleted ReservationStatus = "COMPLETED"
)

// User includes the stored password hash. Never serialize it directly to API clients.
type User struct {
	ID           string
	Email        string
	PasswordHash string
	FirstName    string
	LastName     string
	Role         UserRole
	Status       UserStatus
	Department   *string
	CreatedAt    time.Time
}

type Resource struct {
	ID        string
	Name      string
	Type      ResourceType
	Floor     *int
	Capacity  int
	IsActive  bool
	CreatedAt time.Time
}

// ResourceFeature belongs to a resource; names are unique per resource,
// ignoring case, as defined by migration 000005.
type ResourceFeature struct {
	ResourceID string
	Feature    string
}

type Reservation struct {
	ID          string
	UserID      string
	ResourceID  string
	StartsAt    time.Time
	EndsAt      time.Time
	Status      ReservationStatus
	CheckInCode *string
	ConfirmedAt *time.Time
	CheckedInAt *time.Time
	CreatedAt   time.Time
}
