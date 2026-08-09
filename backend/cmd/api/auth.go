package main

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JoelChinoP/uni-timetable/backend/internal/database"
)

const (
	sessionCookieName = "uni_timetable_session"
	sessionDuration   = time.Hour
)

type AuthUser struct {
	ID            int32  `json:"id"`
	Email         string `json:"email"`
	DisplayName   string `json:"displayName"`
	Role          string `json:"role"`
	EmailVerified bool   `json:"emailVerified"`
	AvatarURL     string `json:"avatarUrl,omitempty"`
}

type sessionClaims struct {
	User      AuthUser `json:"user"`
	ExpiresAt int64    `json:"expiresAt"`
	Nonce     string   `json:"nonce"`
}

type sessionStore struct {
	secret []byte
	now    func() time.Time
}

type authHandler struct {
	queries    *database.Queries
	db         *pgxpool.Pool
	verifier   *oidc.IDTokenVerifier
	sessions   *sessionStore
	adminEmail string
	secure     bool
}

type googleClaims struct {
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
}

func newSessionStore(configuredSecret ...string) *sessionStore {
	secret := ""
	if len(configuredSecret) > 0 {
		secret = strings.TrimSpace(configuredSecret[0])
	}
	if secret == "" {
		value := make([]byte, 32)
		if _, err := rand.Read(value); err != nil {
			panic("session entropy unavailable")
		}
		return &sessionStore{secret: value, now: time.Now}
	}
	return &sessionStore{secret: []byte(secret), now: time.Now}
}

func (sessions *sessionStore) create(user AuthUser) (string, error) {
	nonce := make([]byte, 12)
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}
	payload, err := json.Marshal(sessionClaims{
		User:      user,
		ExpiresAt: sessions.now().Add(sessionDuration).Unix(),
		Nonce:     base64.RawURLEncoding.EncodeToString(nonce),
	})
	if err != nil {
		return "", err
	}
	encoded := base64.RawURLEncoding.EncodeToString(payload)
	return encoded + "." + base64.RawURLEncoding.EncodeToString(sessions.sign([]byte(encoded))), nil
}

func (sessions *sessionStore) get(token string) (AuthUser, bool) {
	encoded, signature, ok := strings.Cut(token, ".")
	if !ok || len(encoded) > 4096 {
		return AuthUser{}, false
	}
	decodedSignature, err := base64.RawURLEncoding.DecodeString(signature)
	if err != nil || !hmac.Equal(decodedSignature, sessions.sign([]byte(encoded))) {
		return AuthUser{}, false
	}
	payload, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return AuthUser{}, false
	}
	var claims sessionClaims
	if err := json.Unmarshal(payload, &claims); err != nil || sessions.now().Unix() >= claims.ExpiresAt {
		return AuthUser{}, false
	}
	return claims.User, true
}

func (sessions *sessionStore) sign(value []byte) []byte {
	mac := hmac.New(sha256.New, sessions.secret)
	_, _ = mac.Write(value)
	return mac.Sum(nil)
}

func (auth *authHandler) sessionUser(r *http.Request) (AuthUser, bool) {
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil || cookie.Value == "" {
		return AuthUser{}, false
	}
	return auth.sessions.get(cookie.Value)
}

func (auth *authHandler) resolveUser(ctx context.Context, identity AuthUser) (AuthUser, error) {
	user, err := auth.accountUser(ctx, identity.Email, identity.DisplayName)
	if err != nil {
		return AuthUser{}, err
	}
	user.EmailVerified = identity.EmailVerified
	user.AvatarURL = identity.AvatarURL
	return user, nil
}

func (auth *authHandler) requireRoles(w http.ResponseWriter, r *http.Request, roles ...string) (AuthUser, bool) {
	identity, ok := auth.sessionUser(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return AuthUser{}, false
	}
	if auth.queries == nil {
		writeError(w, http.StatusServiceUnavailable, "database is not configured")
		return AuthUser{}, false
	}
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()
	user, err := auth.resolveUser(ctx, identity)
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, "database unavailable")
		return AuthUser{}, false
	}
	for _, role := range roles {
		if user.Role == role {
			return user, true
		}
	}
	writeError(w, http.StatusForbidden, "forbidden")
	return AuthUser{}, false
}

func (auth *authHandler) requireAdmin(w http.ResponseWriter, r *http.Request) (AuthUser, bool) {
	return auth.requireRoles(w, r, "ADMIN")
}

func (auth *authHandler) requireEditor(w http.ResponseWriter, r *http.Request) (AuthUser, bool) {
	return auth.requireRoles(w, r, "ADMIN", "EDITOR")
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
	user.AvatarURL = strings.TrimSpace(claims.Picture)

	token, err := auth.sessions.create(user)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "session unavailable")
		return
	}
	auth.setSessionCookie(w, token)
	writeJSON(w, http.StatusOK, user)
}

func (auth *authHandler) me(w http.ResponseWriter, r *http.Request) {
	identity, ok := auth.sessionUser(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()
	user, err := auth.resolveUser(ctx, identity)
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, "database unavailable")
		return
	}
	writeJSON(w, http.StatusOK, user)
}

func (auth *authHandler) logout(w http.ResponseWriter, r *http.Request) {
	// ponytail: stateless logout clears this browser; a copied token expires within one hour.
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
	return auth.accountUserWithQueries(ctx, auth.queries, email, displayName)
}

func (auth *authHandler) accountUserWithQueries(ctx context.Context, queries *database.Queries, email, displayName string) (AuthUser, error) {
	if email == auth.adminEmail {
		user, err := queries.UpsertAdminUser(ctx, database.UpsertAdminUserParams{
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

	user, err := queries.GetUserByEmail(ctx, email)
	if err != nil {
		if err == pgx.ErrNoRows {
			return AuthUser{Email: email, DisplayName: displayName, Role: "VIEWER"}, nil
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
	return "VIEWER"
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
