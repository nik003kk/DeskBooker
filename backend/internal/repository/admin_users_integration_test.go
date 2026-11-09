package repository

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nik003kk/DeskBooker/backend/internal/auth"
	"github.com/nik003kk/DeskBooker/backend/internal/models"
	"net/http/httptest"
	"strings"
	"sync"
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

	if _, err := changer.ChangeUserRole(ctx, admin, strings.ToUpper(admin), models.RoleEmployee); !errors.Is(err, ErrSelfRoleChange) {
		t.Fatalf("uppercase UUID bypassed self protection: %v", err)
	}
	if _, err := changer.ChangeUserRole(ctx, admin, admin, models.RoleAdmin); !errors.Is(err, ErrSelfRoleChange) {
		t.Fatalf("self assignment was allowed: %v", err)
	}
	var left, right string
	for i, dest := range []*string{&left, &right} {
		if err := pool.QueryRow(ctx, "INSERT INTO public.users(email,password_hash,first_name,last_name,role) VALUES($1,'hash','Concurrent','Admin','admin') RETURNING id", []string{"left-admin@example.test", "right-admin@example.test"}[i]).Scan(dest); err != nil {
			t.Fatal(err)
		}
	}
	results := make(chan error, 2)
	var wg sync.WaitGroup
	for _, pair := range [][2]string{{left, right}, {right, left}} {
		wg.Add(1)
		go func(actor, target string) {
			defer wg.Done()
			_, err := changer.ChangeUserRole(ctx, actor, target, models.RoleEmployee)
			results <- err
		}(pair[0], pair[1])
	}
	wg.Wait()
	close(results)
	success, denied := 0, 0
	for err := range results {
		if err == nil {
			success++
		} else if errors.Is(err, ErrAdminRequired) {
			denied++
		} else {
			t.Fatal(err)
		}
	}
	if success != 1 || denied != 1 {
		t.Fatalf("concurrent mutual demotion: success=%d denied=%d", success, denied)
	}
	var remaining int
	if err := pool.QueryRow(ctx, "SELECT count(*) FROM public.users WHERE id=ANY($1::uuid[]) AND role='admin' AND status='active'", []string{left, right}).Scan(&remaining); err != nil || remaining != 1 {
		t.Fatalf("no active administrator remained: %d %v", remaining, err)
	}

	// Reuse one JWT throughout real PostgreSQL role changes. Both privileges
	// and ordinary authenticated routes must observe the next request's user.
	tokens, err := auth.NewJWTService("postgres-current-rights-test-secret-at-least-32-bytes")
	if err != nil {
		t.Fatal(err)
	}
	token, _, err := tokens.Issue(target)
	if err != nil {
		t.Fatal(err)
	}
	router := gin.New()
	router.GET("/private", auth.RequireCurrentJWT(tokens, store), auth.MeHandler(store))
	router.GET("/admin", auth.RequireCurrentJWT(tokens, store), auth.RequireAdministrator(store), func(c *gin.Context) { c.Status(200) })
	router.GET("/manager", auth.RequireCurrentJWT(tokens, store), auth.RequireResourceManager(store), func(c *gin.Context) { c.Status(200) })
	checkAccess := func(path string, want int) {
		t.Helper()
		r := httptest.NewRecorder()
		request := httptest.NewRequest("GET", path, nil)
		request.Header.Set("Authorization", "Bearer "+token)
		router.ServeHTTP(r, request)
		if r.Code != want {
			t.Fatalf("%s: %d %s", path, r.Code, r.Body.String())
		}
	}
	for _, state := range []struct {
		role           models.UserRole
		admin, manager int
	}{{models.RoleManager, 403, 200}, {models.RoleAdmin, 200, 200}, {models.RoleEmployee, 403, 403}} {
		if _, err := changer.ChangeUserRole(ctx, admin, target, state.role); err != nil {
			t.Fatal(err)
		}
		checkAccess("/admin", state.admin)
		checkAccess("/manager", state.manager)
		checkAccess("/private", 200)
	}
	if _, err := pool.Exec(ctx, "UPDATE public.users SET status='suspended' WHERE id=$1", target); err != nil {
		t.Fatal(err)
	}
	checkAccess("/private", 403)
	if _, err := pool.Exec(ctx, "DELETE FROM public.users WHERE id=$1", target); err != nil {
		t.Fatal(err)
	}
	checkAccess("/private", 401)

}
