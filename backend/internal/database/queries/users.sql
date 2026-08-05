-- name: GetUserByEmail :one
SELECT id, email, display_name, role
FROM app.users
WHERE email = $1;

-- name: UpsertAdminUser :one
INSERT INTO app.users (email, display_name, role)
VALUES ($1, $2, 'ADMIN')
ON CONFLICT (email) DO UPDATE SET
  display_name = EXCLUDED.display_name,
  role = 'ADMIN'
RETURNING id, email, display_name, role;

-- name: ListUsers :many
SELECT id, email, display_name, role
FROM app.users
ORDER BY created_at DESC, id DESC;

-- name: CreateUser :one
INSERT INTO app.users (email, display_name)
VALUES ($1, $2)
RETURNING id, email, display_name, role;
