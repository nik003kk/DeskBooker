package repository

import (
	"context"
	"encoding/json"
	"github.com/jackc/pgx/v5/pgxpool"
	"strings"
	"testing"
)

func testUserAdministration(t *testing.T, pool *pgxpool.Pool) {
	ctx := context.Background()
	store := New(pool)
	page, err := store.ListAdminUsers(ctx, 1, 0)
	if err != nil || page.Total < 2 || len(page.Users) != 1 || page.Users[0].Email == "" || page.Limit != 1 {
		t.Fatalf("paginated users: %#v %v", page, err)
	}
	raw, err := json.Marshal(page)
	if err != nil || strings.Contains(string(raw), "password") || strings.Contains(string(raw), "hash") {
		t.Fatal("credentials exposed in user listing")
	}
	empty, err := store.ListAdminUsers(ctx, 1, 1000000)
	if err != nil || len(empty.Users) != 0 || empty.Total != page.Total {
		t.Fatalf("empty page lost total: %#v %v", empty, err)
	}
}
