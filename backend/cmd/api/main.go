package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

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
		Addr:              ":" + port,
		Handler:           routes(db, os.Getenv("CRON_SECRET")),
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
		return nil, fmt.Errorf("parse DATABASE_URL: %w", err)
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

func routes(db *pgxpool.Pool, cronSecret string) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, "{\"ok\":true}\n")
	})

	mux.HandleFunc("GET /ready", func(w http.ResponseWriter, r *http.Request) {
		if db == nil {
			http.Error(w, "database not configured", http.StatusServiceUnavailable)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()
		if err := db.Ping(ctx); err != nil {
			http.Error(w, "database unavailable", http.StatusServiceUnavailable)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	})

	mux.HandleFunc("GET /internal/keepalive", func(w http.ResponseWriter, r *http.Request) {
		if cronSecret == "" || r.Header.Get("Authorization") != "Bearer "+cronSecret {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		if db == nil {
			http.Error(w, "database not configured", http.StatusServiceUnavailable)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()
		if _, err := db.Exec(ctx, "select 1"); err != nil {
			http.Error(w, "database unavailable", http.StatusServiceUnavailable)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	})

	return mux
}
