# 📅 Timetable Backend

Backend para la gestión de horarios universitarios desarrollado con **Go**, **Fiber** y **PostgreSQL**.

## ✅ Requisitos

- **Go** (según `go.mod`: 1.25.4)
- **PostgreSQL** (si vas a habilitar conexión a DB)
- **sqlc** (solo si vas a regenerar código en desarrollo)

## 🚀 Tecnologías

- **Go** - Lenguaje de programación
- **Fiber** - Framework HTTP
- **PostgreSQL** - Base de datos
- **SQLC** - Generador de código SQL type-safe
- **pgx/v5** - Driver/Pool para PostgreSQL
- **godotenv** - Carga de variables desde `.env`

## 📁 Estructura del Proyecto

```
├── cmd/                         # Entrypoint (servidor HTTP)
│   ├── main.go                  # (demo / comentado)
│   └── server.go                # main real
├── internal/
│   ├── api/                     # Rutas HTTP
│   ├── auth/                    # Auth (local/jwt/oauth)
│   └── database/                # DB + SQLC
│       ├── schema.sql           # Schema PostgreSQL
│       ├── queries/             # Queries para SQLC
│       └── sqlc/                # Código generado por SQLC
└── pkg/                         # Config, middlewares y helpers
```

## ⚙️ Instalación

```bash
# Clonar el repositorio
git clone https://github.com/JoelChinoP/timetable_bck.git
cd timetable_bck

# Instalar dependencias
go mod download

# Ejecutar la aplicación
go run ./cmd
```

## 🔐 Configuración (.env)

Este proyecto carga variables desde un archivo `.env` (si existe) y/o desde el entorno.

Variables requeridas por la configuración actual:

```env
# App
GO_ENV=development
APP_NAME=Timetable Api
PORT=8080
CORS_ORIGINS=*

# Auth
JWT_SECRET=change_me

# DB (requeridas por la carga de config)
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=timetable

# Pool (opcionales)
DB_MAX_CONNS=6
DB_MIN_CONNS=1
DB_MAX_CONN_LIFETIME=3600
DB_MAX_CONN_IDLE_TIME=300
```

Nota: la inicialización de base de datos en el servidor está comentada actualmente en `cmd/server.go`. Si la habilitas, se usará la configuración `DB_*` y se verificará/construirá base de datos (seed) para `academic_hours`.

## 🔧 Desarrollo

```bash
# Generar código SQLC
sqlc generate
```

Si no tienes `sqlc` instalado:

```bash
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
```

## 📡 Endpoints

| Método | Ruta      | Descripción    |
| ------ | --------- | -------------- |
| GET    | `/`       | Health simple  |
| GET    | `/status` | Estado/versión |

Las rutas de autenticación existen en `internal/auth` (por ejemplo, `POST /auth/login`), pero su registro está comentado en `internal/api/routes.go` en este momento.
