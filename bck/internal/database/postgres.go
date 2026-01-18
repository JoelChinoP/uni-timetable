package database

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	sqlc "github.com/JoelChinoP/timetable_bck/internal/database/sqlc"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	pool *pgxpool.Pool
	once sync.Once
)

type Config struct {
	Host, Port, User, Password, Database string
	MaxConns, MinConns                   int32
	MaxConnLifetime, MaxConnIdleTime     time.Duration
}

func InitDB(ctx context.Context, cfg Config) {
	once.Do(func() {
		dsn := fmt.Sprintf(
			"postgres://%s:%s@%s:%s/%s?sslmode=disable",
			cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Database,
		)

		poolCfg, err := pgxpool.ParseConfig(dsn)
		if err != nil {
			log.Fatalf("db: unable to parse config: %v", err)
		}

		poolCfg.MaxConns = cfg.MaxConns
		poolCfg.MinConns = cfg.MinConns
		poolCfg.MaxConnLifetime = cfg.MaxConnLifetime
		poolCfg.MaxConnIdleTime = cfg.MaxConnIdleTime

		p, err := pgxpool.NewWithConfig(ctx, poolCfg)
		if err != nil {
			log.Fatalf("db: unable to create pool: %v", err)
		}

		// Verifica conectividad real
		verifyConnection(ctx, p)

		pool = p
		log.Printf("db: pool initialized (db=%s maxConns=%d)", cfg.Database, cfg.MaxConns)
	})
}

func Pool() *pgxpool.Pool {
	if pool == nil {
		log.Fatal("db: Pool called before InitDB")
	}
	return pool
}

func Close() {
	if pool != nil {
		pool.Close()
		log.Println("db: pool closed")
	}
}

func verifyConnection(ctx context.Context, p *pgxpool.Pool) {
	queries := sqlc.New(p)
	count, err := queries.CountAcademicHours(ctx)
	if err != nil {
		p.Close()
		log.Fatalf("db: unable to ping database: %v", err)
	}

	if count == 0 {
		if seedErr := queries.SeedAcademicHours(ctx); seedErr != nil {
			p.Close()
			log.Fatalf("db: unable to seed academic_hours table: %v", seedErr)
		}
	}
}
