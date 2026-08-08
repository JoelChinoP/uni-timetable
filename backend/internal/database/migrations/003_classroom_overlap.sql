BEGIN;

CREATE OR REPLACE FUNCTION app.reject_classroom_overlap() RETURNS trigger
LANGUAGE plpgsql AS $$
DECLARE
  classroom_id INTEGER;
BEGIN
  SELECT id_classroom INTO classroom_id FROM app.groups WHERE id = NEW.id_group;
  IF classroom_id IS NULL THEN
    RETURN NEW;
  END IF;

  PERFORM pg_advisory_xact_lock(classroom_id);
  IF EXISTS (
    SELECT 1
    FROM app.schedule existing
    JOIN app.groups existing_group ON existing_group.id = existing.id_group
    WHERE existing_group.id_classroom = classroom_id
      AND existing.day = NEW.day
      AND existing.id <> NEW.id
      AND existing.start_hour_academic < NEW.start_hour_academic + NEW.duration_hours
      AND NEW.start_hour_academic < existing.start_hour_academic + existing.duration_hours
  ) THEN
    RAISE EXCEPTION 'classroom schedule overlaps' USING ERRCODE = '23P01';
  END IF;
  RETURN NEW;
END;
$$;

DROP TRIGGER IF EXISTS schedule_classroom_overlap ON app.schedule;
CREATE TRIGGER schedule_classroom_overlap
BEFORE INSERT OR UPDATE ON app.schedule
FOR EACH ROW EXECUTE FUNCTION app.reject_classroom_overlap();

COMMIT;
