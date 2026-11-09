package main

import (
    "os"
    "path/filepath"
    "strings"
    "testing"
)

// Verifies migration structure in CI without a database. PostgreSQL behavior
// is covered by the rollback-only backend/scripts/reservations-smoke.sql.
func TestReservationsMigrationContract(t *testing.T) {
    dir, err := filepath.Abs(filepath.Join("..", "..", "migrations"))
    if err != nil {
        t.Fatal(err)
    }
    up, err := os.ReadFile(filepath.Join(dir, "000006_create_reservations.up.sql"))
    if err != nil {
        t.Fatal(err)
    }
    down, err := os.ReadFile(filepath.Join(dir, "000006_create_reservations.down.sql"))
    if err != nil {
        t.Fatal(err)
    }
    sql := strings.ToLower(string(up))
    for _, required := range []string{
        "create table public.reservations",
        "id uuid primary key default gen_random_uuid()",
        "user_id uuid not null references public.users (id)",
        "resource_id uuid not null references public.resources (id)",
        "starts_at timestamptz not null",
        "ends_at timestamptz not null",
        "status text not null default 'pending'",
        "check_in_code text",
        "confirmed_at timestamptz",
        "created_at timestamptz not null default now()",
        "ends_at > starts_at",
        "'confirmed'",
        "'cancelled'",
        "'completed'",
        "reservations_check_in_code_nonempty_check",
    } {
        if !strings.Contains(sql, required) {
            t.Errorf("reservations migration missing %q", required)
        }
    }
    downSQL := strings.ToLower(string(down))
    if !strings.Contains(downSQL, "drop table public.reservations;") {
        t.Error("down migration must drop reservations")
    }
    if strings.Contains(downSQL, "drop table public.users") || strings.Contains(downSQL, "drop table public.resources") {
        t.Error("down migration must preserve parent entities")
    }
}
