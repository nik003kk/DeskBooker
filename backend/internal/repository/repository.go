// Package repository implements the PostgreSQL persistence boundary.
package repository

import (
    "context"
    "database/sql"
    "errors"
    "fmt"
    "time"

    "github.com/jackc/pgx/v5"
    "github.com/nik003kk/DeskBooker/backend/internal/models"
)

// DB is satisfied by *pgxpool.Pool and pgx.Tx; it can be replaced by a fake
// in unit tests. The caller owns and closes the underlying connection pool.
type DB interface {
    QueryRow(context.Context, string, ...any) pgx.Row
    Query(context.Context, string, ...any) (pgx.Rows, error)
}

type Store struct { db DB }

func New(db DB) *Store { return &Store{db: db} }

var ErrNotFound = pgx.ErrNoRows

const userColumns = "id, email, password_hash, first_name, last_name, role, department, created_at"
const resourceColumns = "id, name, type, floor, capacity, is_active, created_at"
const reservationColumns = "id, user_id, resource_id, starts_at, ends_at, status, check_in_code, confirmed_at, created_at"

func scanUser(row pgx.Row) (models.User, error) {
    var u models.User
    var department sql.NullString
    err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.FirstName,
        &u.LastName, &u.Role, &department, &u.CreatedAt)
    if department.Valid { u.Department = &department.String }
    return u, err
}

func scanResource(row pgx.Row) (models.Resource, error) {
    var r models.Resource
    var floor sql.NullInt32
    err := row.Scan(&r.ID, &r.Name, &r.Type, &floor, &r.Capacity,
        &r.IsActive, &r.CreatedAt)
    if floor.Valid {
        n := int(floor.Int32)
        r.Floor = &n
    }
    return r, err
}

func scanReservation(row pgx.Row) (models.Reservation, error) {
    var r models.Reservation
    var code sql.NullString
    var confirmed sql.NullTime
    err := row.Scan(&r.ID, &r.UserID, &r.ResourceID, &r.StartsAt,
        &r.EndsAt, &r.Status, &code, &confirmed, &r.CreatedAt)
    if code.Valid { r.CheckInCode = &code.String }
    if confirmed.Valid { r.ConfirmedAt = &confirmed.Time }
    return r, err
}

func (s *Store) CreateUser(ctx context.Context, in models.CreateUserInput) (models.User, error) {
    if in.Email == "" || in.PasswordHash == "" || in.FirstName == "" || in.LastName == "" {
        return models.User{}, errors.New("email, password hash, first name and last name are required")
    }
    u, err := scanUser(s.db.QueryRow(ctx,
        "INSERT INTO public.users (email, password_hash, first_name, last_name, department) VALUES ($1,$2,$3,$4,$5) RETURNING "+userColumns,
        in.Email, in.PasswordHash, in.FirstName, in.LastName, in.Department))
    if err != nil { return models.User{}, fmt.Errorf("create user: %w", err) }
    return u, nil
}

func (s *Store) GetUser(ctx context.Context, id string) (models.User, error) {
    u, err := scanUser(s.db.QueryRow(ctx,
        "SELECT "+userColumns+" FROM public.users WHERE id=$1", id))
    if err != nil { return models.User{}, fmt.Errorf("get user: %w", err) }
    return u, nil
}

func (s *Store) CreateResource(ctx context.Context, in models.CreateResourceInput) (models.Resource, error) {
    if in.Name == "" || in.Capacity <= 0 {
        return models.Resource{}, errors.New("resource name and positive capacity are required")
    }
    r, err := scanResource(s.db.QueryRow(ctx,
        "INSERT INTO public.resources (name, type, floor, capacity, is_active) VALUES ($1,$2,$3,$4,$5) RETURNING "+resourceColumns,
        in.Name, in.Type, in.Floor, in.Capacity, in.IsActive))
    if err != nil { return models.Resource{}, fmt.Errorf("create resource: %w", err) }
    return r, nil
}

func (s *Store) GetResource(ctx context.Context, id string) (models.Resource, error) {
    r, err := scanResource(s.db.QueryRow(ctx,
        "SELECT "+resourceColumns+" FROM public.resources WHERE id=$1", id))
    if err != nil { return models.Resource{}, fmt.Errorf("get resource: %w", err) }
    return r, nil
}

func (s *Store) ListActiveResources(ctx context.Context, kind *models.ResourceType) ([]models.Resource, error) {
    const q = "SELECT " + resourceColumns + " FROM public.resources WHERE is_active=TRUE AND ($1::text IS NULL OR type=$1) ORDER BY name, id"
    var filter any
    if kind != nil { filter = string(*kind) }
    rows, err := s.db.Query(ctx, q, filter)
    if err != nil { return nil, fmt.Errorf("list resources: %w", err) }
    defer rows.Close()

    result := make([]models.Resource, 0)
    for rows.Next() {
        r, scanErr := scanResource(rows)
        if scanErr != nil { return nil, fmt.Errorf("scan resource: %w", scanErr) }
        result = append(result, r)
    }
    if err = rows.Err(); err != nil { return nil, fmt.Errorf("iterate resources: %w", err) }
    return result, nil
}

func (s *Store) AddFeature(ctx context.Context, feature models.ResourceFeature) error {
    if feature.ResourceID == "" || feature.Feature == "" {
        return errors.New("resource ID and feature are required")
    }
    var stored string
    err := s.db.QueryRow(ctx,
        "INSERT INTO public.resource_features (resource_id, feature) VALUES ($1,$2) RETURNING feature",
        feature.ResourceID, feature.Feature).Scan(&stored)
    if err != nil { return fmt.Errorf("add feature: %w", err) }
    return nil
}

func (s *Store) ListFeatures(ctx context.Context, resourceID string) ([]models.ResourceFeature, error) {
    rows, err := s.db.Query(ctx,
        "SELECT resource_id, feature FROM public.resource_features WHERE resource_id=$1 ORDER BY feature",
        resourceID)
    if err != nil { return nil, fmt.Errorf("list features: %w", err) }
    defer rows.Close()
    result := make([]models.ResourceFeature, 0)
    for rows.Next() {
        var feature models.ResourceFeature
        if err := rows.Scan(&feature.ResourceID, &feature.Feature); err != nil {
            return nil, fmt.Errorf("scan feature: %w", err)
        }
        result = append(result, feature)
    }
    if err := rows.Err(); err != nil { return nil, fmt.Errorf("iterate features: %w", err) }
    return result, nil
}

func (s *Store) CreateReservation(ctx context.Context, in models.CreateReservationInput) (models.Reservation, error) {
    if in.UserID == "" || in.ResourceID == "" || in.StartsAt.IsZero() ||
        in.EndsAt.IsZero() || !in.EndsAt.After(in.StartsAt) {
        return models.Reservation{}, errors.New("user, resource and valid time interval are required")
    }
    r, err := scanReservation(s.db.QueryRow(ctx,
        "INSERT INTO public.reservations (user_id, resource_id, starts_at, ends_at) VALUES ($1,$2,$3,$4) RETURNING "+reservationColumns,
        in.UserID, in.ResourceID, in.StartsAt, in.EndsAt))
    if err != nil { return models.Reservation{}, fmt.Errorf("create reservation: %w", err) }
    return r, nil
}

func (s *Store) GetReservation(ctx context.Context, id string) (models.Reservation, error) {
    r, err := scanReservation(s.db.QueryRow(ctx,
        "SELECT "+reservationColumns+" FROM public.reservations WHERE id=$1", id))
    if err != nil { return models.Reservation{}, fmt.Errorf("get reservation: %w", err) }
    return r, nil
}

func (s *Store) ListUserReservations(ctx context.Context, userID string) ([]models.Reservation, error) {
    rows, err := s.db.Query(ctx,
        "SELECT "+reservationColumns+" FROM public.reservations WHERE user_id=$1 ORDER BY starts_at DESC, id",
        userID)
    if err != nil { return nil, fmt.Errorf("list reservations: %w", err) }
    defer rows.Close()
    result := make([]models.Reservation, 0)
    for rows.Next() {
        r, scanErr := scanReservation(rows)
        if scanErr != nil { return nil, fmt.Errorf("scan reservation: %w", scanErr) }
        result = append(result, r)
    }
    if err := rows.Err(); err != nil { return nil, fmt.Errorf("iterate reservations: %w", err) }
    return result, nil
}

// ConfirmReservation updates the status and confirmation timestamp together.
// The database's exclusion constraint rejects conflicting active bookings.
func (s *Store) ConfirmReservation(ctx context.Context, id string, at time.Time) (models.Reservation, error) {
    if id == "" || at.IsZero() { return models.Reservation{}, errors.New("reservation ID and confirmation time are required") }
    r, err := scanReservation(s.db.QueryRow(ctx,
        "UPDATE public.reservations SET status='CONFIRMED', confirmed_at=$2 WHERE id=$1 AND status='PENDING' RETURNING "+reservationColumns,
        id, at))
    if err != nil { return models.Reservation{}, fmt.Errorf("confirm reservation: %w", err) }
    return r, nil
}
