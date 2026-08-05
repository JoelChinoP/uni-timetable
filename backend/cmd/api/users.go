package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/mail"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/JoelChinoP/uni-timetable/backend/internal/database"
)

func (auth *authHandler) users(w http.ResponseWriter, r *http.Request) {
	if _, ok := auth.requireAdmin(w, r); !ok {
		return
	}
	if auth.queries == nil {
		writeError(w, http.StatusServiceUnavailable, "database is not configured")
		return
	}

	switch r.Method {
	case http.MethodGet:
		auth.listUsers(w, r)
	case http.MethodPost:
		auth.createUser(w, r)
	default:
		w.Header().Set("Allow", "GET, POST")
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (auth *authHandler) listUsers(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	rows, err := auth.queries.ListUsers(ctx)
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, "database unavailable")
		return
	}

	users := make([]AuthUser, 0, len(rows))
	for _, row := range rows {
		users = append(users, AuthUser{
			ID:            row.ID,
			Email:         row.Email,
			DisplayName:   row.DisplayName,
			Role:          string(row.Role),
			EmailVerified: true,
		})
	}
	writeJSON(w, http.StatusOK, users)
}

func (auth *authHandler) createUser(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 16<<10)
	var request struct {
		Email       string `json:"email"`
		DisplayName string `json:"displayName"`
	}
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid user payload")
		return
	}

	email := normalizeEmail(request.Email)
	displayName := strings.TrimSpace(request.DisplayName)
	if !validEmail(email) {
		writeError(w, http.StatusBadRequest, "invalid email")
		return
	}
	if len(displayName) < 2 || len(displayName) > 100 {
		writeError(w, http.StatusBadRequest, "displayName must be 2-100 characters")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	row, err := auth.queries.CreateUser(ctx, database.CreateUserParams{
		Email:       email,
		DisplayName: displayName,
	})
	if err != nil {
		var pgError *pgconn.PgError
		if errors.As(err, &pgError) && pgError.Code == "23505" {
			writeError(w, http.StatusConflict, "email is already registered")
			return
		}
		writeError(w, http.StatusServiceUnavailable, "database unavailable")
		return
	}

	user := AuthUser{
		ID:            row.ID,
		Email:         row.Email,
		DisplayName:   row.DisplayName,
		Role:          string(row.Role),
		EmailVerified: true,
	}
	writeJSON(w, http.StatusCreated, user)
}

func validEmail(email string) bool {
	address, err := mail.ParseAddress(email)
	return err == nil && address.Address == email
}
