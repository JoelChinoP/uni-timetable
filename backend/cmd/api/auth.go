package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/jackc/pgx/v5"

	"github.com/JoelChinoP/uni-timetable/backend/internal/database"
)

const (
	sessionCookieName = "uni_timetable_session"
	sessionDuration   = 7 * 24 * time.Hour
)

type AuthUser struct {
	ID            int32  `json:"id"`
	Email         string `json:"email"`
	DisplayName   string `json:"displayName"`
	Role          string `json:"role"`
	EmailVerified bool   `json:"emailVerified"`
}

type sessionRecord struct {
	user      AuthUser
	expiresAt time.Time
}

type sessionStore struct {
	mu       sync.RWMutex
	sessions map[string]sessionRecord
}

type authHandler struct {
	queries    *database.Queries
	verifier   *oidc.IDTokenVerifier
	sessions   *sessionStore
	adminEmail string
	secure     bool
}

type googleClaims struct {
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Name          string `json:"name"`
}

func newSessionStore() *sessionStore {
	// ponytail: sessions are instance-local; replace with a database table only if restarts require persistence.
	return &sessionStore{sessions: make(map[string]sessionRecord)}
}

func (sessions *sessionStore) create(user AuthUser) (string, error) {
	value := make([]byte, 32)
	if _, err := rand.Read(value); err != nil {
		return "", err
	}

	token := hex.EncodeToString(value)
	now := time.Now()
	sessions.mu.Lock()
	for existingToken, existing := range sessions.sessions {
		if now.After(existing.expiresAt) {
			delete(sessions.sessions, existingToken)
		}
	}
	sessions.sessions[token] = sessionRecord{
		user:      user,
		expiresAt: now.Add(sessionDuration),
	}
	sessions.mu.Unlock()
	return token, nil
}

func (sessions *sessionStore) get(token string) (AuthUser, bool) {
	sessions.mu.RLock()
	record, ok := sessions.sessions[token]
	sessions.mu.RUnlock()

	if !ok || time.Now().After(record.expiresAt) {
		if ok {
			sessions.delete(token)
		}
		return AuthUser{}, false
	}
	return record.user, true
}

func (sessions *sessionStore) delete(token string) {
	sessions.mu.Lock()
	delete(sessions.sessions, token)
	sessions.mu.Unlock()
}

func (auth *authHandler) currentUser(r *http.Request) (AuthUser, bool) {
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil || cookie.Value == "" {
		return AuthUser{}, false
	}
	return auth.sessions.get(cookie.Value)
}

func (auth *authHandler) requireAdmin(w http.ResponseWriter, r *http.Request) (AuthUser, bool) {
	user, ok := auth.currentUser(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return AuthUser{}, false
	}
	if user.Role != "ADMIN" {
		writeError(w, http.StatusForbidden, "forbidden")
		return AuthUser{}, false
	}
	return user, true
}

func (auth *authHandler) login(w http.ResponseWriter, r *http.Request) {
	if auth.verifier == nil {
		writeError(w, http.StatusServiceUnavailable, "Google authentication is not configured")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 64<<10)
	var request struct {
		Credential string `json:"credential"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid credential payload")
		return
	}
	if strings.TrimSpace(request.Credential) == "" {
		writeError(w, http.StatusBadRequest, "credential is required")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	idToken, err := auth.verifier.Verify(ctx, request.Credential)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid Google credential")
		return
	}

	var claims googleClaims
	if err := idToken.Claims(&claims); err != nil {
		writeError(w, http.StatusUnauthorized, "invalid Google claims")
		return
	}
	email := normalizeEmail(claims.Email)
	if email == "" || !claims.EmailVerified {
		writeError(w, http.StatusForbidden, "Google email is not verified")
		return
	}

	displayName := strings.TrimSpace(claims.Name)
	if displayName == "" {
		displayName = displayNameFromEmail(email)
	}

	user, err := auth.accountUser(ctx, email, displayName)
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, "database unavailable")
		return
	}
	user.EmailVerified = true

	token, err := auth.sessions.create(user)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "session unavailable")
		return
	}
	auth.setSessionCookie(w, token)
	writeJSON(w, http.StatusOK, user)
}

func (auth *authHandler) me(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.currentUser(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	writeJSON(w, http.StatusOK, user)
}

func (auth *authHandler) logout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie(sessionCookieName); err == nil {
		auth.sessions.delete(cookie.Value)
	}
	auth.clearSessionCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

func (auth *authHandler) setSessionCookie(w http.ResponseWriter, token string) {
	sameSite := http.SameSiteLaxMode
	if auth.secure {
		sameSite = http.SameSiteNoneMode
	}
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   auth.secure,
		SameSite: sameSite,
		MaxAge:   int(sessionDuration / time.Second),
		Expires:  time.Now().Add(sessionDuration),
	})
}

func (auth *authHandler) clearSessionCookie(w http.ResponseWriter) {
	sameSite := http.SameSiteLaxMode
	if auth.secure {
		sameSite = http.SameSiteNoneMode
	}
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   auth.secure,
		SameSite: sameSite,
		MaxAge:   -1,
	})
}

func (auth *authHandler) accountUser(ctx context.Context, email, displayName string) (AuthUser, error) {
	if auth.queries == nil {
		return AuthUser{Email: email, DisplayName: displayName, Role: defaultRole(email, auth.adminEmail)}, nil
	}

	if email == auth.adminEmail {
		user, err := auth.queries.UpsertAdminUser(ctx, database.UpsertAdminUserParams{
			Email:       email,
			DisplayName: displayName,
		})
		return AuthUser{
			ID:          user.ID,
			Email:       user.Email,
			DisplayName: user.DisplayName,
			Role:        string(user.Role),
		}, err
	}

	user, err := auth.queries.GetUserByEmail(ctx, email)
	if err != nil {
		if err == pgx.ErrNoRows {
			return AuthUser{Email: email, DisplayName: displayName, Role: "USER"}, nil
		}
		return AuthUser{}, err
	}
	return AuthUser{
		ID:          user.ID,
		Email:       user.Email,
		DisplayName: user.DisplayName,
		Role:        string(user.Role),
	}, nil
}

func defaultRole(email, adminEmail string) string {
	if email == adminEmail {
		return "ADMIN"
	}
	return "USER"
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func displayNameFromEmail(email string) string {
	local := strings.TrimSpace(strings.Split(email, "@")[0])
	parts := strings.FieldsFunc(local, func(r rune) bool {
		return r == '.' || r == '_' || r == '-'
	})
	for index, part := range parts {
		if part == "" {
			continue
		}
		parts[index] = strings.ToUpper(part[:1]) + part[1:]
	}
	if name := strings.Join(parts, " "); name != "" {
		return name
	}
	return email
}
