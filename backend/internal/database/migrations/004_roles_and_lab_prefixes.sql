BEGIN;

ALTER TYPE app.user_role RENAME VALUE 'USER' TO 'EDITOR';
ALTER TABLE app.users ALTER COLUMN role SET DEFAULT 'EDITOR';

ALTER TABLE app.courses ALTER COLUMN code TYPE VARCHAR(32);
ALTER TABLE app.courses ALTER COLUMN name TYPE VARCHAR(110);
ALTER TABLE app.courses ALTER COLUMN abbreviation TYPE VARCHAR(20);
ALTER TABLE app.teachers ALTER COLUMN name TYPE VARCHAR(110);

UPDATE app.courses
SET name = CASE
      WHEN lower(name) LIKE 'lab - %' THEN name
      ELSE 'Lab - ' || name
    END,
    abbreviation = CASE
      WHEN abbreviation LIKE 'LAB-%' THEN abbreviation
      ELSE 'LAB-' || regexp_replace(abbreviation, '-L$', '')
    END,
    code = CASE
      WHEN code LIKE 'LAB-%' THEN code
      ELSE 'LAB-' || regexp_replace(code, '-L$', '')
    END
WHERE type = 'LABORATORY';

COMMIT;
