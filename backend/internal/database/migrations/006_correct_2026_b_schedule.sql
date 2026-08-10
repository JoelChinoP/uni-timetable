-- Corrige las ocho sesiones de 2026-B que cambiaron en la extracción integral.
-- El orden evita solapamientos transitorios con el trigger de aulas.
BEGIN;

-- Aula 305: desplaza IHC-C antes de ampliar PPA-B.
UPDATE app.schedule AS s
SET start_hour_academic = 12,
    duration_hours = 3
FROM app.groups AS g
JOIN app.courses AS c ON c.id = g.id_course
JOIN app.classrooms AS cl ON cl.id = g.id_classroom
WHERE s.id_group = g.id
  AND c.code = 'IHC'
  AND c.type = 'THEORY'
  AND g.name = 'C'
  AND cl.code = 'Aula 305'
  AND s.day = 'WEDNESDAY'
  AND s.start_hour_academic = 11
  AND s.duration_hours = 3;

UPDATE app.schedule AS s
SET start_hour_academic = 9,
    duration_hours = 3
FROM app.groups AS g
JOIN app.courses AS c ON c.id = g.id_course
JOIN app.classrooms AS cl ON cl.id = g.id_classroom
WHERE s.id_group = g.id
  AND c.code = 'PPA'
  AND c.type = 'THEORY'
  AND g.name = 'B'
  AND cl.code = 'Aula 305'
  AND s.day = 'WEDNESDAY'
  AND s.start_hour_academic = 9
  AND s.duration_hours = 2;

-- Aula 201: elimina los tres bloques incorrectos del inicio de CS-A.
UPDATE app.schedule AS s
SET start_hour_academic = 6,
    duration_hours = 2
FROM app.groups AS g
JOIN app.courses AS c ON c.id = g.id_course
JOIN app.classrooms AS cl ON cl.id = g.id_classroom
WHERE s.id_group = g.id
  AND c.code = 'CS-CONST'
  AND c.name = 'Construcción de Software'
  AND c.type = 'THEORY'
  AND g.name = 'A'
  AND cl.code = 'Aula 201'
  AND s.day = 'TUESDAY'
  AND s.start_hour_academic = 3
  AND s.duration_hours = 5;

-- Aula 202: desplaza AS-C y GPS-B antes de ampliar NE-B.
UPDATE app.schedule AS s
SET start_hour_academic = 14,
    duration_hours = 2
FROM app.groups AS g
JOIN app.courses AS c ON c.id = g.id_course
JOIN app.classrooms AS cl ON cl.id = g.id_classroom
WHERE s.id_group = g.id
  AND c.code = 'AS'
  AND c.type = 'THEORY'
  AND g.name = 'C'
  AND cl.code = 'Aula 202'
  AND s.day = 'MONDAY'
  AND s.start_hour_academic = 13
  AND s.duration_hours = 3;

UPDATE app.schedule AS s
SET start_hour_academic = 12,
    duration_hours = 2
FROM app.groups AS g
JOIN app.courses AS c ON c.id = g.id_course
JOIN app.classrooms AS cl ON cl.id = g.id_classroom
WHERE s.id_group = g.id
  AND c.code = 'GPS'
  AND c.type = 'THEORY'
  AND g.name = 'B'
  AND cl.code = 'Aula 202'
  AND s.day = 'MONDAY'
  AND s.start_hour_academic = 11
  AND s.duration_hours = 2;

UPDATE app.schedule AS s
SET start_hour_academic = 9,
    duration_hours = 3
FROM app.groups AS g
JOIN app.courses AS c ON c.id = g.id_course
JOIN app.classrooms AS cl ON cl.id = g.id_classroom
WHERE s.id_group = g.id
  AND c.code = 'NE'
  AND c.type = 'THEORY'
  AND g.name = 'B'
  AND cl.code = 'Aula 202'
  AND s.day = 'MONDAY'
  AND s.start_hour_academic = 9
  AND s.duration_hours = 2;

-- Aula 205: corrige GSTI-A y reduce PPA-C.
UPDATE app.schedule AS s
SET start_hour_academic = 1,
    duration_hours = 2
FROM app.groups AS g
JOIN app.courses AS c ON c.id = g.id_course
JOIN app.classrooms AS cl ON cl.id = g.id_classroom
WHERE s.id_group = g.id
  AND c.code = 'GSTI'
  AND c.type = 'THEORY'
  AND g.name = 'A'
  AND cl.code = 'Aula 205'
  AND s.day = 'TUESDAY'
  AND s.start_hour_academic = 1
  AND s.duration_hours = 3;

UPDATE app.schedule AS s
SET start_hour_academic = 4,
    duration_hours = 3
FROM app.groups AS g
JOIN app.courses AS c ON c.id = g.id_course
JOIN app.classrooms AS cl ON cl.id = g.id_classroom
WHERE s.id_group = g.id
  AND c.code = 'PPA'
  AND c.type = 'THEORY'
  AND g.name = 'C'
  AND cl.code = 'Aula 205'
  AND s.day = 'TUESDAY'
  AND s.start_hour_academic = 4
  AND s.duration_hours = 5;

COMMIT;
