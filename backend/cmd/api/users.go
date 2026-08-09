package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/mail"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/JoelChinoP/uni-timetable/backend/internal/database"
)

type managedUserJSON struct {
	ID          int32  `json:"id"`
	Email       string `json:"email"`
	DisplayName string `json:"displayName"`
	Role        string `json:"role"`
}

func (auth *authHandler) users(w http.ResponseWriter, r *http.Request) {
	if _, ok := auth.requireAdmin(w, r); !ok {
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
	page, pageSize, ok := userPageParams(r)
	if !ok {
		writeError(w, http.StatusBadRequest, "page y pageSize deben ser enteros positivos; pageSize máximo 100")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	rows, err := auth.queries.ListUsersPage(ctx, database.ListUsersPageParams{
		PageSize: int32(pageSize), OffsetRows: int32((page - 1) * pageSize),
	})
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, "database unavailable")
		return
	}
	total, err := auth.queries.CountUsers(ctx)
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, "database unavailable")
		return
	}
	users := make([]managedUserJSON, 0, len(rows))
	for _, row := range rows {
		users = append(users, managedUserJSON{ID: row.ID, Email: row.Email, DisplayName: row.DisplayName, Role: string(row.Role)})
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"items": users, "page": page, "pageSize": pageSize, "total": total,
	})
}

func userPageParams(r *http.Request) (int, int, bool) {
	const maxOffset = 1<<31 - 1
	page, pageSize := 1, 10
	var err error
	if value := r.URL.Query().Get("page"); value != "" {
		page, err = strconv.Atoi(value)
		if err != nil {
			return 0, 0, false
		}
	}
	if value := r.URL.Query().Get("pageSize"); value != "" {
		pageSize, err = strconv.Atoi(value)
		if err != nil {
			return 0, 0, false
		}
	}
	return page, pageSize, page >= 1 && pageSize >= 1 && pageSize <= 100 && page <= maxOffset/pageSize+1
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
		writeError(w, http.StatusBadRequest, "correo inválido")
		return
	}
	if len(displayName) < 2 || len(displayName) > 100 {
		writeError(w, http.StatusBadRequest, "el nombre debe tener entre 2 y 100 caracteres")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()
	row, err := auth.queries.CreateUser(ctx, database.CreateUserParams{Email: email, DisplayName: displayName})
	if err != nil {
		if writeUserConflict(err, w) {
			return
		}
		writeError(w, http.StatusServiceUnavailable, "database unavailable")
		return
	}
	writeJSON(w, http.StatusCreated, managedUserJSON{ID: row.ID, Email: row.Email, DisplayName: row.DisplayName, Role: string(row.Role)})
}

func (auth *authHandler) userByID(w http.ResponseWriter, r *http.Request) {
	identity, ok := auth.sessionUser(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	if auth.db == nil || auth.queries == nil {
		writeError(w, http.StatusServiceUnavailable, "database is not configured")
		return
	}
	id, ok := pathID(r, "id")
	if !ok {
		writeError(w, http.StatusBadRequest, "id inválido")
		return
	}

	var request struct {
		Email       string `json:"email"`
		DisplayName string `json:"displayName"`
		Role        string `json:"role"`
	}
	if r.Method == http.MethodPut {
		r.Body = http.MaxBytesReader(w, r.Body, 16<<10)
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&request); err != nil {
			writeError(w, http.StatusBadRequest, "invalid user payload")
			return
		}
		request.Email = normalizeEmail(request.Email)
		request.DisplayName = strings.TrimSpace(request.DisplayName)
		if !validEmail(request.Email) || len(request.DisplayName) < 2 || len(request.DisplayName) > 100 || (request.Role != "ADMIN" && request.Role != "EDITOR") {
			writeError(w, http.StatusBadRequest, "correo, nombre o rol inválido")
			return
		}
	} else if r.Method != http.MethodDelete {
		w.Header().Set("Allow", "PUT, DELETE")
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	tx, err := auth.db.Begin(ctx)
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, "database unavailable")
		return
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, "SELECT pg_advisory_xact_lock($1)", int64(0x55534552)); err != nil {
		writeError(w, http.StatusServiceUnavailable, "database unavailable")
		return
	}
	queries := auth.queries.WithTx(tx)
	actor, err := auth.accountUserWithQueries(ctx, queries, identity.Email, identity.DisplayName)
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, "database unavailable")
		return
	}
	if actor.Role != "ADMIN" {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}
	target, err := queries.GetUserByID(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		writeError(w, http.StatusNotFound, "usuario no encontrado")
		return
	}
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, "database unavailable")
		return
	}
	if actor.ID == target.ID || target.Email == auth.adminEmail {
		writeError(w, http.StatusConflict, "la cuenta administrativa activa no puede modificarse ni eliminarse")
		return
	}
	if target.Role == database.AppUserRoleADMIN && (r.Method == http.MethodDelete || request.Role != "ADMIN") {
		count, err := queries.CountAdmins(ctx)
		if err != nil {
			writeError(w, http.StatusServiceUnavailable, "database unavailable")
			return
		}
		if count <= 1 {
			writeError(w, http.StatusConflict, "debe permanecer al menos un administrador")
			return
		}
	}

	if r.Method == http.MethodDelete {
		if _, err := queries.DeleteUser(ctx, id); err != nil {
			writeError(w, http.StatusServiceUnavailable, "database unavailable")
			return
		}
		if err := tx.Commit(ctx); err != nil {
			writeError(w, http.StatusServiceUnavailable, "database unavailable")
			return
		}
		w.WriteHeader(http.StatusNoContent)
		return
	}

	row, err := queries.UpdateUser(ctx, database.UpdateUserParams{
		ID: id, Email: request.Email, DisplayName: request.DisplayName, Role: database.AppUserRole(request.Role),
	})
	if err != nil {
		if writeUserConflict(err, w) {
			return
		}
		writeError(w, http.StatusServiceUnavailable, "database unavailable")
		return
	}
	if err := tx.Commit(ctx); err != nil {
		writeError(w, http.StatusServiceUnavailable, "database unavailable")
		return
	}
	writeJSON(w, http.StatusOK, managedUserJSON{ID: row.ID, Email: row.Email, DisplayName: row.DisplayName, Role: string(row.Role)})
}

func writeUserConflict(err error, w http.ResponseWriter) bool {
	var pgError *pgconn.PgError
	if errors.As(err, &pgError) && pgError.Code == "23505" {
		writeError(w, http.StatusConflict, "el correo ya está registrado")
		return true
	}
	return false
}

func validEmail(email string) bool {
	address, err := mail.ParseAddress(email)
	return err == nil && address.Address == email
}
