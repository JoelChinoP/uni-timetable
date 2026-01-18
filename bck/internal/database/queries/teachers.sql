-- name: GetTeacher :one
SELECT * FROM teachers
WHERE id = $1 LIMIT 1;

-- name: ListTeachers :many
SELECT * FROM teachers
ORDER BY name;

-- name: CreateTeacher :one
INSERT INTO teachers (
  name,
  last_name
) VALUES (
  $1, $2
) RETURNING *;

-- name: UpdateTeacher :one
UPDATE teachers
SET name = $2,
    last_name = $3
WHERE id = $1
RETURNING *;

-- name: DeleteTeacher :exec
DELETE FROM teachers
WHERE id = $1;
