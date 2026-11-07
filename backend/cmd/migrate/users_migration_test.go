package main

import (
    "os"
    "path/filepath"
    "strings"
    "testing"
)

// This contract test catches omissions in the migration even when CI has no DB.
// Run backend/scripts/users-smoke.sql for actual PostgreSQL constraint tests.
func TestUsersMigrationContract(t *testing.T) {
    dir, err := filepath.Abs(filepath.Join("..", "..", "migrations"))
    if err != nil { t.Fatal(err) }
    up, err := os.ReadFile(filepath.Join(dir, "000003_create_users.up.sql"))
    if err != nil { t.Fatal(err) }
    down, err := os.ReadFile(filepath.Join(dir, "000003_create_users.down.sql"))
    if err != nil { t.Fatal(err) }

    sql := strings.ToLower(string(up))
    for _, want := range []string{
        "create table public.users",
        "id uuid primary key default gen_random_uuid()",
        "email text not null",
        "password_hash text not null",
        "first_name text not null",
        "last_name text not null",
        "role text not null",
        "department text",
        "created_at timestamptz not null default now()",
        "create unique index users_email_lower_unique_idx",
        "lower(email)",
    } {
        if !strings.Contains(sql, want) {
            t.Errorf("users migration missing %q", want)
        }
    }
    if !strings.Contains(strings.ToLower(string(down)), "drop table public.users") {
        t.Error("down migration must drop public.users")
    }
    if strings.Contains(strings.ToLower(string(down)), "cascade") && strings.Contains(strings.ToLower(string(down)), "drop table public.users cascade") {
        t.Error("down migration must not cascade onto dependent application objects")
    }
}
