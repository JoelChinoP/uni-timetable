package main

import (
	"context"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JoelChinoP/uni-timetable/backend/internal/database"
)

type routeConfig struct {
	db             *pgxpool.Pool
	cronSecret     string
	googleVerifier *oidc.IDTokenVerifier
	adminEmail     string
	allowedOrigins []string
	secureCookies  bool
	sessions       *sessionStore
}

func main() {
	db, err := newPool(os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	// ponytail: the pool lives for the process; Vercel discards it with the instance.

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr: ":" + port,
		Handler: routes(routeConfig{
			db:             db,
			cronSecret:     os.Getenv("CRON_SECRET"),
			googleVerifier: newGoogleVerifier(os.Getenv("GOOGLE_CLIENT_ID")),
			adminEmail:     normalizeEmail(os.Getenv("ADMIN_EMAIL")),
			allowedOrigins: originsFromEnv(os.Getenv("FRONTEND_ORIGINS")),
			secureCookies:  os.Getenv("COOKIE_SECURE") == "true",
		}),
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("API listening on :%s", port)
	log.Fatal(server.ListenAndServe())
}

func newPool(databaseURL string) (*pgxpool.Pool, error) {
	if strings.TrimSpace(databaseURL) == "" {
		return nil, nil
	}

	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		// ponytail: do not echo DATABASE_URL into logs; Supabase passwords with symbols need percent-encoding.
		return nil, errors.New("DATABASE_URL must be a valid PostgreSQL URL")
	}

	config.MaxConns = 2
	config.MinConns = 0
	config.MinIdleConns = 0
	config.MaxConnIdleTime = 2 * time.Minute
	config.MaxConnLifetime = 30 * time.Minute
	config.MaxConnLifetimeJitter = 5 * time.Minute
	config.HealthCheckPeriod = time.Minute
	config.PingTimeout = 3 * time.Second
	config.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeExec
	config.ConnConfig.RuntimeParams["application_name"] = "uni-timetable-api"

	return pgxpool.NewWithConfig(context.Background(), config)
}

func newGoogleVerifier(clientID string) *oidc.IDTokenVerifier {
	if strings.TrimSpace(clientID) == "" {
		return nil
	}
	ctx := oidc.ClientContext(context.Background(), &http.Client{Timeout: 5 * time.Second})
	keySet := oidc.NewRemoteKeySet(ctx, "https://www.googleapis.com/oauth2/v3/certs")
	return oidc.NewVerifier("https://accounts.google.com", keySet, &oidc.Config{ClientID: clientID})
}

func originsFromEnv(value string) []string {
	origins := make([]string, 0, 2)
	for origin := range strings.SplitSeq(value, ",") {
		if origin = strings.TrimSpace(origin); origin != "" {
			origins = append(origins, origin)
		}
	}
	if len(origins) == 0 {
		return []string{"http://127.0.0.1:5173"}
	}
	return origins
}

func routes(config routeConfig) http.Handler {
	var queries *database.Queries
	if config.db != nil {
		queries = database.New(config.db)
	}
	sessions := config.sessions
	if sessions == nil {
		sessions = newSessionStore()
	}
	auth := &authHandler{
		queries:    queries,
		verifier:   config.googleVerifier,
		sessions:   sessions,
		adminEmail: config.adminEmail,
		secure:     config.secureCookies,
	}
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, "{\"ok\":true}\n")
	})

	mux.HandleFunc("POST /auth/login", auth.login)
	mux.HandleFunc("GET /auth/me", auth.me)
	mux.HandleFunc("POST /auth/logout", auth.logout)
	mux.HandleFunc("/users", auth.users)

	termLabel := os.Getenv("TERM_LABEL")
	if termLabel == "" {
		termLabel = "2026-B"
	}
	planner := &plannerHandler{queries: queries, termLabel: termLabel}
	catalog := &catalogHandler{queries: queries, db: config.db, auth: auth}
	shared := &sharedHandler{queries: queries, auth: auth}

	mux.HandleFunc("GET /planner/dashboard", planner.dashboard)
	mux.HandleFunc("/shared", shared.shared)
	mux.HandleFunc("GET /shared/{id}", shared.getShared)
	mux.HandleFunc("/classrooms", catalog.classrooms)
	mux.HandleFunc("DELETE /classrooms/{id}", catalog.deleteClassroom)
	mux.HandleFunc("/teachers", catalog.teachers)
	mux.HandleFunc("DELETE /teachers/{id}", catalog.deleteTeacher)
	mux.HandleFunc("POST /courses", catalog.courses)
	mux.HandleFunc("DELETE /courses/{id}", catalog.deleteCourse)
	mux.HandleFunc("POST /groups", catalog.groups)
	mux.HandleFunc("DELETE /groups/{id}", catalog.deleteGroup)

	mux.HandleFunc("GET /ready", func(w http.ResponseWriter, r *http.Request) {
		if config.db == nil {
			http.Error(w, "database not configured", http.StatusServiceUnavailable)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()
		if err := config.db.Ping(ctx); err != nil {
			http.Error(w, "database unavailable", http.StatusServiceUnavailable)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	})

	mux.HandleFunc("GET /internal/keepalive", func(w http.ResponseWriter, r *http.Request) {
		if config.cronSecret == "" || r.Header.Get("Authorization") != "Bearer "+config.cronSecret {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		if config.db == nil {
			http.Error(w, "database not configured", http.StatusServiceUnavailable)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()
		if _, err := config.db.Exec(ctx, "select 1"); err != nil {
			http.Error(w, "database unavailable", http.StatusServiceUnavailable)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	})

	return withCORS(mux, config.allowedOrigins)
}

func withCORS(next http.Handler, origins []string) http.Handler {
	allowed := make(map[string]struct{}, len(origins))
	for _, origin := range origins {
		allowed[origin] = struct{}{}
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin == "" {
			next.ServeHTTP(w, r)
			return
		}

		if _, ok := allowed[origin]; !ok {
			writeError(w, http.StatusForbidden, "origin not allowed")
			return
		}

		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		w.Header().Add("Vary", "Origin")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
