package repository

import (
    "context"
    "errors"
    "strings"
    "testing"

    "github.com/jackc/pgx/v5"
)

func TestGetUserByEmailUsesBoundNormalizedQuery(t *testing.T) {
    db := &stubDB{err: pgx.ErrNoRows}
    _, err := New(db).GetUserByEmail(context.Background(), "Ana@Gmail.COM")
    if !errors.Is(err, ErrNotFound) {
        t.Fatalf("expected ErrNotFound for unknown user, got %v", err)
    }
    if !strings.Contains(db.sql, "WHERE LOWER(email)=LOWER($1)") ||
        strings.Contains(db.sql, "Ana@Gmail.COM") {
        t.Fatalf("user input embedded into SQL: %q", db.sql)
    }
    if len(db.args) != 1 || db.args[0] != "Ana@Gmail.COM" {
        t.Fatalf("unexpected bound arguments %#v", db.args)
    }
}
