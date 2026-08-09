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

	"github.com/jackc/pgx/v5"
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
	courseCodePattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9 .-]{0,31}$`)
	groupNamePattern  = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9-]{0,9}$`)
	colorPattern      = regexp.MustCompile(`^#[0-9A-Fa-f]{6}$`)
)

var courseColors = []string{
	"#1677ff", "#7c3aed", "#e11d73", "#e03131", "#f76707",
	"#d99500", "#00a878", "#008f9c", "#0077b6", "#5f3dc4",
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
	case "23P01":
		writeError(w, http.StatusConflict, "el aula ya está ocupada en ese horario")
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
	_, ok := catalog.auth.requireEditor(w, r)
	return ok
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

type classroomRequest struct {
	Code     string `json:"code"`
	Type     string `json:"type"`
	Floor    *int16 `json:"floor"`
	Capacity *int16 `json:"capacity"`
}

func classroomParams(request classroomRequest) (database.CreateClassroomParams, error) {
	code := strings.TrimSpace(request.Code)
	if len(code) == 0 || len(code) > 32 {
		return database.CreateClassroomParams{}, fmt.Errorf("code debe tener entre 1 y 32 caracteres")
	}
	if request.Type != "THEORY" && request.Type != "LABORATORY" {
		return database.CreateClassroomParams{}, fmt.Errorf("type debe ser THEORY o LABORATORY")
	}
	params := database.CreateClassroomParams{Code: code, Type: database.AppModeType(request.Type)}
	if request.Floor != nil {
		if *request.Floor < 0 || *request.Floor > 40 {
			return database.CreateClassroomParams{}, fmt.Errorf("floor fuera de rango")
		}
		params.Floor = pgtype.Int2{Int16: *request.Floor, Valid: true}
	}
	if request.Capacity != nil {
		if *request.Capacity <= 0 {
			return database.CreateClassroomParams{}, fmt.Errorf("capacity debe ser positiva")
		}
		params.Capacity = pgtype.Int2{Int16: *request.Capacity, Valid: true}
	}
	return params, nil
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
	var request classroomRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid classroom payload")
		return
	}

	params, err := classroomParams(request)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
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

func (catalog *catalogHandler) updateClassroom(w http.ResponseWriter, r *http.Request) {
	if !requireCatalogUser(catalog, w, r) {
		return
	}
	id, ok := pathID(r, "id")
	if !ok {
		writeError(w, http.StatusBadRequest, "id inválido")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 16<<10)
	var request classroomRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid classroom payload")
		return
	}
	params, err := classroomParams(request)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	ctx, cancel := catalogTimeout(r)
	defer cancel()
	existingType, err := catalog.queries.GetClassroomType(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		writeError(w, http.StatusNotFound, "aula no encontrada")
		return
	}
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, "database unavailable")
		return
	}
	if existingType != params.Type {
		writeError(w, http.StatusBadRequest, "no se puede cambiar la modalidad de un aula")
		return
	}
	row, err := catalog.queries.UpdateClassroom(ctx, database.UpdateClassroomParams{
		ID: id, Code: params.Code, Type: params.Type, Floor: params.Floor, Capacity: params.Capacity,
	})
	if err != nil {
		if writeConflict(err, w) {
			return
		}
		writeError(w, http.StatusServiceUnavailable, "database unavailable")
		return
	}
	writeJSON(w, http.StatusOK, classroomJSON{ID: row.ID, Code: row.Code, Type: string(row.Type),
		Floor: nullableInt16(row.Floor), Capacity: nullableInt16(row.Capacity)})
}

func nullableInt16(value pgtype.Int2) *int16 {
	if !value.Valid {
		return nil
	}
	return &value.Int16
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

type teacherRequest struct {
	Name     string `json:"name"`
	LastName string `json:"lastName"`
}

func teacherNames(request teacherRequest) (string, string, error) {
	name := strings.TrimSpace(request.Name)
	lastName := strings.TrimSpace(request.LastName)
	if len(name) < 2 || len(name) > 100 || len(lastName) < 2 || len(lastName) > 100 {
		return "", "", fmt.Errorf("name y lastName deben tener entre 2 y 100 caracteres")
	}
	return name, lastName, nil
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
	var request teacherRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid teacher payload")
		return
	}

	name, lastName, err := teacherNames(request)
	if err != nil {
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

func (catalog *catalogHandler) updateTeacher(w http.ResponseWriter, r *http.Request) {
	if !requireCatalogUser(catalog, w, r) {
		return
	}
	id, ok := pathID(r, "id")
	if !ok {
		writeError(w, http.StatusBadRequest, "id inválido")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 16<<10)
	var request teacherRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid teacher payload")
		return
	}
	name, lastName, err := teacherNames(request)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	ctx, cancel := catalogTimeout(r)
	defer cancel()
	row, err := catalog.queries.UpdateTeacher(ctx, database.UpdateTeacherParams{ID: id, Name: name, LastName: lastName})
	if errors.Is(err, pgx.ErrNoRows) {
		writeError(w, http.StatusNotFound, "docente no encontrado")
		return
	}
	if err != nil {
		if writeConflict(err, w) {
			return
		}
		writeError(w, http.StatusServiceUnavailable, "database unavailable")
		return
	}
	writeJSON(w, http.StatusOK, teacherItemJSON{ID: row.ID, Name: row.Name, LastName: row.LastName,
		FullName: strings.TrimSpace(row.Name + " " + row.LastName)})
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

type courseRequest struct {
	Name           string `json:"name"`
	Abbreviation   string `json:"abbreviation"`
	Type           string `json:"type"`
	AcademicYear   int16  `json:"academicYear"`
	TheoryCourseID *int32 `json:"theoryCourseId"`
	TeacherID      *int32 `json:"teacherId"`
	Credits        *int16 `json:"credits"`
	Color          string `json:"color"`
}

type courseValues struct {
	name           string
	abbreviation   string
	code           string
	color          string
	typeValue      database.AppModeType
	academicYear   int16
	credits        pgtype.Int2
	theoryCourseID pgtype.Int4
	teacherID      pgtype.Int4
}

func normalizeCourseRequest(request courseRequest) (courseValues, error) {
	values := courseValues{
		name:         strings.TrimSpace(request.Name),
		abbreviation: strings.ToUpper(strings.TrimSpace(request.Abbreviation)),
		typeValue:    database.AppModeType(request.Type),
		academicYear: request.AcademicYear,
	}
	if request.Type != "THEORY" && request.Type != "LABORATORY" {
		return values, fmt.Errorf("type debe ser THEORY o LABORATORY")
	}
	if request.Type == "LABORATORY" {
		if strings.HasPrefix(strings.ToLower(values.name), "lab - ") {
			values.name = strings.TrimSpace(values.name[len("lab - "):])
		}
		values.abbreviation = strings.TrimPrefix(values.abbreviation, "LAB-")
		values.abbreviation = strings.TrimSuffix(values.abbreviation, "-L")
		if values.abbreviation != "" {
			values.abbreviation += "-L"
		}
	}
	if len(values.name) < 3 || len(values.name) > 100 {
		return values, fmt.Errorf("name debe tener entre 3 y 100 caracteres")
	}
	if values.abbreviation == "" || len(values.abbreviation) > 20 {
		return values, fmt.Errorf("abbreviation debe tener entre 1 y 20 caracteres")
	}
	if request.AcademicYear < 1 || request.AcademicYear > 5 {
		return values, fmt.Errorf("academicYear debe estar entre 1 y 5")
	}
	if request.Type == "LABORATORY" && request.TheoryCourseID == nil {
		return values, fmt.Errorf("un laboratorio necesita su curso de teoría")
	}

	values.code = values.abbreviation
	if !courseCodePattern.MatchString(values.code) {
		return values, fmt.Errorf("código de curso inválido")
	}

	values.color = strings.TrimSpace(request.Color)
	if values.color == "" {
		hash := 0
		for _, r := range values.name {
			hash = (hash*31 + int(r)) % len(courseColors)
		}
		values.color = courseColors[hash]
	} else if !colorPattern.MatchString(values.color) {
		return values, fmt.Errorf("color debe ser hexadecimal #RRGGBB")
	}
	if request.Credits != nil {
		if *request.Credits <= 0 || *request.Credits > 40 {
			return values, fmt.Errorf("credits fuera de rango")
		}
		values.credits = pgtype.Int2{Int16: *request.Credits, Valid: true}
	}
	if request.TheoryCourseID != nil {
		values.theoryCourseID = pgtype.Int4{Int32: *request.TheoryCourseID, Valid: true}
	}
	if request.TeacherID != nil {
		values.teacherID = pgtype.Int4{Int32: *request.TeacherID, Valid: true}
	}
	return values, nil
}

func validTheoryCourse(ctx context.Context, queries *database.Queries, values courseValues) (bool, error) {
	if values.typeValue == database.AppModeTypeTHEORY {
		return true, nil
	}
	meta, err := queries.GetTheoryCourseMeta(ctx, values.theoryCourseID.Int32)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return meta.Type == database.AppModeTypeTHEORY && meta.CareerCode == defaultCareerCode, nil
}

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
	var request courseRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid course payload")
		return
	}

	values, err := normalizeCourseRequest(request)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	ctx, cancel := catalogTimeout(r)
	defer cancel()
	validTheory, err := validTheoryCourse(ctx, catalog.queries, values)
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, "database unavailable")
		return
	}
	if !validTheory {
		writeError(w, http.StatusBadRequest, "el curso relacionado debe ser una teoría de la misma carrera")
		return
	}

	params := database.CreateCourseParams{
		Code:           values.code,
		Name:           values.name,
		Abbreviation:   values.abbreviation,
		Credits:        values.credits,
		Color:          values.color,
		Type:           values.typeValue,
		Code_2:         defaultCareerCode,
		IDCourseTheory: values.theoryCourseID,
		AcademicYear:   values.academicYear,
		IDTeacher:      values.teacherID,
	}

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

func (catalog *catalogHandler) updateCourse(w http.ResponseWriter, r *http.Request) {
	if !requireCatalogUser(catalog, w, r) {
		return
	}
	id, ok := pathID(r, "id")
	if !ok {
		writeError(w, http.StatusBadRequest, "id inválido")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 16<<10)
	var request courseRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid course payload")
		return
	}
	values, err := normalizeCourseRequest(request)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	ctx, cancel := catalogTimeout(r)
	defer cancel()
	existing, err := catalog.queries.GetCourseMeta(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		writeError(w, http.StatusNotFound, "curso no encontrado")
		return
	}
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, "database unavailable")
		return
	}
	if existing.Type != values.typeValue {
		writeError(w, http.StatusBadRequest, "no se puede cambiar la modalidad de un curso")
		return
	}
	validTheory, err := validTheoryCourse(ctx, catalog.queries, values)
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, "database unavailable")
		return
	}
	if !validTheory {
		writeError(w, http.StatusBadRequest, "el curso relacionado debe ser una teoría de la misma carrera")
		return
	}
	row, err := catalog.queries.UpdateCourse(ctx, database.UpdateCourseParams{
		ID: id, Code: values.code, Name: values.name, Abbreviation: values.abbreviation,
		Credits: values.credits, Color: values.color, IDCourseTheory: values.theoryCourseID,
		AcademicYear: values.academicYear, IDTeacher: values.teacherID,
	})
	if err != nil {
		if writeConflict(err, w) {
			return
		}
		writeError(w, http.StatusServiceUnavailable, "database unavailable")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"id": row.ID, "code": row.Code, "name": row.Name})
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

type groupRequest struct {
	CourseID    int32            `json:"courseId"`
	Name        string           `json:"name"`
	ClassroomID *int32           `json:"classroomId"`
	Sessions    []sessionPayload `json:"sessions"`
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
	var request groupRequest
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
	if errors.Is(err, pgx.ErrNoRows) {
		writeError(w, http.StatusNotFound, "curso no encontrado")
		return
	}
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, "database unavailable")
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

func (catalog *catalogHandler) updateGroup(w http.ResponseWriter, r *http.Request) {
	if !requireCatalogUser(catalog, w, r) {
		return
	}
	id, ok := pathID(r, "id")
	if !ok {
		writeError(w, http.StatusBadRequest, "id inválido")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 32<<10)
	var request groupRequest
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
	groupMeta, err := queries.GetGroupMeta(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		writeError(w, http.StatusNotFound, "grupo no encontrado")
		return
	}
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, "database unavailable")
		return
	}
	if request.CourseID != groupMeta.IDCourse {
		writeError(w, http.StatusBadRequest, "no se puede mover un grupo a otro curso")
		return
	}

	var classroomID pgtype.Int4
	if request.ClassroomID != nil {
		classroomType, err := queries.GetClassroomType(ctx, *request.ClassroomID)
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusBadRequest, "aula no encontrada")
			return
		}
		if err != nil {
			writeError(w, http.StatusServiceUnavailable, "database unavailable")
			return
		}
		if classroomType != groupMeta.Type {
			writeError(w, http.StatusBadRequest, "el aula no corresponde a la modalidad del curso")
			return
		}
		classroomID = pgtype.Int4{Int32: *request.ClassroomID, Valid: true}
	}

	if err := queries.DeleteGroupSessions(ctx, id); err != nil {
		writeError(w, http.StatusServiceUnavailable, "database unavailable")
		return
	}
	group, err := queries.UpdateGroup(ctx, database.UpdateGroupParams{
		ID: id, Code: fmt.Sprintf("%d-%s", request.CourseID, name), Name: name, IDClassroom: classroomID,
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
			IDGroup: id, Day: database.AppWeekDay(session.Day),
			StartHourAcademic: session.StartHourAcademic, DurationHours: session.DurationHours,
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
	writeJSON(w, http.StatusOK, map[string]any{"id": group.ID, "code": group.Code, "name": group.Name})
}

func (catalog *catalogHandler) deleteGroup(w http.ResponseWriter, r *http.Request) {
	catalog.deleteByID(w, r, "id", catalog.queries.DeleteGroup)
}
