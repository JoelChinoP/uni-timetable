package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JoelChinoP/uni-timetable/backend/internal/database"
)

type catalogHandler struct {
	queries *database.Queries
	db      *pgxpool.Pool
	auth    *authHandler
}

var (
	courseCodePattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9 .-]{0,23}$`)
	groupNamePattern  = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9-]{0,9}$`)
	colorPattern      = regexp.MustCompile(`^#[0-9A-Fa-f]{6}$`)
)

var courseColors = []string{
	"#3b82f6", "#8b5cf6", "#ec4899", "#ef4444", "#f97316",
	"#eab308", "#22c55e", "#14b8a6", "#06b6d4", "#6366f1",
}

func catalogTimeout(r *http.Request) (context.Context, context.CancelFunc) {
	return context.WithTimeout(r.Context(), 5*time.Second)
}

func writeConflict(err error, w http.ResponseWriter) bool {
	var pgError *pgconn.PgError
	if !errors.As(err, &pgError) {
		return false
	}
	switch pgError.Code {
	case "23505":
		writeError(w, http.StatusConflict, "el registro ya existe")
	case "23503":
		writeError(w, http.StatusConflict, "el registro está en uso")
	case "23514", "23502":
		writeError(w, http.StatusBadRequest, "los datos no cumplen las reglas del catálogo")
	default:
		return false
	}
	return true
}

func pathID(r *http.Request, name string) (int32, bool) {
	value, err := strconv.ParseInt(r.PathValue(name), 10, 32)
	return int32(value), err == nil && value > 0
}

func requireCatalogUser(catalog *catalogHandler, w http.ResponseWriter, r *http.Request) bool {
	user, ok := catalog.auth.currentUser(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return false
	}
	if user.ID == 0 {
		writeError(w, http.StatusForbidden, "tu cuenta aún no está registrada")
		return false
	}
	if catalog.queries == nil {
		writeError(w, http.StatusServiceUnavailable, "database is not configured")
		return false
	}
	return true
}

// ---------- Aulas ----------

func (catalog *catalogHandler) classrooms(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		catalog.listClassrooms(w, r)
	case http.MethodPost:
		catalog.createClassroom(w, r)
	default:
		w.Header().Set("Allow", "GET, POST")
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (catalog *catalogHandler) listClassrooms(w http.ResponseWriter, r *http.Request) {
	if catalog.queries == nil {
		writeError(w, http.StatusServiceUnavailable, "database is not configured")
		return
	}
	ctx, cancel := catalogTimeout(r)
	defer cancel()

	rows, err := catalog.queries.ListClassrooms(ctx)
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, "database unavailable")
		return
	}
	writeJSON(w, http.StatusOK, mapClassrooms(rows))
}

type classroomJSON struct {
	ID       int32  `json:"id"`
	Code     string `json:"code"`
	Type     string `json:"type"`
	Floor    *int16 `json:"floor"`
	Capacity *int16 `json:"capacity"`
}

func mapClassrooms(rows []database.ListClassroomsRow) []classroomJSON {
	items := make([]classroomJSON, 0, len(rows))
	for _, row := range rows {
		item := classroomJSON{ID: row.ID, Code: row.Code, Type: string(row.Type)}
		if row.Floor.Valid {
			item.Floor = &row.Floor.Int16
		}
		if row.Capacity.Valid {
			item.Capacity = &row.Capacity.Int16
		}
		items = append(items, item)
	}
	return items
}

func (catalog *catalogHandler) createClassroom(w http.ResponseWriter, r *http.Request) {
	if !requireCatalogUser(catalog, w, r) {
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 16<<10)
	var request struct {
		Code     string `json:"code"`
		Type     string `json:"type"`
		Floor    *int16 `json:"floor"`
		Capacity *int16 `json:"capacity"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid classroom payload")
		return
	}

	code := strings.TrimSpace(request.Code)
	if len(code) == 0 || len(code) > 32 {
		writeError(w, http.StatusBadRequest, "code debe tener entre 1 y 32 caracteres")
		return
	}
	if request.Type != "THEORY" && request.Type != "LABORATORY" {
		writeError(w, http.StatusBadRequest, "type debe ser THEORY o LABORATORY")
		return
	}
	params := database.CreateClassroomParams{Code: code, Type: database.AppModeType(request.Type)}
	if request.Floor != nil {
		if *request.Floor < 0 || *request.Floor > 40 {
			writeError(w, http.StatusBadRequest, "floor fuera de rango")
			return
		}
		params.Floor = pgtype.Int2{Int16: *request.Floor, Valid: true}
	}
	if request.Capacity != nil {
		if *request.Capacity <= 0 {
			writeError(w, http.StatusBadRequest, "capacity debe ser positiva")
			return
		}
		params.Capacity = pgtype.Int2{Int16: *request.Capacity, Valid: true}
	}

	ctx, cancel := catalogTimeout(r)
	defer cancel()

	row, err := catalog.queries.CreateClassroom(ctx, params)
	if err != nil {
		if writeConflict(err, w) {
			return
		}
		writeError(w, http.StatusServiceUnavailable, "database unavailable")
		return
	}
	writeJSON(w, http.StatusCreated, mapClassrooms([]database.ListClassroomsRow{{
		ID: row.ID, Code: row.Code, Type: row.Type, Floor: row.Floor, Capacity: row.Capacity,
	}})[0])
}

func (catalog *catalogHandler) deleteClassroom(w http.ResponseWriter, r *http.Request) {
	catalog.deleteByID(w, r, "id", catalog.queries.DeleteClassroom)
}

// ---------- Docentes ----------

func (catalog *catalogHandler) teachers(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		catalog.listTeachers(w, r)
	case http.MethodPost:
		catalog.createTeacher(w, r)
	default:
		w.Header().Set("Allow", "GET, POST")
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

type teacherItemJSON struct {
	ID       int32  `json:"id"`
	Name     string `json:"name"`
	LastName string `json:"lastName"`
	FullName string `json:"fullName"`
}

func mapTeachers(rows []database.ListTeachersRow) []teacherItemJSON {
	items := make([]teacherItemJSON, 0, len(rows))
	for _, row := range rows {
		items = append(items, teacherItemJSON{
			ID:       row.ID,
			Name:     row.Name,
			LastName: row.LastName,
			FullName: strings.TrimSpace(row.Name + " " + row.LastName),
		})
	}
	return items
}

func (catalog *catalogHandler) listTeachers(w http.ResponseWriter, r *http.Request) {
	if catalog.queries == nil {
		writeError(w, http.StatusServiceUnavailable, "database is not configured")
		return
	}
	ctx, cancel := catalogTimeout(r)
	defer cancel()

	rows, err := catalog.queries.ListTeachers(ctx, defaultCareerCode)
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, "database unavailable")
		return
	}
	writeJSON(w, http.StatusOK, mapTeachers(rows))
}

func (catalog *catalogHandler) createTeacher(w http.ResponseWriter, r *http.Request) {
	if !requireCatalogUser(catalog, w, r) {
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 16<<10)
	var request struct {
		Name     string `json:"name"`
		LastName string `json:"lastName"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid teacher payload")
		return
	}

	name := strings.TrimSpace(request.Name)
	lastName := strings.TrimSpace(request.LastName)
	if len(name) < 2 || len(name) > 100 || len(lastName) < 2 || len(lastName) > 100 {
		writeError(w, http.StatusBadRequest, "name y lastName deben tener entre 2 y 100 caracteres")
		return
	}

	ctx, cancel := catalogTimeout(r)
	defer cancel()

	tx, err := catalog.db.Begin(ctx)
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, "database unavailable")
		return
	}
	defer tx.Rollback(ctx)

	queries := catalog.queries.WithTx(tx)
	teacherID, err := queries.UpsertTeacher(ctx, database.UpsertTeacherParams{Name: name, LastName: lastName})
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, "database unavailable")
		return
	}
	if err := queries.LinkTeacherToCareer(ctx, database.LinkTeacherToCareerParams{
		Code:      defaultCareerCode,
		IDTeacher: teacherID,
	}); err != nil {
		writeError(w, http.StatusServiceUnavailable, "database unavailable")
		return
	}
	if err := tx.Commit(ctx); err != nil {
		writeError(w, http.StatusServiceUnavailable, "database unavailable")
		return
	}
	writeJSON(w, http.StatusCreated, teacherItemJSON{
		ID:       teacherID,
		Name:     name,
		LastName: lastName,
		FullName: strings.TrimSpace(name + " " + lastName),
	})
}

func (catalog *catalogHandler) deleteTeacher(w http.ResponseWriter, r *http.Request) {
	catalog.deleteByID(w, r, "id", catalog.queries.DeleteTeacher)
}

func (catalog *catalogHandler) deleteByID(
	w http.ResponseWriter,
	r *http.Request,
	pathName string,
	deleteFn func(context.Context, int32) (int64, error),
) {
	if !requireCatalogUser(catalog, w, r) {
		return
	}
	id, ok := pathID(r, pathName)
	if !ok {
		writeError(w, http.StatusBadRequest, "id inválido")
		return
	}
	ctx, cancel := catalogTimeout(r)
	defer cancel()

	affected, err := deleteFn(ctx, id)
	if err != nil {
		if writeConflict(err, w) {
			return
		}
		writeError(w, http.StatusServiceUnavailable, "database unavailable")
		return
	}
	if affected == 0 {
		writeError(w, http.StatusNotFound, "registro no encontrado")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ---------- Cursos ----------

func (catalog *catalogHandler) courses(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	catalog.createCourse(w, r)
}

func (catalog *catalogHandler) createCourse(w http.ResponseWriter, r *http.Request) {
	if !requireCatalogUser(catalog, w, r) {
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 16<<10)
	var request struct {
		Name           string `json:"name"`
		Abbreviation   string `json:"abbreviation"`
		Type           string `json:"type"`
		AcademicYear   int16  `json:"academicYear"`
		TheoryCourseID *int32 `json:"theoryCourseId"`
		TeacherID      *int32 `json:"teacherId"`
		Credits        *int16 `json:"credits"`
		Color          string `json:"color"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid course payload")
		return
	}

	name := strings.TrimSpace(request.Name)
	abbreviation := strings.ToUpper(strings.TrimSpace(request.Abbreviation))
	if len(name) < 3 || len(name) > 100 {
		writeError(w, http.StatusBadRequest, "name debe tener entre 3 y 100 caracteres")
		return
	}
	if abbreviation == "" || len(abbreviation) > 10 {
		writeError(w, http.StatusBadRequest, "abbreviation debe tener entre 1 y 10 caracteres")
		return
	}
	if request.Type != "THEORY" && request.Type != "LABORATORY" {
		writeError(w, http.StatusBadRequest, "type debe ser THEORY o LABORATORY")
		return
	}
	if request.AcademicYear < 1 || request.AcademicYear > 5 {
		writeError(w, http.StatusBadRequest, "academicYear debe estar entre 1 y 5")
		return
	}
	if request.Type == "LABORATORY" && request.TheoryCourseID == nil {
		writeError(w, http.StatusBadRequest, "un laboratorio necesita su curso de teoría")
		return
	}

	code := abbreviation
	if request.Type == "LABORATORY" {
		// ponytail: sufijo fijo evita colisiones con la UNIQUE (career, name, abbreviation).
		code += "-L"
		abbreviation += "-L"
	}
	if !courseCodePattern.MatchString(code) {
		writeError(w, http.StatusBadRequest, "código de curso inválido")
		return
	}

	color := strings.TrimSpace(request.Color)
	if color == "" {
		// ponytail: palette + stable hash keeps colors consistent without a picker.
		hash := 0
		for _, r := range name {
			hash = (hash*31 + int(r)) % len(courseColors)
		}
		color = courseColors[hash]
	} else if !colorPattern.MatchString(color) {
		writeError(w, http.StatusBadRequest, "color debe ser hexadecimal #RRGGBB")
		return
	}

	params := database.CreateCourseParams{
		Code:         code,
		Name:         name,
		Abbreviation: abbreviation,
		Color:        color,
		Type:         database.AppModeType(request.Type),
		Code_2:       defaultCareerCode,
		AcademicYear: request.AcademicYear,
	}
	if request.Credits != nil {
		if *request.Credits <= 0 || *request.Credits > 40 {
			writeError(w, http.StatusBadRequest, "credits fuera de rango")
			return
		}
		params.Credits = pgtype.Int2{Int16: *request.Credits, Valid: true}
	}
	if request.TheoryCourseID != nil {
		params.IDCourseTheory = pgtype.Int4{Int32: *request.TheoryCourseID, Valid: true}
	}
	if request.TeacherID != nil {
		params.IDTeacher = pgtype.Int4{Int32: *request.TeacherID, Valid: true}
	}

	ctx, cancel := catalogTimeout(r)
	defer cancel()

	row, err := catalog.queries.CreateCourse(ctx, params)
	if err != nil {
		if writeConflict(err, w) {
			return
		}
		writeError(w, http.StatusServiceUnavailable, "database unavailable")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"id": row.ID, "code": row.Code, "name": row.Name})
}

func (catalog *catalogHandler) deleteCourse(w http.ResponseWriter, r *http.Request) {
	catalog.deleteByID(w, r, "id", catalog.queries.DeleteCourse)
}

// ---------- Grupos y sus horarios ----------

type sessionPayload struct {
	Day               string `json:"day"`
	StartHourAcademic int16  `json:"startHourAcademic"`
	DurationHours     int16  `json:"durationHours"`
}

func validateSessions(sessions []sessionPayload) error {
	if len(sessions) == 0 || len(sessions) > 14 {
		return fmt.Errorf("registra entre 1 y 14 sesiones")
	}
	validDays := map[string]bool{}
	for _, day := range plannerDayOrder {
		validDays[day] = true
	}
	for i, session := range sessions {
		if !validDays[session.Day] {
			return fmt.Errorf("sesión %d: día inválido", i+1)
		}
		if session.StartHourAcademic < 1 || session.StartHourAcademic > 15 {
			return fmt.Errorf("sesión %d: hora de inicio fuera de rango", i+1)
		}
		if session.DurationHours < 1 || session.DurationHours > 6 {
			return fmt.Errorf("sesión %d: duración fuera de rango", i+1)
		}
		if session.StartHourAcademic+session.DurationHours > 16 {
			return fmt.Errorf("sesión %d: se sale del día académico", i+1)
		}
		for j := i + 1; j < len(sessions); j++ {
			other := sessions[j]
			if session.Day != other.Day {
				continue
			}
			endSession := session.StartHourAcademic + session.DurationHours
			endOther := other.StartHourAcademic + other.DurationHours
			if session.StartHourAcademic < endOther && other.StartHourAcademic < endSession {
				return fmt.Errorf("sesiones %d y %d se cruzan el %s", i+1, j+1, session.Day)
			}
		}
	}
	return nil
}

func (catalog *catalogHandler) groups(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	catalog.createGroup(w, r)
}

func (catalog *catalogHandler) createGroup(w http.ResponseWriter, r *http.Request) {
	if !requireCatalogUser(catalog, w, r) {
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 32<<10)
	var request struct {
		CourseID    int32            `json:"courseId"`
		Name        string           `json:"name"`
		ClassroomID *int32           `json:"classroomId"`
		Sessions    []sessionPayload `json:"sessions"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid group payload")
		return
	}

	name := strings.ToUpper(strings.TrimSpace(request.Name))
	if !groupNamePattern.MatchString(name) {
		writeError(w, http.StatusBadRequest, "name debe ser corto (A, B, C…)")
		return
	}
	if err := validateSessions(request.Sessions); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	ctx, cancel := catalogTimeout(r)
	defer cancel()

	tx, err := catalog.db.Begin(ctx)
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, "database unavailable")
		return
	}
	defer tx.Rollback(ctx)

	queries := catalog.queries.WithTx(tx)
	course, err := queries.GetCourseMeta(ctx, request.CourseID)
	if err != nil {
		writeError(w, http.StatusNotFound, "curso no encontrado")
		return
	}

	var classroomID pgtype.Int4
	if request.ClassroomID != nil {
		classroomType, err := queries.GetClassroomType(ctx, *request.ClassroomID)
		if err != nil {
			writeError(w, http.StatusBadRequest, "aula no encontrada")
			return
		}
		if classroomType != course.Type {
			writeError(w, http.StatusBadRequest, "el aula no corresponde a la modalidad del curso")
			return
		}
		classroomID = pgtype.Int4{Int32: *request.ClassroomID, Valid: true}
	}

	group, err := queries.CreateGroup(ctx, database.CreateGroupParams{
		Code:        fmt.Sprintf("%d-%s", request.CourseID, name),
		Name:        name,
		IDCourse:    request.CourseID,
		IDClassroom: classroomID,
	})
	if err != nil {
		if writeConflict(err, w) {
			return
		}
		writeError(w, http.StatusServiceUnavailable, "database unavailable")
		return
	}

	for _, session := range request.Sessions {
		err := queries.CreateGroupSession(ctx, database.CreateGroupSessionParams{
			IDGroup:           group.ID,
			Day:               database.AppWeekDay(session.Day),
			StartHourAcademic: session.StartHourAcademic,
			DurationHours:     session.DurationHours,
		})
		if err != nil {
			if writeConflict(err, w) {
				return
			}
			writeError(w, http.StatusServiceUnavailable, "database unavailable")
			return
		}
	}

	if err := tx.Commit(ctx); err != nil {
		writeError(w, http.StatusServiceUnavailable, "database unavailable")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"id": group.ID, "code": group.Code, "name": group.Name})
}

func (catalog *catalogHandler) deleteGroup(w http.ResponseWriter, r *http.Request) {
	catalog.deleteByID(w, r, "id", catalog.queries.DeleteGroup)
}
