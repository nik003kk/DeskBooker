// Command repository-smoke verifies the repository against a migrated local DB.
// The transaction is always rolled back; no test data is persisted.
package main

import (
    "context"
    "errors"
    "fmt"
    "log"
    "os"
    "time"

    "github.com/jackc/pgx/v5"
    "github.com/nik003kk/DeskBooker/backend/internal/config"
    "github.com/nik003kk/DeskBooker/backend/internal/database"
    "github.com/nik003kk/DeskBooker/backend/internal/models"
    "github.com/nik003kk/DeskBooker/backend/internal/repository"
)

func main() {
    if err := config.LoadLocalEnv(); err != nil {
        log.Fatal(err)
    }
    ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
    defer cancel()

    pool, err := database.Open(ctx)
    if err != nil { log.Fatal(err) }
    defer pool.Close()

    tx, err := pool.Begin(ctx)
    if err != nil { log.Fatal(err) }
    defer func() {
        // Rollback even if one of the checks fails.
        _ = tx.Rollback(context.Background())
    }()

    if err := check(ctx, repository.New(tx)); err != nil {
        log.Fatal(err)
    }
    fmt.Fprintln(os.Stdout, "DB-020 repository smoke tests passed: user, resource, features, reservation, confirmation, rollback")
}

func check(ctx context.Context, store *repository.Store) error {
    user, err := store.CreateUser(ctx, models.CreateUserInput{
        Email: "db020-smoke@example.test", PasswordHash: "placeholder-hash",
        FirstName: "Smoke", LastName: "Test",
    })
    if err != nil { return fmt.Errorf("create user: %w", err) }
    savedUser, err := store.GetUser(ctx, user.ID)
    if err != nil || savedUser.Email != user.Email {
        return fmt.Errorf("get user mismatch: %w", err)
    }

    floor := 1
    resource, err := store.CreateResource(ctx, models.CreateResourceInput{
        Name: "DB020 Smoke Desk", Type: models.ResourceDesk,
        Floor: &floor, Capacity: 1, IsActive: true,
    })
    if err != nil { return fmt.Errorf("create resource: %w", err) }
    savedResource, err := store.GetResource(ctx, resource.ID)
    if err != nil || savedResource.Name != resource.Name {
        return fmt.Errorf("get resource mismatch: %w", err)
    }
    if err := store.AddFeature(ctx, models.ResourceFeature{
        ResourceID: resource.ID, Feature: "monitor",
    }); err != nil { return fmt.Errorf("add feature: %w", err) }
    features, err := store.ListFeatures(ctx, resource.ID)
    if err != nil || len(features) != 1 || features[0].Feature != "monitor" {
        return fmt.Errorf("list feature mismatch: %w", err)
    }

    resourceType := models.ResourceDesk
    resources, err := store.ListActiveResources(ctx, &resourceType)
    if err != nil { return fmt.Errorf("list resources: %w", err) }
    found := false
    for _, r := range resources { if r.ID == resource.ID { found = true } }
    if !found { return errors.New("active resource was not listed") }

    start := time.Date(2032, 6, 1, 9, 0, 0, 0, time.UTC)
    booking, err := store.CreateReservation(ctx, models.CreateReservationInput{
        UserID: user.ID, ResourceID: resource.ID, StartsAt: start,
        EndsAt: start.Add(time.Hour),
    })
    if err != nil { return fmt.Errorf("create reservation: %w", err) }
    if booking.Status != models.ReservationPending {
        return fmt.Errorf("unexpected initial status: %s", booking.Status)
    }

    savedBooking, err := store.GetReservation(ctx, booking.ID)
    if err != nil || savedBooking.ID != booking.ID {
        return fmt.Errorf("get reservation mismatch: %w", err)
    }
    bookings, err := store.ListUserReservations(ctx, user.ID)
    if err != nil || len(bookings) != 1 || bookings[0].ID != booking.ID {
        return fmt.Errorf("list reservations mismatch: %w", err)
    }
    confirmed, err := store.ConfirmReservation(ctx, booking.ID, time.Now().UTC())
    if err != nil || confirmed.Status != models.ReservationConfirmed ||
        confirmed.ConfirmedAt == nil {
        return fmt.Errorf("confirm reservation mismatch: %w", err)
    }

    // A second confirmation is not valid. Verify ErrNoRows is preserved.
    _, err = store.ConfirmReservation(ctx, booking.ID, time.Now().UTC())
    if !errors.Is(err, pgx.ErrNoRows) {
        return fmt.Errorf("expected ErrNoRows for second confirmation, got %v", err)
    }
    return nil
}
