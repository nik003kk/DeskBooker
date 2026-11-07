package server

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

type fakeDatabase struct{ err error }

func (db fakeDatabase) Ping(context.Context) error { return db.err }

func TestHealthEndpoint(t *testing.T) {
	router := NewRouter(fakeDatabase{})
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/health", nil)

	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", response.Code, response.Body.String())
	}

	var body struct {
		Status  string `json:"status"`
		Service string `json:"service"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid JSON response: %v", err)
	}
	if body.Status != "ok" || body.Service != "deskbooker-api" {
		t.Fatalf("unexpected response: %+v", body)
	}
}

func TestDatabaseHealthEndpoint(t *testing.T) {
	tests := []struct {
		name     string
		db       fakeDatabase
		wantCode int
	}{
		{name: "connected", db: fakeDatabase{}, wantCode: http.StatusOK},
		{name: "unavailable", db: fakeDatabase{err: errors.New("connection failed")}, wantCode: http.StatusServiceUnavailable},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			router := NewRouter(tc.db)
			response := httptest.NewRecorder()
			router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/health/db", nil))
			if response.Code != tc.wantCode {
				t.Fatalf("expected %d, got %d: %s", tc.wantCode, response.Code, response.Body.String())
			}
			if tc.db.err != nil && response.Body.String() != "{\"status\":\"unavailable\"}" {
				t.Fatalf("unexpected error response: %s", response.Body.String())
			}
		})
	}
}

func TestUnknownEndpointReturns404(t *testing.T) {
	router := NewRouter(fakeDatabase{})
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/unknown", nil))

	if response.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", response.Code)
	}
}
