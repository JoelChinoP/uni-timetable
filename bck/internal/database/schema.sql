-- ============================================
-- ENUMS
-- ============================================
CREATE TYPE mode_type AS ENUM ('THEORY', 'LABORATORY');
CREATE TYPE week_day AS ENUM ('MONDAY','TUESDAY','WEDNESDAY','THURSDAY','FRIDAY','SATURDAY');

-- ============================================
-- HORAS ACADÉMICAS (lookup table)
-- ============================================
CREATE TABLE academic_hours (
  hour_number SMALLINT PRIMARY KEY, -- 1 a 15 horas académicas
  start_time  TIME NOT NULL,
  end_time    TIME NOT NULL,

  CHECK (hour_number BETWEEN 1 AND 15),
  CHECK (end_time > start_time)
);
-- Execute
INSERT INTO academic_hours (hour_number, start_time, end_time) VALUES
(1,  '07:00', '07:50'),
(2,  '07:50', '08:50'),
(3,  '08:50', '09:40'),
(4,  '09:40', '10:30'),
(5,  '10:40', '11:30'),
(6,  '11:30', '12:20'),
(7,  '12:20', '13:10'),
(8,  '13:10', '14:00'),
(9,  '14:00', '14:50'),
(10, '14:50', '15:40'),
(11, '15:50', '16:40'),
(12, '16:40', '17:30'),
(13, '17:40', '18:30'),
(14, '18:30', '19:20'),
(15, '19:20', '20:10');

-- ============================================
-- DOCENTES
-- ============================================
CREATE TABLE teachers (
  id         SERIAL PRIMARY KEY,
  name       VARCHAR(100) NOT NULL,
  last_name  VARCHAR(100) NOT NULL,
  created_at TIMESTAMPTZ DEFAULT NOW(),

  CONSTRAINT unique_teacher_name_last_name UNIQUE (name, last_name)
);

-- ============================================
-- SALONES / AULAS
-- ============================================
CREATE TABLE classrooms (
  id         SERIAL PRIMARY KEY,
  code       VARCHAR(10) NOT NULL, -- Ejemplo: "101"
  type       mode_type NOT NULL,
  floor      SMALLINT, -- piso
  capacity   SMALLINT,
  created_at TIMESTAMPTZ DEFAULT NOW(),

  CONSTRAINT unique_classroom_code_type UNIQUE (code, type)
);

-- ============================================
-- CURSOS
-- ============================================
CREATE TABLE courses (
  id                SERIAL PRIMARY KEY,
  code              VARCHAR(24) NOT NULL UNIQUE, -- NANO ID
  name              VARCHAR(100) NOT NULL,
  abbreviation      VARCHAR(10) NOT NULL,
  color             VARCHAR(7) NOT NULL,
  type              mode_type NOT NULL,
  id_course_theory  INTEGER,
  academic_year     SMALLINT NOT NULL, -- 1 a 5 año
  id_teacher        INTEGER,
  created_at        TIMESTAMPTZ DEFAULT NOW(),

  CONSTRAINT unique_course_name_abbreviation UNIQUE (name, abbreviation),
  CONSTRAINT valid_academic_year CHECK (academic_year > 0 AND academic_year <= 5),
  CONSTRAINT valid_type_course CHECK (
    (type = 'LABORATORY' AND id_course_theory IS NOT NULL) OR
    (type = 'THEORY' AND id_course_theory IS NULL)
  ),
  CONSTRAINT fk_course_teacher FOREIGN KEY (id_teacher) REFERENCES teachers(id) ON DELETE SET NULL,
  CONSTRAINT fk_course_theory FOREIGN KEY (id_course_theory) REFERENCES courses(id) ON DELETE SET NULL
);

-- ============================================
-- GRUPOS
-- ============================================
CREATE TABLE groups (
  id         SERIAL PRIMARY KEY,
  code       VARCHAR(24) NOT NULL UNIQUE, -- NANO ID
  name       VARCHAR(10) NOT NULL,
  id_course  INTEGER NOT NULL,
  id_classroom INTEGER,
  created_at TIMESTAMPTZ DEFAULT NOW(),

  CONSTRAINT unique_group_per_course UNIQUE(id_course, name),
  CONSTRAINT fk_group_course FOREIGN KEY (id_course) REFERENCES courses(id) ON DELETE CASCADE,
  CONSTRAINT fk_group_classroom FOREIGN KEY (id_classroom) REFERENCES classrooms(id) ON DELETE SET NULL
);

-- ============================================
-- HORARIOS
-- ============================================
CREATE TABLE schedule (
  id                  SERIAL PRIMARY KEY,
  id_group            INTEGER NOT NULL,
  day                 week_day NOT NULL,
  start_hour_academic SMALLINT NOT NULL,
  duration_hours      SMALLINT NOT NULL, -- máximo 4 horas académicas
/*   id_classroom        INTEGER, --opcional solo en caso de asignación fija */
  created_at          TIMESTAMPTZ DEFAULT NOW(),

  CONSTRAINT valid_duration CHECK (duration_hours > 0 AND duration_hours <= 4),
  CONSTRAINT fk_schedule_group FOREIGN KEY (id_group) REFERENCES groups(id) ON DELETE CASCADE,
  CONSTRAINT fk_schedule_start_hour FOREIGN KEY (start_hour_academic) REFERENCES academic_hours(hour_number) ON DELETE CASCADE
);

-- Índices
CREATE INDEX idx_course_code ON courses(code);
CREATE INDEX idx_course_type ON courses(type);
CREATE INDEX idx_course_academic_year ON courses(academic_year);
CREATE INDEX idx_group_code ON groups(code);