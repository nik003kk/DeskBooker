package repository

import (
    "context"
    "errors"
    "strings"
    "testing"
    "time"

    "github.com/jackc/pgx/v5"
    "github.com/nik003kk/DeskBooker/backend/internal/models"
)

type failingRow struct { err error }
func (r failingRow) Scan(_ ...any) error { return r.err }
type stubDB struct {
    sql string
    args []any
    err error
}
func (db *stubDB) QueryRow(_ context.Context, sql string, args ...any) pgx.Row {
    db.sql, db.args = sql, args
    if db.err == nil { db.err = pgx.ErrNoRows }
    return failingRow{db.err}
}
func (db *stubDB) Query(_ context.Context, sql string, args ...any) (pgx.Rows, error) {
    db.sql, db.args = sql, args
    if db.err == nil { db.err = errors.New("query unavailable") }
    return nil, db.err
}

func TestGetUserPreservesNotFound(t *testing.T) {
    db := &stubDB{err: pgx.ErrNoRows}
    _, err := New(db).GetUser(context.Background(), "fake-uuid")
    if !errors.Is(err, ErrNotFound) { t.Fatalf("expected ErrNotFound, got %v", err) }
    if !strings.Contains(db.sql, "FROM public.users WHERE id=$1") {
        t.Fatalf("unexpected query: %s", db.sql)
    }
}

func TestCreateUserUsesBoundParameters(t *testing.T) {
    db := &stubDB{err: pgx.ErrNoRows}
    _, _ = New(db).CreateUser(context.Background(), models.CreateUserInput{
        Email: "person@example.test", PasswordHash: "secret-hash",
        FirstName: "First", LastName: "Last",
    })
    if len(db.args) != 5 || db.args[0] != "person@example.test" ||
        db.args[1] != "secret-hash" || strings.Contains(db.sql, "person@example.test") {
        t.Fatalf("SQL must use bound parameters, got query %q and arguments %#v", db.sql, db.args)
    }
}

func TestCreateReservationRejectsInvalidTimeWithoutQuery(t *testing.T) {
    db := &stubDB{}
    start := time.Date(2030, 1, 1, 12, 0, 0, 0, time.UTC)
    _, err := New(db).CreateReservation(context.Background(), models.CreateReservationInput{
        UserID: "u", ResourceID: "r", StartsAt: start, EndsAt: start,
    })
    if err == nil { t.Fatal("expected invalid interval error") }
    if db.sql != "" { t.Fatal("invalid reservation must not query the database") }
}

func TestCreateReservationUsesDatabaseExclusionConstraint(t *testing.T) {
    db := &stubDB{err: errors.New("conflicting exclusion constraint")}
    start := time.Date(2030, 1, 1, 12, 0, 0, 0, time.UTC)
    _, err := New(db).CreateReservation(context.Background(), models.CreateReservationInput{
        UserID: "u", ResourceID: "r", StartsAt: start, EndsAt: start.Add(time.Hour),
    })
    if err == nil || !strings.Contains(err.Error(), "conflicting exclusion constraint") {
        t.Fatalf("expected propagated DB constraint error, got %v", err)
    }
    if len(db.args) != 4 || !strings.Contains(db.sql, "RETURNING") {
        t.Fatalf("unexpected statement: %q, %#v", db.sql, db.args)
    }
}

func TestListErrorsPropagate(t *testing.T) {
    db := &stubDB{err: errors.New("offline")}
    _, err := New(db).ListActiveResources(context.Background(), nil)
    if err == nil || !strings.Contains(err.Error(), "offline") { t.Fatalf("expected DB error: %v", err) }
    if !strings.Contains(db.sql, "is_active=TRUE") { t.Fatalf("missing active filter: %s", db.sql) }
}
