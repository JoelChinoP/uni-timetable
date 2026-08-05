# Uni Timetable

Planificador universitario con Svelte estático en la raíz y una API mínima en Go dentro de
`backend/`.

```text
Svelte + Vite (CDN global de Vercel)
                |
                | HTTPS
                v
Go net/http (Vercel gru1)
                |
                | Supavisor Transaction Pooler :6543
                v
Supabase PostgreSQL (sa-east-1)
```

La app tiene tres rutas principales: `/` muestra el tablero base, `/login` inicia sesión con Google
y `/panel` es exclusivo para administradores. Go valida el ID token de Google, crea una cookie
`HttpOnly` y consulta `app.users`; el primer correo de `ADMIN_EMAIL` se promueve a `ADMIN` al
conectarse. Los usuarios registrados manualmente tienen rol `USER`.

## Estructura

```text
.
├── src/
│   ├── App.svelte               # Rutas /, /login y /panel
│   └── lib/pages/               # Login y administración
├── public/                      # Archivos estáticos
├── scripts/dev.mjs              # Inicia Vite y Go
├── vercel.json                  # Proyecto web
└── backend/
    ├── cmd/api/                 # Servidor net/http detectado por Vercel
    ├── internal/database/
    │   ├── schema.sql           # Bootstrap SQL para una base vacía
    │   ├── queries/             # SQL fuente de SQLC
    │   ├── *.sql.go, db.go      # Código generado por SQLC
    │   └── migrations/          # Cambios SQL posteriores
    ├── sqlc.yaml
    ├── .env.example
    ├── go.mod
    └── vercel.json              # Proyecto API, región y cron
```

## Requisitos

- Node.js `^20.19.0 || >=22.12.0`
- npm
- Go `1.26.5`
- PostgreSQL opcional durante el desarrollo del frontend
- SQLC solo cuando cambien las queries; si no está instalado, se puede ejecutar con Docker

## Desarrollo

```bash
npm ci
npm run dev
```

`npm run dev` inicia:

| Servicio | URL                     |
| -------- | ----------------------- |
| Svelte   | `http://127.0.0.1:5173` |
| Go       | `http://127.0.0.1:8080` |

Crea `.env.local` en la raíz para el navegador y `backend/.env` para la API. El script carga el
archivo del backend antes de iniciar ambos procesos.

```env
VITE_GOOGLE_CLIENT_ID=CLIENT_ID.apps.googleusercontent.com
VITE_API_URL=
```

`VITE_API_URL` vacío hace que desarrollo use el mismo origen de Vite, mientras sus proxies
`/auth`, `/users` y `/api/*` se reenvían al backend. En producción sí debe contener la URL pública
de la API, por ejemplo `https://api.midominio.com`.

```env
DATABASE_URL=
GOOGLE_CLIENT_ID=CLIENT_ID.apps.googleusercontent.com
FRONTEND_ORIGINS=http://127.0.0.1:5173
ADMIN_EMAIL=jchinop@unsa.edu.pe
COOKIE_SECURE=false
```

`DATABASE_URL` puede quedar vacío para probar el frontend o login efímero: Go verifica Google y
determina el rol desde `ADMIN_EMAIL`, pero `/users` devolverá `503` hasta configurar PostgreSQL.
Para que la cookie sobreviva entre dominios distintos en producción, usa `COOKIE_SECURE=true`.

Vite redirige `/api/*` al backend local quitando el prefijo. Por ejemplo,
`http://127.0.0.1:5173/api/health` llega a `http://127.0.0.1:8080/health`.

Comandos útiles:

```bash
npm run build
npm run lint
npm run check
npm run format

# tras cambiar backend/internal/database/queries o schema.sql
docker run --rm -v "$PWD/backend:/src" -w /src sqlc/sqlc generate
```

## API Base

| Método | Ruta                  | Respuesta                                        |
| ------ | --------------------- | ------------------------------------------------ |
| `GET`  | `/health`             | Proceso disponible, sin tocar PostgreSQL         |
| `POST` | `/auth/login`         | Valida Google y crea cookie de sesión            |
| `GET`  | `/auth/me`            | Devuelve la sesión vigente o `401`               |
| `POST` | `/auth/logout`        | Borra cookie y sesión                            |
| `GET`  | `/users`              | Lista usuarios; requiere `ADMIN`                 |
| `POST` | `/users`              | Crea usuario `USER`; requiere `ADMIN`            |
| `GET`  | `/ready`              | `204` si PostgreSQL responde; `503` en otro caso |
| `GET`  | `/internal/keepalive` | Ejecuta `select 1` con autorización de cron      |

Las consultas de usuarios se declaran una vez en `internal/database/queries/users.sql`, se compilan
con SQLC `v1.31.1` a código Go tipado y se ejecutan sobre el pool existente, usando el modo `Exec`
requerido por Transaction Pooler. No hay ORM ni reflexión en los endpoints.

El pool se crea una vez por instancia, no abre una conexión durante el arranque y usa:

- `pgx/v5` `QueryExecModeExec`, compatible con Transaction Pooler sin prepared statements.
- `MaxConns=2`, `MinConns=0` y `MinIdleConns=0`.
- Timeouts de tres segundos para readiness y keepalive.
- `application_name=uni-timetable-api`.

## Supabase

1. Crea el proyecto en `South America (São Paulo)`, región `sa-east-1`.
2. Obtén `DATABASE_MIGRATION_URL` desde una conexión directa o Session Pooler `:5432`.
3. Para una base vacía, aplica el bootstrap completo que ya incluye `app.users`.

```bash
psql "$DATABASE_MIGRATION_URL" -v ON_ERROR_STOP=1 -f backend/internal/database/schema.sql
```

Para una base creada antes de usuarios, aplica la migración incremental:

```bash
psql "$DATABASE_MIGRATION_URL" -v ON_ERROR_STOP=1 -f backend/internal/database/migrations/001_create_users.sql
```

`ADMIN_EMAIL` no se inserta como SQL estático: al autenticar ese correo con Google, la API hace un
`upsert` en `app.users` con rol `ADMIN`. Define el correo en `backend/.env` y en Vercel API.

4. Crea un rol de runtime con una contraseña aleatoria larga. No guardes esa contraseña en el
   repositorio.

```sql
create role app_runtime
with
  login
  password 'REEMPLAZAR_CON_UN_SECRETO_LARGO'
  nosuperuser
  nocreatedb
  nocreaterole
  noinherit;

alter role app_runtime set search_path = app, public;

grant connect on database postgres to app_runtime;
grant usage on schema app to app_runtime;
grant usage on type app.mode_type, app.week_day, app.user_role to app_runtime;
grant select, insert, update, delete on all tables in schema app to app_runtime;
grant usage, select on all sequences in schema app to app_runtime;

alter default privileges for role postgres in schema app
grant select, insert, update, delete on tables to app_runtime;

alter default privileges for role postgres in schema app
grant usage, select on sequences to app_runtime;

alter default privileges for role postgres in schema app
grant usage on types to app_runtime;
```

Si las migraciones futuras crean objetos con un rol distinto de `postgres`, reemplaza ese nombre en
`alter default privileges` por el rol que realmente sea propietario.

5. En **Connect -> Transaction pooler**, copia la cadena exacta de puerto `6543`, cambia el usuario
   por `app_runtime.PROJECT_REF` y úsala como `DATABASE_URL`. Conserva `sslmode=require` y
   percent-encodea caracteres especiales de la contraseña.

No construyas el hostname manualmente. `DATABASE_MIGRATION_URL` no debe usarse en la Function y la
API nunca ejecuta `schema.sql` al arrancar. Si no usarás la Data API de Supabase, puedes deshabilitarla
y mantener el acceso únicamente desde el backend.

## Vercel

Crea dos proyectos desde este mismo repositorio:

| Proyecto | Root Directory       | Preset | Dominio sugerido    |
| -------- | -------------------- | ------ | ------------------- |
| Web      | raíz del repositorio | Vite   | `app.midominio.com` |
| API      | `backend`            | Go     | `api.midominio.com` |

La raíz contiene `vercel.json` para el build estático en `dist`. `backend/vercel.json` fija
`framework: go`, ejecuta la Function en `gru1` y registra el cron diario.

Configura en el proyecto web:

| Variable                | Uso                               |
| ----------------------- | --------------------------------- |
| `VITE_GOOGLE_CLIENT_ID` | Client ID público OAuth de Google |
| `VITE_API_URL`          | URL pública del proyecto API      |

Y en el proyecto API, para **Production**:

| Variable           | Uso                                           |
| ------------------ | --------------------------------------------- |
| `DATABASE_URL`     | Transaction Pooler de Supabase, puerto `6543` |
| `GOOGLE_CLIENT_ID` | Mismo Client ID usado por el navegador        |
| `FRONTEND_ORIGINS` | Origen exacto del frontend, sin barra final   |
| `ADMIN_EMAIL`      | Correo promovido a `ADMIN` en su primer login |
| `COOKIE_SECURE`    | `true` en HTTPS para `SameSite=None; Secure`  |
| `CRON_SECRET`      | Secreto aleatorio de al menos 16 caracteres   |

Vercel define `PORT` automáticamente. Para desarrollo local, su valor predeterminado es `8080`.

El cron ejecuta `/internal/keepalive` a las `11:00 UTC`, aproximadamente `06:00` de Lima. Vercel
Hobby solo admite una ejecución diaria y puede dispararla en cualquier momento de esa hora. Vercel
envía `CRON_SECRET` como `Authorization: Bearer <CRON_SECRET>`.

El keepalive genera actividad mínima, pero no mantiene caliente la instancia Go ni garantiza que un
proyecto Supabase Free no sea pausado. Supabase solo garantiza ausencia de pausas por inactividad en
planes de pago.

## Documentación Vigente

Configuración contrastada el 4 de agosto de 2026 con fuentes oficiales:

- [Vercel Go Runtime](https://vercel.com/docs/functions/runtimes/go)
- [Vercel monorepos](https://vercel.com/docs/monorepos)
- [Vercel project configuration](https://vercel.com/docs/project-configuration/vercel-json)
- [Vercel Cron Jobs](https://vercel.com/docs/cron-jobs/manage-cron-jobs)
- [Vite static deployment](https://vite.dev/guide/static-deploy.html)
- [Supabase database connections](https://supabase.com/docs/guides/database/connecting-to-postgres)
- [Supabase regions](https://supabase.com/docs/guides/platform/regions)
- [Supabase project pausing](https://supabase.com/docs/guides/platform/free-project-pausing)
- [Google ID token verification](https://developers.google.com/identity/gsi/web/guides/verify-google-id-token)
- [coreos/go-oidc](https://pkg.go.dev/github.com/coreos/go-oidc/v3/oidc)
- [sqlc Go + pgx](https://docs.sqlc.dev/en/latest/guides/using-go-and-pgx.html)
- [pgxpool](https://pkg.go.dev/github.com/jackc/pgx/v5/pgxpool)
- [PostgreSQL schemas and privileges](https://www.postgresql.org/docs/current/ddl-schemas.html)
