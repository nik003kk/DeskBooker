package main

import (
	"strings"
	"testing"
)

func TestRequireDevelopment(t *testing.T) {
	for _, tc := range []struct {
		env   string
		args  []string
		allow bool
	}{
		{"development", nil, true},
		{"production", nil, false},
		{"", nil, false},
		{"test", nil, false},
		{"development", []string{"drop"}, false},
	} {
		err := requireDevelopment(tc.env, tc.args)
		if (err == nil) != tc.allow {
			t.Errorf("env=%q args=%v: unexpected error %v", tc.env, tc.args, err)
		}
	}
}

func TestSeedContainsAllResourceCategories(t *testing.T) {
	for _, table := range []string{
		"demo_seed.users",
		"demo_seed.desks",
		"demo_seed.meeting_rooms",
		"demo_seed.parking_spaces",
		"demo_seed.equipment",
	} {
		if !strings.Contains(seedSQL, "INSERT INTO "+table) {
			t.Errorf("fixture SQL missing %s", table)
		}
	}
	if count := strings.Count(seedSQL, "ON CONFLICT (id) DO UPDATE"); count != 5 {
		t.Errorf("expected 5 idempotent upserts, got %d", count)
	}
	for _, role := range []string{"'admin'", "'manager'", "'employee'"} {
		if !strings.Contains(seedSQL, role) {
			t.Errorf("missing test role %s", role)
		}
	}
}
