-- Cambios del tablero 2026-B: corrección de la hora académica 2 según la fuente
-- oficial (07:50-08:40), aulas con códigos largos ("Laboratorio 01"), sesiones
-- de hasta 6 bloques, créditos opcionales y enlaces para compartir horarios.
BEGIN;

UPDATE app.academic_hours SET end_time = '08:40' WHERE hour_number = 2;

ALTER TABLE app.classrooms ALTER COLUMN code TYPE VARCHAR(32);

ALTER TABLE app.courses ALTER COLUMN credits DROP NOT NULL;

ALTER TABLE app.schedule DROP CONSTRAINT schedule_duration_hours_check;
ALTER TABLE app.schedule
  ADD CONSTRAINT schedule_duration_hours_check CHECK (duration_hours BETWEEN 1 AND 6);

-- ponytail: short random ids make share links unguessable without a lookup; selection is a validated {courseId: groupId} map.
CREATE TABLE app.shared_timetables (
  id CHAR(10) PRIMARY KEY,
  selection JSONB NOT NULL,
  created_by INTEGER REFERENCES app.users(id) ON DELETE SET NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

  CHECK (jsonb_typeof(selection) = 'object')
);

COMMIT;
