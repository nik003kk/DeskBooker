package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAllowedOrigins(t *testing.T) {
	defaults, err := AllowedOrigins("")
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := defaults["http://localhost:5173"]; !ok {
		t.Fatal("expected local Vite origin")
	}

	configured, err := AllowedOrigins("http://localhost:5173, https://office.example.com")
	if err != nil || len(configured) != 2 {
		t.Fatalf("expected two configured origins, got %v: %v", configured, err)
	}
	for _, invalid := range []string{"*", "https://example.com/path", "https://example.com/", "ftp://example.com", "https://user:pw@example.com", "http://localhost:5173,"} {
		if _, err := AllowedOrigins(invalid); err == nil {
			t.Errorf("expected invalid CORS origin %q to fail", invalid)
		}
	}
}

func TestCORSRequests(t *testing.T) {
	router := NewRouter(fakeDatabase{}, mustOrigins(t, ""))
	tests := []struct {
		name, origin, method, path, requestMethod string
		wantCode                                 int
		wantOrigin                               string
	}{
		{"allowed GET", "http://localhost:5173", "GET", "/api/health", "", 200, "http://localhost:5173"},
		{"disallowed GET", "https://evil.example", "GET", "/api/health", "", 200, ""},
		{"no origin", "", "GET", "/api/health", "", 200, ""},
		{"allowed preflight", "http://localhost:5173", "OPTIONS", "/api/health", "GET", 204, "http://localhost:5173"},
		{"disallowed preflight", "https://evil.example", "OPTIONS", "/api/health", "GET", 403, ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(tc.method, tc.path, nil)
			if tc.origin != "" {
				req.Header.Set("Origin", tc.origin)
			}
			if tc.requestMethod != "" {
				req.Header.Set("Access-Control-Request-Method", tc.requestMethod)
				req.Header.Set("Access-Control-Request-Headers", "content-type")
			}
			router.ServeHTTP(w, req)
			if w.Code != tc.wantCode {
				t.Fatalf("status: want %d, got %d: %s", tc.wantCode, w.Code, w.Body.String())
			}
			if got := w.Header().Get("Access-Control-Allow-Origin"); got != tc.wantOrigin {
				t.Fatalf("origin: want %q, got %q", tc.wantOrigin, got)
			}
			if tc.requestMethod != "" && tc.wantCode == 204 {
				if w.Header().Get("Access-Control-Allow-Headers") == "" {
					t.Fatal("expected preflight allowed headers")
				}
				if w.Header().Get("Vary") != "Origin" {
					t.Fatal("expected Vary: Origin")
				}
			}
			if tc.wantCode == 403 && !strings.Contains(w.Body.String(), "cors_origin_denied") {
				t.Fatalf("expected CORS JSON error: %s", w.Body.String())
			}
		})
	}
}

func TestRESTNotFoundUsesJSON(t *testing.T) {
	router := NewRouter(fakeDatabase{}, mustOrigins(t, ""))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/not-found", nil))
	if w.Code != http.StatusNotFound || !strings.Contains(w.Header().Get("Content-Type"), "application/json") ||
		!strings.Contains(w.Body.String(), "not_found") {
		t.Fatalf("expected JSON 404, got %d %q", w.Code, w.Body.String())
	}
}

func mustOrigins(t *testing.T, config string) map[string]struct{} {
	t.Helper()
	result, err := AllowedOrigins(config)
	if err != nil {
		t.Fatal(err)
	}
	return result
}
