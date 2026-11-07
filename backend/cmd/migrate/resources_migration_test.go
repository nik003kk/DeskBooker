package main

import (
    "os"
    "path/filepath"
    "strings"
    "testing"
)

// Contract check without a PostgreSQL instance; resources-smoke.sql tests live constraints.
func TestResourcesMigrationContract(t *testing.T) {
    directory, err := filepath.Abs(filepath.Join("..", "..", "migrations"))
    if err != nil {
        t.Fatal(err)
    }

    up, err := os.ReadFile(filepath.Join(directory, "000004_create_resources.up.sql"))
    if err != nil {
        t.Fatal(err)
    }
    down, err := os.ReadFile(filepath.Join(directory, "000004_create_resources.down.sql"))
    if err != nil {
        t.Fatal(err)
    }

    sql := strings.ToLower(string(up))
    for _, required := range []string{
        "create table public.resources",
        "id uuid primary key default gen_random_uuid()",
        "name text not null",
        "type text not null",
        "floor integer",
        "capacity integer not null default 1",
        "is_active boolean not null default true",
        "created_at timestamptz not null default now()",
        "'desk'",
        "'meeting_room'",
        "'parking_spot'",
        "capacity > 0",
        "create unique index resources_name_lower_unique_idx",
        "lower(name)",
    } {
        if !strings.Contains(sql, required) {
            t.Errorf("resources migration missing %q", required)
        }
    }

    downSQL := strings.ToLower(string(down))
    if !strings.Contains(downSQL, "drop table public.resources;") {
        t.Error("down migration must remove public.resources")
    }
    if strings.Contains(downSQL, "drop table public.resources cascade") {
        t.Error("rollback must not cascade into dependent objects")
    }
}
