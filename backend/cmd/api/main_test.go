package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRoutesWithoutDatabase(t *testing.T) {
	handler := routes(nil, "test-secret")
	tests := []struct {
		name   string
		path   string
		auth   string
		status int
	}{
		{name: "health", path: "/health", status: http.StatusOK},
		{name: "ready", path: "/ready", status: http.StatusServiceUnavailable},
		{name: "keepalive rejects missing auth", path: "/internal/keepalive", status: http.StatusUnauthorized},
		{name: "keepalive needs database", path: "/internal/keepalive", auth: "Bearer test-secret", status: http.StatusServiceUnavailable},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, test.path, nil)
			request.Header.Set("Authorization", test.auth)
			response := httptest.NewRecorder()

			handler.ServeHTTP(response, request)
			if response.Code != test.status {
				t.Fatalf("got status %d, want %d", response.Code, test.status)
			}
		})
	}
}
