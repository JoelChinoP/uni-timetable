// Seed carga el horario 2026-B de Ingeniería de Sistemas en la base.
// Es idempotente: los grupos ya existentes se omiten junto con sus sesiones.
//
//	go -C backend run ./cmd/seed
package main

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed horarios_2026_B_data.json
var dataJSON []byte

const (
	careerCode = "SIS"
	careerName = "Ingeniería de Sistemas"
)

type seedData struct {
	Period    string `json:"period"`
	TimeSlots []struct {
		Start string `json:"start"`
		End   string `json:"end"`
	} `json:"time_slots"`
	Courses map[string]struct {
		Name        string   `json:"name"`
		Year        int      `json:"year"`
		Elective    bool     `json:"elective"`
		TheoryCodes []string `json:"theory_codes"`
		LabCodes    []string `json:"lab_codes"`
	} `json:"courses"`
	Sessions []struct {
		Year      int      `json:"year"`
		Modality  string   `json:"modality"`
		Location  string   `json:"location"`
		Day       string   `json:"day"`
		Start     string   `json:"start"`
		End       string   `json:"end"`
		BlockList []string `json:"block_list"`
		CourseKey string   `json:"course_key"`
		Group     string   `json:"group"`
	} `json:"sessions"`
}

var courseColors = []string{
	"#3b82f6", "#8b5cf6", "#ec4899", "#ef4444", "#f97316",
	"#eab308", "#22c55e", "#14b8a6", "#06b6d4", "#6366f1",
}

var dayToEnum = map[string]string{
	"Lunes": "MONDAY", "Martes": "TUESDAY", "Miércoles": "WEDNESDAY",
	"Jueves": "THURSDAY", "Viernes": "FRIDAY", "Sábado": "SATURDAY",
}

func main() {
	var data seedData
	if err := json.Unmarshal(dataJSON, &data); err != nil {
		log.Fatalf("data embebida inválida: %v", err)
	}

	// ponytail: DATABASE_MIGRATION_URL (session pooler) primero; sin ORM ni prepared statements.
	databaseURL := os.Getenv("DATABASE_MIGRATION_URL")
	if strings.TrimSpace(databaseURL) == "" {
		databaseURL = os.Getenv("DATABASE_URL")
	}
	if strings.TrimSpace(databaseURL) == "" {
		loadDotEnv(".env")
		loadDotEnv("backend/.env")
		databaseURL = os.Getenv("DATABASE_MIGRATION_URL")
		if databaseURL == "" {
			databaseURL = os.Getenv("DATABASE_URL")
		}
	}
	if strings.TrimSpace(databaseURL) == "" {
		log.Fatal("define DATABASE_URL o DATABASE_MIGRATION_URL")
	}

	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		log.Fatal("DATABASE_URL inválida")
	}
	config.MaxConns = 2
	config.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeExec
	config.ConnConfig.RuntimeParams["application_name"] = "uni-timetable-seed"

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		log.Fatalf("conexión: %v", err)
	}
	defer pool.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	if err := seed(ctx, pool, &data); err != nil {
		log.Fatalf("seed: %v", err)
	}
}

func seed(ctx context.Context, pool *pgxpool.Pool, data *seedData) error {
	slotStart := make(map[string]int)
	slotEnd := make(map[string]int)
	for index, slot := range data.TimeSlots {
		slotStart[slot.Start] = index + 1
		slotEnd[slot.End] = index + 1
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var careerID int
	err = tx.QueryRow(ctx, `
		INSERT INTO app.careers (code, name) VALUES ($1, $2)
		ON CONFLICT (code) DO UPDATE SET name = EXCLUDED.name
		RETURNING id`, careerCode, careerName).Scan(&careerID)
	if err != nil {
		return fmt.Errorf("career: %w", err)
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO app.career_available_hours (id_career, hour_number)
		SELECT $1, hour_number FROM app.academic_hours
		ON CONFLICT DO NOTHING`, careerID); err != nil {
		return fmt.Errorf("career hours: %w", err)
	}

	classroomID := make(map[string]int) // "Aula 302|THEORY"
	locations := map[string]string{}    // location -> THEORY|LABORATORY
	for _, session := range data.Sessions {
		modality := "THEORY"
		if session.Modality == "Laboratorio" {
			modality = "LABORATORY"
		}
		locations[session.Location] = modality
	}
	for location, modality := range locations {
		var id int
		err := tx.QueryRow(ctx, `
			INSERT INTO app.classrooms (code, type) VALUES ($1, $2::app.mode_type)
			ON CONFLICT (code, type) DO UPDATE SET code = EXCLUDED.code
			RETURNING id`, location, modality).Scan(&id)
		if err != nil {
			return fmt.Errorf("classroom %s: %w", location, err)
		}
		classroomID[location+"|"+modality] = id
	}

	hasTheory := map[string]bool{}
	hasLab := map[string]bool{}
	for _, session := range data.Sessions {
		if session.Modality == "Laboratorio" {
			hasLab[session.CourseKey] = true
		} else {
			hasTheory[session.CourseKey] = true
		}
	}

	keys := make([]string, 0, len(data.Courses))
	for key := range data.Courses {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	courseID := make(map[string]int) // key|THEORY / key|LABORATORY
	for index, key := range keys {
		course := data.Courses[key]
		if hasTheory[key] && len(course.TheoryCodes) > 0 {
			id, err := insertCourse(ctx, tx, careerID, course.Year, course.Name, course.TheoryCodes[0], key, "THEORY", nil, courseColors[index%len(courseColors)])
			if err != nil {
				return fmt.Errorf("curso %s: %w", key, err)
			}
			courseID[key+"|THEORY"] = id
		}
	}
	for _, key := range keys {
		course := data.Courses[key]
		if hasLab[key] && len(course.LabCodes) > 0 {
			theoryID, ok := courseID[key+"|THEORY"]
			if !ok {
				return fmt.Errorf("curso %s tiene laboratorio sin teoría", key)
			}
			id, err := insertCourse(ctx, tx, careerID, course.Year, course.Name, course.LabCodes[0]+"-L", key+"-L", "LABORATORY", &theoryID, courseColors[indexOf(keys, key)%len(courseColors)])
			if err != nil {
				return fmt.Errorf("curso lab %s: %w", key, err)
			}
			courseID[key+"|LABORATORY"] = id
		}
	}

	groupsSeen := map[string]int{}
	skippedGroups := map[string]bool{}
	var newGroups, sessions int
	for _, session := range data.Sessions {
		modality := "THEORY"
		if session.Modality == "Laboratorio" {
			modality = "LABORATORY"
		}
		cID, ok := courseID[session.CourseKey+"|"+modality]
		if !ok {
			return fmt.Errorf("sesión sin curso: %s %s", session.CourseKey, modality)
		}
		key := fmt.Sprintf("%d|%s", cID, session.Group)
		if skippedGroups[key] {
			continue
		}

		groupID, seen := groupsSeen[key]
		if !seen {
			err := tx.QueryRow(ctx, `
				INSERT INTO app.groups (code, name, id_course, id_classroom)
				VALUES ($1, $2, $3, $4)
				ON CONFLICT (id_course, name) DO NOTHING
				RETURNING id`,
				fmt.Sprintf("%d-%s", cID, session.Group), session.Group, cID,
				classroomID[session.Location+"|"+modality]).Scan(&groupID)
			if err == pgx.ErrNoRows {
				var existing int
				if err := tx.QueryRow(ctx, `SELECT id FROM app.groups WHERE id_course = $1 AND name = $2`, cID, session.Group).Scan(&existing); err != nil {
					return fmt.Errorf("grupo existente %s: %w", key, err)
				}
				groupsSeen[key] = existing
				skippedGroups[key] = true
				continue
			}
			if err != nil {
				return fmt.Errorf("grupo %s: %w", key, err)
			}
			groupsSeen[key] = groupID
			newGroups++
		}

		start, okStart := slotStart[session.Start]
		endSlot, okEnd := slotEnd[session.End]
		if !okStart || !okEnd || endSlot < start {
			return fmt.Errorf("sesión fuera de bloques: %s %s %s-%s", session.CourseKey, session.Day, session.Start, session.End)
		}
		if duration := endSlot - start + 1; duration != len(session.BlockList) {
			return fmt.Errorf("duración inconsistente: %s %s %s-%s (%d bloques != %d slots)",
				session.CourseKey, session.Day, session.Start, session.End, len(session.BlockList), duration)
		}

		tag, err := tx.Exec(ctx, `
			INSERT INTO app.schedule (id_group, day, start_hour_academic, duration_hours)
			VALUES ($1, $2::app.week_day, $3, $4)
			ON CONFLICT (id_group, day, start_hour_academic) DO NOTHING`,
			groupID, dayToEnum[session.Day], start, endSlot-start+1)
		if err != nil {
			return fmt.Errorf("schedule %s %s: %w", session.CourseKey, session.Day, err)
		}
		sessions += int(tag.RowsAffected())
	}

	// Integridad referencial cruzada: ninguna aula debe quedar con sesiones en modalidad opuesta.
	var orphan int
	if err := tx.QueryRow(ctx, `
		SELECT count(*) FROM app.schedule s
		JOIN app.groups g ON g.id = s.id_group
		JOIN app.courses c ON c.id = g.id_course
		JOIN app.classrooms cl ON cl.id = g.id_classroom
		WHERE cl.type <> c.type`).Scan(&orphan); err != nil {
		return err
	}
	if orphan != 0 {
		return fmt.Errorf("integridad: %d sesiones con aula de modalidad incorrecta", orphan)
	}
	var overlaps int
	if err := tx.QueryRow(ctx, `
		SELECT count(*)
		FROM app.schedule left_session
		JOIN app.groups left_group ON left_group.id = left_session.id_group
		JOIN app.schedule right_session ON right_session.id > left_session.id
		  AND right_session.day = left_session.day
		  AND right_session.start_hour_academic < left_session.start_hour_academic + left_session.duration_hours
		  AND left_session.start_hour_academic < right_session.start_hour_academic + right_session.duration_hours
		JOIN app.groups right_group ON right_group.id = right_session.id_group
		  AND right_group.id_classroom = left_group.id_classroom
		WHERE left_group.id_classroom IS NOT NULL`).Scan(&overlaps); err != nil {
		return err
	}
	if overlaps != 0 {
		return fmt.Errorf("integridad: %d cruces de aula", overlaps)
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}
	log.Printf("periodo %s: %d grupos nuevos, %d omitidos, %d sesiones insertadas (fuente: %d)",
		data.Period, newGroups, len(skippedGroups), sessions, len(data.Sessions))
	return nil
}

func insertCourse(ctx context.Context, tx pgx.Tx, careerID, year int, name, abbreviation, code, modality string, theoryID *int, color string) (int, error) {
	var id int
	err := tx.QueryRow(ctx, `
		INSERT INTO app.courses
		  (code, name, abbreviation, color, type, id_career, id_course_theory, academic_year)
		VALUES ($1, $2, $3, $4, $5::app.mode_type, $6, $7, $8)
		ON CONFLICT (code) DO UPDATE SET name = EXCLUDED.name
		RETURNING id`, code, name, abbreviation, color, modality, careerID, theoryID, year).Scan(&id)
	return id, err
}

func indexOf(keys []string, target string) int {
	for index, key := range keys {
		if key == target {
			return index
		}
	}
	return 0
}

// ponytail: .env mínimo KEY=VALUE; el shell o dev.mjs también pueden inyectarlas.
func loadDotEnv(path string) {
	content, err := os.ReadFile(path)
	if err != nil {
		return
	}
	for line := range strings.Lines(string(content)) {
		key, value, ok := strings.Cut(strings.TrimSpace(line), "=")
		if !ok || strings.HasPrefix(key, "#") || os.Getenv(key) != "" {
			continue
		}
		_ = os.Setenv(key, strings.Trim(value, `"'`))
	}
}
