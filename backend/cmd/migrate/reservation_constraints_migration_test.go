package main

import (
    "os"
    "path/filepath"
    "strings"
    "testing"
)

// Lightweight SQL contract test for GitHub Actions without PostgreSQL.
// The actual exclusion/index behavior is verified by reservation-constraints-smoke.sql.
func TestReservationConstraintsMigrationContract(t *testing.T) {
    directory, err := filepath.Abs(filepath.Join("..", "..", "migrations"))
    if err != nil {
        t.Fatal(err)
    }
    up, err := os.ReadFile(filepath.Join(directory, "000007_reservation_constraints.up.sql"))
    if err != nil {
        t.Fatal(err)
    }
    down, err := os.ReadFile(filepath.Join(directory, "000007_reservation_constraints.down.sql"))
    if err != nil {
        t.Fatal(err)
    }

    sql := strings.ToLower(string(up))
    for _, required := range []string{
        "create extension if not exists btree_gist",
        "create index reservations_user_starts_at_idx",
        "(user_id, starts_at desc)",
        "create index reservations_resource_starts_at_idx",
        "(resource_id, starts_at)",
        "create unique index reservations_check_in_code_lower_unique_idx",
        "lower(check_in_code)",
        "where check_in_code is not null",
        "add constraint reservations_no_active_overlap",
        "exclude using gist",
        "resource_id with =",
        "tstzrange(starts_at, ends_at, '[)') with &&",
        "where (status in ('pending', 'confirmed'))",
    } {
        if !strings.Contains(sql, required) {
            t.Errorf("reservation constraints migration missing %q", required)
        }
    }

    rollback := strings.ToLower(string(down))
    for _, required := range []string{
        "drop constraint reservations_no_active_overlap",
        "drop index public.reservations_check_in_code_lower_unique_idx",
        "drop index public.reservations_resource_starts_at_idx",
        "drop index public.reservations_user_starts_at_idx",
    } {
        if !strings.Contains(rollback, required) {
            t.Errorf("rollback missing %q", required)
        }
    }
    if strings.Contains(rollback, "drop table") || strings.Contains(rollback, "drop extension") {
        t.Error("rollback must preserve reservation data and shared extensions")
    }
}
