package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidAction(t *testing.T) {
	for _, tc := range []struct {
		action string
		want bool
	}{
		{"up", true},
		{"down", true},
		{"status", true},
		{"", false},
		{"force", false},
		{"drop", false},
		{"up 1", false},
	} {
		if got := validAction(tc.action); got != tc.want {
			t.Errorf("validAction(%q) = %t, want %t", tc.action, got, tc.want)
		}
	}
}

func TestRunRejectsInvalidCommandBeforeConnecting(t *testing.T) {
	for _, args := range [][]string{nil, {"invalid"}, {"up", "down"}} {
		if err := run(args); err == nil {
			t.Errorf("expected usage error for %q", args)
		}
	}
}

func TestMigrationsDirectory(t *testing.T) {
	root := t.TempDir()
	backend := filepath.Join(root, "backend")
	migrations := filepath.Join(backend, "migrations")
	if err := os.MkdirAll(migrations, 0700); err != nil {
		t.Fatal(err)
	}

	original, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(original) })

	for _, cwd := range []string{root, backend} {
		if err := os.Chdir(cwd); err != nil {
			t.Fatal(err)
		}
		got, err := migrationsDirectory()
		if err != nil {
			t.Fatal(err)
		}
		gotInfo, err := os.Stat(got)
		if err != nil {
			t.Fatal(err)
		}
		expectedInfo, err := os.Stat(migrations)
		if err != nil {
			t.Fatal(err)
		}
		if !os.SameFile(gotInfo, expectedInfo) {
			t.Fatalf("expected %q and %q to refer to the same directory", migrations, got)
		}
	}
}
