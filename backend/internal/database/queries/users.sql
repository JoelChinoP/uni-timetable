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

-- name: ListUsersPage :many
SELECT id, email, display_name, role
FROM app.users
ORDER BY created_at DESC, id DESC
LIMIT sqlc.arg(page_size)::int OFFSET sqlc.arg(offset_rows)::int;

-- name: CountUsers :one
SELECT count(*) FROM app.users;

-- name: CreateUser :one
INSERT INTO app.users (email, display_name)
VALUES ($1, $2)
RETURNING id, email, display_name, role;

-- name: GetUserByID :one
SELECT id, email, display_name, role
FROM app.users
WHERE id = $1;

-- name: UpdateUser :one
UPDATE app.users
SET email = $2, display_name = $3, role = $4
WHERE id = $1
RETURNING id, email, display_name, role;

-- name: DeleteUser :execrows
DELETE FROM app.users WHERE id = $1;

-- name: CountAdmins :one
SELECT count(*) FROM app.users WHERE role = 'ADMIN';
