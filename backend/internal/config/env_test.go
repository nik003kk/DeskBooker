package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadReader(t *testing.T) {
	values := map[string]string{"POSTGRES_PORT": "5433"}
	lookup := func(key string) (string, bool) { v, ok := values[key]; return v, ok }
	setenv := func(key, value string) error { values[key] = value; return nil }
	err := LoadReader(strings.NewReader("# comment\nPOSTGRES_PORT=5432\nPOSTGRES_PASSWORD='secret=with spaces'\nexport PORT=8090\n"), lookup, setenv)
	if err != nil {
		t.Fatal(err)
	}
	if values["POSTGRES_PORT"] != "5433" {
		t.Fatal("existing environment variable was overwritten")
	}
	if values["POSTGRES_PASSWORD"] != "secret=with spaces" || values["PORT"] != "8090" {
		t.Fatal("expected .env values to be parsed")
	}
}

func TestLoadReaderRejectsInvalidSyntaxWithoutSecrets(t *testing.T) {
	for _, text := range []string{"BAD KEY=sensitive", "BAD_ENTRY", "SECRET='unterminated"} {
		err := LoadReader(strings.NewReader(text), func(string) (string, bool) { return "", false }, func(string, string) error { return nil })
		if err == nil {
			t.Fatalf("expected error for invalid .env entry")
		}
		if strings.Contains(err.Error(), "sensitive") || strings.Contains(err.Error(), "unterminated") {
			t.Fatal("error message exposed secret value")
		}
	}
}

func TestPort(t *testing.T) {
	for _, tc := range []struct {
		value, expected string
		ok              bool
	}{
		{"", "8080", true},
		{"8090", "8090", true},
		{"0", "", false},
		{"65536", "", false},
		{"invalid", "", false},
	} {
		got, err := Port(func(string) string { return tc.value })
		if (err == nil) != tc.ok || got != tc.expected {
			t.Fatalf("Port(%q) = %q, %v", tc.value, got, err)
		}
	}
}

func TestLoadLocalEnvFromBackendDirectory(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "compose.yaml"), []byte("services: {}\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".env"), []byte("DB008_TEST_VALUE=from-root\n"), 0600); err != nil {
		t.Fatal(err)
	}
	backend := filepath.Join(root, "backend")
	if err := os.Mkdir(backend, 0700); err != nil {
		t.Fatal(err)
	}
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(backend); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(originalDir) })
	t.Setenv("DB008_TEST_VALUE", "already-set")
	if err := LoadLocalEnv(); err != nil {
		t.Fatal(err)
	}
	if got := os.Getenv("DB008_TEST_VALUE"); got != "already-set" {
		t.Fatalf("expected explicit env override, got %q", got)
	}
}
