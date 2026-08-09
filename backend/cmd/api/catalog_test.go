package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestValidateSessions(t *testing.T) {
	valid := []sessionPayload{{Day: "MONDAY", StartHourAcademic: 1, DurationHours: 2}}
	if err := validateSessions(valid); err != nil {
		t.Fatalf("valid payload rejected: %v", err)
	}

	cases := []struct {
		name     string
		sessions []sessionPayload
	}{
		{name: "empty", sessions: nil},
		{name: "too many", sessions: make([]sessionPayload, 15)},
		{name: "bad day", sessions: []sessionPayload{{Day: "SUNDAY", StartHourAcademic: 1, DurationHours: 2}}},
		{name: "start too low", sessions: []sessionPayload{{Day: "MONDAY", StartHourAcademic: 0, DurationHours: 2}}},
		{name: "start too high", sessions: []sessionPayload{{Day: "MONDAY", StartHourAcademic: 16, DurationHours: 1}}},
		{name: "zero duration", sessions: []sessionPayload{{Day: "MONDAY", StartHourAcademic: 1, DurationHours: 0}}},
		{name: "too long", sessions: []sessionPayload{{Day: "MONDAY", StartHourAcademic: 1, DurationHours: 7}}},
		{name: "overruns day", sessions: []sessionPayload{{Day: "MONDAY", StartHourAcademic: 14, DurationHours: 3}}},
		{name: "overlap same start", sessions: []sessionPayload{
			{Day: "MONDAY", StartHourAcademic: 1, DurationHours: 2},
			{Day: "MONDAY", StartHourAcademic: 1, DurationHours: 1},
		}},
		{name: "overlap inside span", sessions: []sessionPayload{
			{Day: "MONDAY", StartHourAcademic: 1, DurationHours: 4},
			{Day: "MONDAY", StartHourAcademic: 3, DurationHours: 2},
		}},
		{name: "contained span", sessions: []sessionPayload{
			{Day: "TUESDAY", StartHourAcademic: 3, DurationHours: 1},
			{Day: "TUESDAY", StartHourAcademic: 1, DurationHours: 6},
		}},
	}
	for _, test := range cases {
		if err := validateSessions(test.sessions); err == nil {
			t.Fatalf("%s: expected rejection", test.name)
		}
	}

	backToBack := []sessionPayload{
		{Day: "MONDAY", StartHourAcademic: 1, DurationHours: 2},
		{Day: "MONDAY", StartHourAcademic: 3, DurationHours: 2},
		{Day: "TUESDAY", StartHourAcademic: 1, DurationHours: 2},
	}
	if err := validateSessions(backToBack); err != nil {
		t.Fatalf("adjacent sessions on same day must be allowed: %v", err)
	}
}

func TestCatalogWriteGates(t *testing.T) {
	handler := routes(routeConfig{})
	posts := []struct {
		name string
		path string
	}{
		{"create classroom", "/classrooms"},
		{"create teacher", "/teachers"},
		{"create course", "/courses"},
		{"create group", "/groups"},
	}
	for _, test := range posts {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, test.path, strings.NewReader("{}"))
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)
			if response.Code != http.StatusUnauthorized {
				t.Fatalf("POST %s: got %d, want 401", test.path, response.Code)
			}
		})
	}

	t.Run("privileges fail closed without database", func(t *testing.T) {
		auth := &authHandler{sessions: newSessionStore()}
		token, err := auth.sessions.create(AuthUser{ID: 1, Email: "x@unsa.edu.pe", Role: "ADMIN"})
		if err != nil {
			t.Fatal(err)
		}
		catalog := &catalogHandler{auth: auth}
		request := httptest.NewRequest(http.MethodPost, "/groups", strings.NewReader("{}"))
		request.AddCookie(&http.Cookie{Name: sessionCookieName, Value: token})
		response := httptest.NewRecorder()
		catalog.groups(response, request)
		if response.Code != http.StatusServiceUnavailable {
			t.Fatalf("got %d, want 503", response.Code)
		}
	})
}

func TestSharedRoutesWithoutDatabase(t *testing.T) {
	handler := routes(routeConfig{})

	tests := []struct {
		name   string
		method string
		path   string
		status int
	}{
		{"dashboard needs db", http.MethodGet, "/planner/dashboard", http.StatusServiceUnavailable},
		{"share needs db", http.MethodPost, "/shared", http.StatusServiceUnavailable},
		{"get share malformed id", http.MethodGet, "/shared/abc", http.StatusNotFound},
		{"get share needs db", http.MethodGet, "/shared/AbCdEfGh12", http.StatusServiceUnavailable},
		{"classrooms list needs db", http.MethodGet, "/classrooms", http.StatusServiceUnavailable},
		{"teachers list needs db", http.MethodGet, "/teachers", http.StatusServiceUnavailable},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(test.method, test.path, nil)
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)
			want := test.status
			if want == 0 {
				want = http.StatusServiceUnavailable
			}
			if response.Code != want {
				t.Fatalf("%s %s: got %d, want %d", test.method, test.path, response.Code, want)
			}
		})
	}
}

func TestSessionStoreExpiry(t *testing.T) {
	now := time.Now()
	sessions := newSessionStore("0123456789abcdef0123456789abcdef")
	sessions.now = func() time.Time { return now }
	token, err := sessions.create(AuthUser{ID: 1})
	if err != nil {
		t.Fatal(err)
	}
	now = now.Add(sessionDuration)

	if _, ok := sessions.get(token); ok {
		t.Fatal("expired session must not authenticate")
	}
}
