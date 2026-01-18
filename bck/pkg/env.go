package pkg

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/joho/godotenv"

	"github.com/JoelChinoP/timetable_bck/internal/database"
)

type AppConfig struct {
	Env         string
	AppName     string
	Port        string
	JWTSecret   string
	CORSOrigins string

	Database database.Config
}

var (
	cfgOnce sync.Once
	cfg     AppConfig
	cfgErr  error
)

// LoadConfig carga .env (si existe), lee variables de entorno, aplica defaults y valida.
// Se cachea el resultado (singleton) para uso eficiente en todo el proceso.
func LoadConfig() (AppConfig, error) {
	cfgOnce.Do(func() {
		// No sobreescribe variables ya presentes en el entorno del proceso.
		_ = godotenv.Load()

		env := getEnv("GO_ENV", "development")

		port := getEnv("PORT", "8080")
		if strings.TrimSpace(port) == "" {
			cfgErr = fmt.Errorf("config: PORT no puede estar vacío")
			return
		}

		jwtSecret := strings.TrimSpace(os.Getenv("JWT_SECRET"))
		if jwtSecret == "" {
			cfgErr = fmt.Errorf("config: JWT_SECRET es requerido")
			return
		}

		dbMaxConns, err := getEnvInt32("DB_MAX_CONNS", 6)
		if err != nil {
			cfgErr = err
			return
		}
		dbMinConns, err := getEnvInt32("DB_MIN_CONNS", 1)
		if err != nil {
			cfgErr = err
			return
		}
		dbMaxConnLifetime, err := getEnvSeconds("DB_MAX_CONN_LIFETIME", 3600)
		if err != nil {
			cfgErr = err
			return
		}
		dbMaxConnIdleTime, err := getEnvSeconds("DB_MAX_CONN_IDLE_TIME", 300)
		if err != nil {
			cfgErr = err
			return
		}

		dbHost := strings.TrimSpace(os.Getenv("DB_HOST"))
		dbPort := strings.TrimSpace(os.Getenv("DB_PORT"))
		dbUser := strings.TrimSpace(os.Getenv("DB_USER"))
		dbName := strings.TrimSpace(os.Getenv("DB_NAME"))
		// DB_PASSWORD puede ser vacío en algunos entornos.
		dbPassword := os.Getenv("DB_PASSWORD")

		missing := make([]string, 0, 4)
		if dbHost == "" {
			missing = append(missing, "DB_HOST")
		}
		if dbPort == "" {
			missing = append(missing, "DB_PORT")
		}
		if dbUser == "" {
			missing = append(missing, "DB_USER")
		}
		if dbName == "" {
			missing = append(missing, "DB_NAME")
		}
		if len(missing) > 0 {
			cfgErr = fmt.Errorf("config: variables requeridas faltantes: %s", strings.Join(missing, ", "))
			return
		}

		cfg = AppConfig{
			Env:         env,
			AppName:     getEnv("APP_NAME", "Timetable Api"),
			Port:        port,
			JWTSecret:   jwtSecret,
			CORSOrigins: getEnv("CORS_ORIGINS", "*"),
			Database: database.Config{
				Host:            dbHost,
				Port:            dbPort,
				User:            dbUser,
				Password:        dbPassword,
				Database:        dbName,
				MaxConns:        dbMaxConns,
				MinConns:        dbMinConns,
				MaxConnLifetime: dbMaxConnLifetime,
				MaxConnIdleTime: dbMaxConnIdleTime,
			},
		}
	})

	return cfg, cfgErr
}

func MustLoadConfig() AppConfig {
	c, err := LoadConfig()
	if err != nil {
		panic(err)
	}
	return c
}

func getEnv(key, def string) string {
	val := strings.TrimSpace(os.Getenv(key))
	if val == "" {
		return def
	}
	return val
}

func getEnvInt32(key string, def int32) (int32, error) {
	val := strings.TrimSpace(os.Getenv(key))
	if val == "" {
		return def, nil
	}

	i, err := strconv.ParseInt(val, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("config: %s inválido (%q): %w", key, val, err)
	}
	return int32(i), nil
}

func getEnvSeconds(key string, defSeconds int64) (time.Duration, error) {
	val := strings.TrimSpace(os.Getenv(key))
	if val == "" {
		return time.Duration(defSeconds) * time.Second, nil
	}

	i, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("config: %s inválido (%q): %w", key, val, err)
	}
	if i < 0 {
		return 0, fmt.Errorf("config: %s no puede ser negativo (%q)", key, val)
	}
	return time.Duration(i) * time.Second, nil
}
