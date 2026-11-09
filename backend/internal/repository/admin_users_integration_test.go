package repository

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nik003kk/DeskBooker/backend/internal/models"
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
	var admin, target string
	for i, dest := range []*string{&admin, &target} {
		role := "employee"
		if i == 0 {
			role = "admin"
		}
		err := pool.QueryRow(ctx, "INSERT INTO public.users(email,password_hash,first_name,last_name,role) VALUES($1,'not-exposed','Admin','Fixture',$2) RETURNING id", []string{"role-admin@example.test", "role-target@example.test"}[i], role).Scan(dest)
		if err != nil {
			t.Fatal(err)
		}
	}
	changer := NewUserAdminStore(pool)
	for _, role := range []models.UserRole{models.RoleManager, models.RoleAdmin, models.RoleEmployee} {
		updated, err := changer.ChangeUserRole(ctx, admin, target, role)
		if err != nil || updated.ID != target || updated.Role != role {
			t.Fatalf("change role %s: %#v %v", role, updated, err)
		}
		actual, err := New(pool).GetUser(ctx, target)
		if err != nil || actual.Role != role {
			t.Fatal("role not persisted", err)
		}
	}
	if _, err := changer.ChangeUserRole(ctx, target, admin, models.RoleEmployee); !errors.Is(err, ErrAdminRequired) {
		t.Fatalf("employee changed admin: %v", err)
	}
	if _, err := changer.ChangeUserRole(ctx, admin, admin, models.RoleEmployee); !errors.Is(err, ErrSelfRoleChange) {
		t.Fatalf("admin changed self: %v", err)
	}

}
