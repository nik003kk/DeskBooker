package repository

import (
	"context"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nik003kk/DeskBooker/backend/internal/auth"
	"github.com/nik003kk/DeskBooker/backend/internal/models"
	"github.com/nik003kk/DeskBooker/backend/internal/resources"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type categoryAPIResource struct {
	models.ResourceDTO
	Features []string `json:"features"`
}

func testResourceCategories(t *testing.T, pool *pgxpool.Pool, employee string) {
	ctx := context.Background()
	store := New(pool)
	var admin string
	if err := pool.QueryRow(ctx, "INSERT INTO public.users(email,password_hash,first_name,last_name,role) VALUES('category-admin@example.test','hash','Category','Admin','admin') RETURNING id").Scan(&admin); err != nil {
		t.Fatal(err)
	}
	tokens, err := auth.NewJWTService("resource-category-test-secret-at-least-32-bytes")
	if err != nil {
		t.Fatal(err)
	}
	adminToken, _, _ := tokens.Issue(admin)
	employeeToken, _, _ := tokens.Issue(employee)
	router := gin.New()
	current := auth.RequireCurrentJWT(tokens, store)
	manager := auth.RequireResourceManager(store)
	router.GET("/api/resources", current, resources.SearchHandler(store))
	router.POST("/api/resources", current, manager, resources.CreateHandler(store))
	router.PATCH("/api/resources/:id", current, manager, resources.EditHandler(store))
	call := func(t *testing.T, method, path, body, token string, want int) *httptest.ResponseRecorder {
		t.Helper()
		request := httptest.NewRequest(method, path, strings.NewReader(body))
		request.Header.Set("Authorization", "Bearer "+token)
		request.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		if response.Code != want {
			t.Fatalf("%s %s: %d %s", method, path, response.Code, response.Body.String())
		}
		return response
	}
	find := func(t *testing.T, kind, id, suffix string) *categoryAPIResource {
		t.Helper()
		response := call(t, "GET", "/api/resources?type="+kind+suffix, "", employeeToken, 200)
		var body struct {
			Resources []categoryAPIResource `json:"resources"`
		}
		if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
			t.Fatal(err)
		}
		for _, item := range body.Resources {
			if string(item.Type) != kind || !item.IsActive {
				t.Fatal("wrong category or inactive resource returned")
			}
			if item.ID == id {
				copy := item
				return &copy
			}
		}
		return nil
	}
	kinds := []models.ResourceType{models.ResourceDesk, models.ResourceMeetingRoom, models.ResourceParkingSpot}
	loc, _ := time.LoadLocation("Europe/Ljubljana")
	start := time.Date(2099, 1, 5, 9, 0, 0, 0, loc)
	end := start.Add(time.Hour)
	for index, kind := range kinds {
		t.Run(string(kind), func(t *testing.T) {
			raw, _ := json.Marshal(models.CreateResourceInput{Name: "Category fixture " + string(kind), Type: kind, Capacity: 1, IsActive: true})
			response := call(t, "POST", "/api/resources", string(raw), adminToken, 201)
			var created struct {
				Resource models.ResourceDTO `json:"resource"`
			}
			if err := json.Unmarshal(response.Body.Bytes(), &created); err != nil {
				t.Fatal(err)
			}
			id := created.Resource.ID
			if found := find(t, string(kind), id, ""); found == nil || found.Name != created.Resource.Name {
				t.Fatal("admin-created resource missing")
			}
			for _, other := range kinds {
				if other != kind && find(t, string(other), id, "") != nil {
					t.Fatal("resource leaked into another category")
				}
			}
			name := "Category updated " + string(kind)
			patch, _ := json.Marshal(map[string]any{"name": name, "capacity": 2, "floor": 0, "features": []string{"CategoryFeature", "USB-C"}})
			call(t, "PATCH", "/api/resources/"+id, string(patch), adminToken, 200)
			updated := find(t, string(kind), id, "&features=CategoryFeature,USB-C")
			if updated == nil || updated.Name != name || updated.Capacity != 2 || updated.Floor == nil || *updated.Floor != 0 || len(updated.Features) != 2 {
				t.Fatalf("edited fields/equipment missing: %#v", updated)
			}
			interval := "&date=2099-01-05&start=09:00&end=10:00"
			if find(t, string(kind), id, interval) == nil {
				t.Fatal("unbooked resource unavailable")
			}
			if _, err := NewBookingStore(pool).CreateConfirmedReservation(ctx, models.CreateReservationInput{UserID: employee, ResourceID: id, StartsAt: start, EndsAt: end}); err != nil {
				t.Fatal(err)
			}
			if find(t, string(kind), id, interval) != nil {
				t.Fatal("booked resource remained available")
			}
			if find(t, string(kind), id, "") == nil {
				t.Fatal("catalog without interval lost an active resource")
			}
			moved := kinds[(index+1)%len(kinds)]
			call(t, "PATCH", "/api/resources/"+id, `{"type":"`+string(moved)+`"}`, adminToken, 200)
			if find(t, string(kind), id, "") != nil || find(t, string(moved), id, "") == nil {
				t.Fatal("edited resource type did not move categories")
			}
			call(t, "PATCH", "/api/resources/"+id, `{"is_active":false}`, adminToken, 200)
			if find(t, string(moved), id, "") != nil {
				t.Fatal("deactivated resource remained in category")
			}
		})
	}
}
