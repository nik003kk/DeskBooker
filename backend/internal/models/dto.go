package models

import "time"

// Input DTOs are intentionally separate from persisted entities. They do not
// allow callers to set database-generated IDs, created_at, or user roles.
type CreateUserInput struct {
    Email string `json:"email"`
    PasswordHash string `json:"-"`
    FirstName string `json:"first_name"`
    LastName string `json:"last_name"`
    Department *string `json:"department,omitempty"`
}

type CreateResourceInput struct {
    Name string `json:"name"`
    Type ResourceType `json:"type"`
    Floor *int `json:"floor,omitempty"`
    Capacity int `json:"capacity"`
    IsActive bool `json:"is_active"`
}

type CreateReservationInput struct {
    UserID string `json:"user_id"`
    ResourceID string `json:"resource_id"`
    StartsAt time.Time `json:"starts_at"`
    EndsAt time.Time `json:"ends_at"`
}

type UserDTO struct {
    ID string `json:"id"`
    Email string `json:"email"`
    FirstName string `json:"first_name"`
    LastName string `json:"last_name"`
    Role UserRole `json:"role"`
    Status UserStatus `json:"status"`
    Department *string `json:"department,omitempty"`
    CreatedAt time.Time `json:"created_at"`
}

func (u User) Public() UserDTO {
    return UserDTO{
        ID: u.ID, Email: u.Email, FirstName: u.FirstName,
        LastName: u.LastName, Role: u.Role, Status: u.Status, Department: u.Department,
        CreatedAt: u.CreatedAt,
    }
}

type ResourceDTO struct {
    ID string `json:"id"`
    Name string `json:"name"`
    Type ResourceType `json:"type"`
    Floor *int `json:"floor,omitempty"`
    Capacity int `json:"capacity"`
    IsActive bool `json:"is_active"`
    CreatedAt time.Time `json:"created_at"`
}

func (r Resource) Public() ResourceDTO {
    return ResourceDTO{
        ID: r.ID, Name: r.Name, Type: r.Type, Floor: r.Floor,
        Capacity: r.Capacity, IsActive: r.IsActive, CreatedAt: r.CreatedAt,
    }
}

type ReservationDTO struct {
    ID string `json:"id"`
    UserID string `json:"user_id"`
    ResourceID string `json:"resource_id"`
    StartsAt time.Time `json:"starts_at"`
    EndsAt time.Time `json:"ends_at"`
    Status ReservationStatus `json:"status"`
    ConfirmedAt *time.Time `json:"confirmed_at,omitempty"`
    CreatedAt time.Time `json:"created_at"`
}

func (r Reservation) Public() ReservationDTO {
    // Deliberately does not expose check-in codes in ordinary reservation views.
    return ReservationDTO{
        ID: r.ID, UserID: r.UserID, ResourceID: r.ResourceID,
        StartsAt: r.StartsAt, EndsAt: r.EndsAt, Status: r.Status,
        ConfirmedAt: r.ConfirmedAt, CreatedAt: r.CreatedAt,
    }
}
