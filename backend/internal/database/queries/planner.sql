-- name: ListAcademicHours :many
SELECT hour_number, start_time, end_time
FROM app.academic_hours
ORDER BY hour_number;

-- name: ListPlannerCourses :many
SELECT c.id, c.code, c.name, c.abbreviation, c.summary, c.credits, c.color, c.type,
       c.academic_year, c.id_course_theory, t.id AS teacher_id,
       CAST(COALESCE(t.name || ' ' || t.last_name, '') AS text) AS teacher_name
FROM app.courses c
JOIN app.careers cr ON cr.id = c.id_career
LEFT JOIN app.teachers t ON t.id = c.id_teacher
WHERE cr.code = $1
ORDER BY c.academic_year, c.name, c.type;

-- name: ListPlannerGroups :many
SELECT g.id, g.id_course, g.name, cl.code AS classroom_code
FROM app.groups g
JOIN app.courses c ON c.id = g.id_course
JOIN app.careers cr ON cr.id = c.id_career
LEFT JOIN app.classrooms cl ON cl.id = g.id_classroom
WHERE cr.code = $1
ORDER BY g.name;

-- name: ListPlannerSessions :many
SELECT s.id, s.id_group, g.id_course, s.day, s.start_hour_academic, s.duration_hours,
       c.type, c.name AS course_name, cl.code AS classroom_code
FROM app.schedule s
JOIN app.groups g ON g.id = s.id_group
JOIN app.courses c ON c.id = g.id_course
JOIN app.careers cr ON cr.id = c.id_career
LEFT JOIN app.classrooms cl ON cl.id = g.id_classroom
WHERE cr.code = $1
ORDER BY s.day, s.start_hour_academic;
