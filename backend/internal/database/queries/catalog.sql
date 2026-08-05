-- name: ListClassrooms :many
SELECT id, code, type, floor, capacity
FROM app.classrooms
ORDER BY code;

-- name: CreateClassroom :one
INSERT INTO app.classrooms (code, type, floor, capacity)
VALUES ($1, $2, $3, $4)
RETURNING id, code, type, floor, capacity;

-- name: DeleteClassroom :execrows
DELETE FROM app.classrooms WHERE id = $1;

-- name: ListTeachers :many
SELECT t.id, t.name, t.last_name
FROM app.teachers t
JOIN app.career_teachers ct ON ct.id_teacher = t.id
JOIN app.careers cr ON cr.id = ct.id_career
WHERE cr.code = $1
ORDER BY t.last_name, t.name;

-- name: UpsertTeacher :one
INSERT INTO app.teachers (name, last_name)
VALUES ($1, $2)
ON CONFLICT (name, last_name) DO UPDATE SET name = EXCLUDED.name
RETURNING id;

-- name: LinkTeacherToCareer :exec
INSERT INTO app.career_teachers (id_career, id_teacher)
SELECT cr.id, $2
FROM app.careers cr
WHERE cr.code = $1
ON CONFLICT DO NOTHING;

-- name: DeleteTeacher :execrows
DELETE FROM app.teachers WHERE id = $1;

-- name: CreateCourse :one
INSERT INTO app.courses
  (code, name, abbreviation, credits, color, type, id_career, id_course_theory, academic_year, id_teacher)
VALUES ($1, $2, $3, $4, $5, $6, (SELECT cr.id FROM app.careers cr WHERE cr.code = $7), $8, $9, $10)
RETURNING id, code, name;

-- name: DeleteCourse :execrows
DELETE FROM app.courses WHERE id = $1;

-- name: GetCourseMeta :one
SELECT c.type, cr.code AS career_code
FROM app.courses c
JOIN app.careers cr ON cr.id = c.id_career
WHERE c.id = $1;

-- name: GetClassroomType :one
SELECT type FROM app.classrooms WHERE id = $1;

-- name: CreateGroup :one
INSERT INTO app.groups (code, name, id_course, id_classroom)
VALUES ($1, $2, $3, $4)
RETURNING id, code, name;

-- name: CreateGroupSession :exec
INSERT INTO app.schedule (id_group, day, start_hour_academic, duration_hours)
VALUES ($1, $2, $3, $4);

-- name: DeleteGroup :execrows
DELETE FROM app.groups WHERE id = $1;

-- name: CreateSharedTimetable :exec
INSERT INTO app.shared_timetables (id, selection, created_by)
VALUES ($1, $2::text::jsonb, $3);

-- name: GetSharedTimetable :one
SELECT selection FROM app.shared_timetables WHERE id = $1;
