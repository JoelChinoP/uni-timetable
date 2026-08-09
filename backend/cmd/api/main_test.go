package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRoutesWithoutDatabase(t *testing.T) {
	handler := routes(routeConfig{
		cronSecret:     "test-secret",
		allowedOrigins: []string{"http://127.0.0.1:5173"},
	})
	tests := []struct {
		name   string
		method string
		path   string
		auth   string
		status int
	}{
		{name: "health", method: http.MethodGet, path: "/health", status: http.StatusOK},
		{name: "ready", method: http.MethodGet, path: "/ready", status: http.StatusServiceUnavailable},
		{name: "auth session missing", method: http.MethodGet, path: "/auth/me", status: http.StatusUnauthorized},
		{name: "login needs Google config", method: http.MethodPost, path: "/auth/login", status: http.StatusServiceUnavailable},
		{name: "users need auth", method: http.MethodGet, path: "/users", status: http.StatusUnauthorized},
		{name: "keepalive rejects missing auth", method: http.MethodGet, path: "/internal/keepalive", status: http.StatusUnauthorized},
		{name: "keepalive needs database", method: http.MethodGet, path: "/internal/keepalive", auth: "Bearer test-secret", status: http.StatusServiceUnavailable},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(test.method, test.path, nil)
			request.Header.Set("Authorization", test.auth)
			response := httptest.NewRecorder()

			handler.ServeHTTP(response, request)
			if response.Code != test.status {
				t.Fatalf("%s %s: got status %d, want %d", test.method, test.path, response.Code, test.status)
			}
		})
	}
}

func TestSessionStore(t *testing.T) {
	sessions := newSessionStore("0123456789abcdef0123456789abcdef")
	secondInstance := newSessionStore("0123456789abcdef0123456789abcdef")
	user := AuthUser{ID: 1, Email: "jchinop@unsa.edu.pe", DisplayName: "Joel", Role: "ADMIN", EmailVerified: true}

	token, err := sessions.create(user)
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	if got, ok := secondInstance.get(token); !ok || got != user {
		t.Fatalf("stored session = %+v, %t; want %+v, true", got, ok, user)
	}
	if _, ok := secondInstance.get(token + "tampered"); ok {
		t.Fatal("tampered session authenticated")
	}
}

func TestIdentityHelpers(t *testing.T) {
	if got := displayNameFromEmail("joel.chino@unsa.edu.pe"); got != "Joel Chino" {
		t.Fatalf("displayNameFromEmail = %q", got)
	}
	if !validEmail("joel@unsa.edu.pe") || validEmail("not-an-email") {
		t.Fatal("validEmail failed")
	}
}

func TestUserPageParamsRejectsOffsetOverflow(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/users?page=214748366&pageSize=10", nil)
	if _, _, ok := userPageParams(request); ok {
		t.Fatal("userPageParams accepted an offset that overflows int32")
	}
}
