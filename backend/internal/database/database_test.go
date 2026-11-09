package database

import (
	"net/url"
	"strings"
	"testing"
)

func env(values map[string]string) func(string) string {
	return func(key string) string { return values[key] }
}

func TestConnectionStringDefaults(t *testing.T) {
	dsn, err := ConnectionString(env(map[string]string{
		"POSTGRES_PASSWORD": "my password@#",
		"POSTGRES_PORT":     "5433",
	}))
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := url.Parse(dsn)
	if err != nil {
		t.Fatal(err)
	}
	if parsed.Host != "localhost:5433" || parsed.Path != "/deskbooker" {
		t.Fatalf("unexpected database URL: %s", parsed.Redacted())
	}
	password, ok := parsed.User.Password()
	if !ok || password != "my password@#" {
		t.Fatal("password was not safely URL-encoded")
	}
	if parsed.Query().Get("sslmode") != "disable" {
		t.Fatal("expected local sslmode=disable")
	}
}

func TestConnectionStringRequiresPassword(t *testing.T) {
	_, err := ConnectionString(env(map[string]string{}))
	if err == nil || !strings.Contains(err.Error(), "POSTGRES_PASSWORD") {
		t.Fatalf("expected missing password error, got: %v", err)
	}
}

func TestConnectionStringInvalidPort(t *testing.T) {
	for _, port := range []string{"abc", "0", "65536"} {
		_, err := ConnectionString(env(map[string]string{
			"POSTGRES_PASSWORD": "test",
			"POSTGRES_PORT":     port,
		}))
		if err == nil {
			t.Fatalf("expected error for port %q", port)
		}
	}
}

func TestConnectionStringExplicitURL(t *testing.T) {
	want := "postgres://someone:secret@db.example:5432/app?sslmode=require"
	got, err := ConnectionString(env(map[string]string{"DATABASE_URL": want}))
	if err != nil || got != want {
		t.Fatalf("DATABASE_URL override not respected: %v", err)
	}
}
