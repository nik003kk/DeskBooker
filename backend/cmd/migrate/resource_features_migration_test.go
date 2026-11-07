package main

import (
    "os"
    "path/filepath"
    "strings"
    "testing"
)

// Structural checks run in CI without PostgreSQL; integration smoke tests
// validate constraint behavior against a real local database.
func TestResourceFeaturesMigrationContract(t *testing.T) {
    directory, err := filepath.Abs(filepath.Join("..", "..", "migrations"))
    if err != nil {
        t.Fatal(err)
    }
    up, err := os.ReadFile(filepath.Join(directory, "000005_create_resource_features.up.sql"))
    if err != nil {
        t.Fatal(err)
    }
    down, err := os.ReadFile(filepath.Join(directory, "000005_create_resource_features.down.sql"))
    if err != nil {
        t.Fatal(err)
    }

    sql := strings.ToLower(string(up))
    for _, required := range []string{
        "create table public.resource_features",
        "resource_id uuid not null",
        "references public.resources (id) on delete cascade",
        "feature text not null",
        "resource_features_feature_nonempty_check",
        "create unique index resource_features_resource_feature_unique_idx",
        "(resource_id, lower(feature))",
    } {
        if !strings.Contains(sql, required) {
            t.Errorf("resource_features migration missing %q", required)
        }
    }
    downSQL := strings.ToLower(string(down))
    if !strings.Contains(downSQL, "drop table public.resource_features;") {
        t.Error("down migration must remove public.resource_features")
    }
    if strings.Contains(downSQL, "drop table public.resources") {
        t.Error("down migration must not remove parent resources")
    }
}
