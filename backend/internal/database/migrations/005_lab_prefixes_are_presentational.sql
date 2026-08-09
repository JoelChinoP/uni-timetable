BEGIN;

UPDATE app.courses
SET name = regexp_replace(name, '^Lab[[:space:]]*-[[:space:]]*', '', 'i'),
    abbreviation = regexp_replace(regexp_replace(abbreviation, '^LAB-', '', 'i'), '-L$', '', 'i') || '-L',
    code = regexp_replace(regexp_replace(code, '^LAB-', '', 'i'), '-L$', '', 'i') || '-L'
WHERE type = 'LABORATORY';

COMMIT;
