package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JoelChinoP/uni-timetable/backend/internal/database"
)

// Integración real contra PostgreSQL. El esquema original se aparta y restaura
// exactamente al finalizar, aun cuando una prueba falle.
func testPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL no definida")
	}
	if os.Getenv("ALLOW_DESTRUCTIVE_TESTS") != "1" {
		t.Skip("define ALLOW_DESTRUCTIVE_TESTS=1 para ejecutar integración")
	}

	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	config.MaxConns = 2
	config.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeExec

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)

	schema, err := os.ReadFile("../../internal/database/schema.sql")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var backupExists bool
	if err := pool.QueryRow(ctx, "SELECT to_regnamespace('app_test_backup') IS NOT NULL").Scan(&backupExists); err != nil {
		t.Fatal(err)
	}
	if backupExists {
		t.Fatal("app_test_backup existe; restaúralo manualmente antes de ejecutar las pruebas")
	}
	var appExists bool
	if err := pool.QueryRow(ctx, "SELECT to_regnamespace('app') IS NOT NULL").Scan(&appExists); err != nil {
		t.Fatal(err)
	}
	if appExists {
		if _, err := pool.Exec(ctx, "ALTER SCHEMA app RENAME TO app_test_backup"); err != nil {
			t.Fatal(err)
		}
	}
	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cleanupCancel()
		if _, err := pool.Exec(cleanupCtx, "DROP SCHEMA IF EXISTS app CASCADE"); err != nil {
			t.Errorf("limpiar esquema de pruebas: %v", err)
			return
		}
		if appExists {
			if _, err := pool.Exec(cleanupCtx, "ALTER SCHEMA app_test_backup RENAME TO app"); err != nil {
				t.Errorf("restaurar esquema original: %v", err)
			}
		}
	})
	if _, err := pool.Exec(ctx, string(schema)); err != nil {
		t.Fatalf("schema bootstrap: %v", err)
	}
	return pool
}

func registerUser(t *testing.T, pool *pgxpool.Pool, email string, role string) {
	t.Helper()
	_, err := pool.Exec(context.Background(),
		"INSERT INTO app.users (email, display_name, role) VALUES ($1, 'Test', $2::app.user_role)", email, role)
	if err != nil {
		t.Fatal(err)
	}
}

func authCookie(t *testing.T, auth *authHandler, user AuthUser) *http.Cookie {
	t.Helper()
	token, err := auth.sessions.create(user)
	if err != nil {
		t.Fatal(err)
	}
	return &http.Cookie{Name: sessionCookieName, Value: token}
}

func doJSON(handler http.Handler, method, path, body string, cookie *http.Cookie) *httptest.ResponseRecorder {
	var reader *strings.Reader
	if body == "" {
		reader = strings.NewReader("")
	} else {
		reader = strings.NewReader(body)
	}
	request := httptest.NewRequest(method, path, reader)
	if cookie != nil {
		request.AddCookie(cookie)
	}
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	return response
}

func TestIntegrationCatalogFlow(t *testing.T) {
	pool := testPool(t)
	queries := database.New(pool)
	sessions := newSessionStore()
	auth := &authHandler{queries: queries, sessions: sessions}
	handler := routes(routeConfig{db: pool, sessions: sessions})

	// El rol de la cookie se ignora: la BD resuelve EDITOR y no permite gestionar usuarios.
	registerUser(t, pool, "user@unsa.edu.pe", "EDITOR")
	cookie := authCookie(t, auth, AuthUser{ID: 99, Email: "user@unsa.edu.pe", Role: "ADMIN", EmailVerified: true})

	mustSeed := func() {
		_, err := pool.Exec(context.Background(), `
			INSERT INTO app.careers (code, name) VALUES ('SIS', 'Ingeniería de Sistemas')`)
		if err != nil {
			t.Fatal(err)
		}
	}
	mustSeed()

	if response := doJSON(handler, "GET", "/users", "", cookie); response.Code != http.StatusForbidden {
		t.Fatalf("EDITOR no debe listar usuarios: %d", response.Code)
	}
	registerUser(t, pool, "admin@unsa.edu.pe", "ADMIN")
	adminCookie := authCookie(t, auth, AuthUser{Email: "admin@unsa.edu.pe", Role: "VIEWER", EmailVerified: true})
	response := doJSON(handler, "GET", "/users?page=1&pageSize=1", "", adminCookie)
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"pageSize":1`) || !strings.Contains(response.Body.String(), `"total":2`) {
		t.Fatalf("paginated users: %d %s", response.Code, response.Body.String())
	}
	response = doJSON(handler, "POST", "/users", `{"email":"new@unsa.edu.pe","displayName":"Nueva Editora"}`, adminCookie)
	if response.Code != http.StatusCreated || !strings.Contains(response.Body.String(), `"role":"EDITOR"`) {
		t.Fatalf("create editor: %d %s", response.Code, response.Body.String())
	}
	response = doJSON(handler, "PUT", "/users/3", `{"email":"editora@unsa.edu.pe","displayName":"Editora Actualizada","role":"EDITOR"}`, adminCookie)
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), "Editora Actualizada") {
		t.Fatalf("update editor: %d %s", response.Code, response.Body.String())
	}
	response = doJSON(handler, "DELETE", "/users/3", "", adminCookie)
	if response.Code != http.StatusNoContent {
		t.Fatalf("delete editor: %d %s", response.Code, response.Body.String())
	}
	response = doJSON(handler, "DELETE", "/users/2", "", adminCookie)
	if response.Code != http.StatusConflict {
		t.Fatalf("self delete admin: %d, want 409", response.Code)
	}

	response = doJSON(handler, "POST", "/classrooms", `{"code":"Aula 999","type":"THEORY"}`, cookie)
	if response.Code != http.StatusCreated {
		t.Fatalf("create classroom: %d %s", response.Code, response.Body.String())
	}

	response = doJSON(handler, "POST", "/classrooms", `{"code":"Aula 999","type":"THEORY"}`, cookie)
	if response.Code != http.StatusConflict {
		t.Fatalf("classroom duplicate: %d, want 409", response.Code)
	}
	response = doJSON(handler, "PUT", "/classrooms/1", `{"code":"Aula 999A","type":"THEORY","floor":3,"capacity":35}`, cookie)
	if response.Code != http.StatusOK {
		t.Fatalf("update classroom: %d %s", response.Code, response.Body.String())
	}

	response = doJSON(handler, "POST", "/classrooms", `{"code":"X","type":"GYM"}`, cookie)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("bad classroom type: %d, want 400", response.Code)
	}

	response = doJSON(handler, "POST", "/teachers", `{"name":"Ada","lastName":"Lovelace"}`, cookie)
	if response.Code != http.StatusCreated {
		t.Fatalf("create teacher: %d %s", response.Code, response.Body.String())
	}
	createTeacherAgain := doJSON(handler, "POST", "/teachers", `{"name":"Ada","lastName":"Lovelace"}`, cookie)
	if createTeacherAgain.Code != http.StatusCreated {
		t.Fatalf("teacher re-register must be idempotent: %d", createTeacherAgain.Code)
	}
	response = doJSON(handler, "PUT", "/teachers/1", `{"name":"Ada","lastName":"Byron"}`, cookie)
	if response.Code != http.StatusOK {
		t.Fatalf("update teacher: %d %s", response.Code, response.Body.String())
	}

	response = doJSON(handler, "GET", "/teachers", "", nil)
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), "Byron") {
		t.Fatalf("teachers list: %d %s", response.Code, response.Body.String())
	}

	response = doJSON(handler, "POST", "/courses",
		`{"name":"Cálculo","abbreviation":"CALC","type":"THEORY","academicYear":1,"teacherId":1}`, cookie)
	if response.Code != http.StatusCreated {
		t.Fatalf("create course: %d %s", response.Code, response.Body.String())
	}
	response = doJSON(handler, "PUT", "/courses/1",
		`{"name":"Cálculo I","abbreviation":"CALC","type":"THEORY","academicYear":1,"teacherId":1}`, cookie)
	if response.Code != http.StatusOK {
		t.Fatalf("update course: %d %s", response.Code, response.Body.String())
	}

	groupBody := `{"courseId":1,"name":"A","classroomId":1,"sessions":[
		{"day":"MONDAY","startHourAcademic":1,"durationHours":2},
		{"day":"WEDNESDAY","startHourAcademic":3,"durationHours":3}]}`
	response = doJSON(handler, "POST", "/groups", groupBody, cookie)
	if response.Code != http.StatusCreated {
		t.Fatalf("create group: %d %s", response.Code, response.Body.String())
	}
	roomConflictBody := `{"courseId":1,"name":"B","classroomId":1,"sessions":[
		{"day":"MONDAY","startHourAcademic":2,"durationHours":2}]}`
	response = doJSON(handler, "POST", "/groups", roomConflictBody, cookie)
	if response.Code != http.StatusConflict {
		t.Fatalf("classroom overlap: %d %s, want 409", response.Code, response.Body.String())
	}

	response = doJSON(handler, "POST", "/groups", groupBody, cookie)
	if response.Code != http.StatusConflict {
		t.Fatalf("group duplicate: %d, want 409", response.Code)
	}

	overlapBody := `{"courseId":1,"name":"B","sessions":[
		{"day":"MONDAY","startHourAcademic":1,"durationHours":4},
		{"day":"MONDAY","startHourAcademic":3,"durationHours":2}]}`
	response = doJSON(handler, "POST", "/groups", overlapBody, cookie)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("overlapping sessions: %d, want 400", response.Code)
	}

	wrongClassroom := `{"courseId":1,"name":"C","classroomId":99,"sessions":[{"day":"MONDAY","startHourAcademic":5,"durationHours":2}]}`
	response = doJSON(handler, "POST", "/groups", wrongClassroom, cookie)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("missing classroom: %d, want 400", response.Code)
	}

	labCourse := doJSON(handler, "POST", "/courses",
		`{"name":"Cálculo","abbreviation":"CALC","type":"LABORATORY","academicYear":1,"theoryCourseId":1}`, cookie)
	if labCourse.Code != http.StatusCreated {
		t.Fatalf("create lab course: %d %s", labCourse.Code, labCourse.Body.String())
	}

	labGroupWrongRoom := `{"courseId":2,"name":"A","classroomId":1,"sessions":[{"day":"TUESDAY","startHourAcademic":9,"durationHours":2}]}`
	response = doJSON(handler, "POST", "/groups", labGroupWrongRoom, cookie)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("lab group in theory room: %d, want 400", response.Code)
	}

	response = doJSON(handler, "GET", "/planner/dashboard", "", nil)
	if response.Code != http.StatusOK {
		t.Fatalf("dashboard: %d", response.Code)
	}
	var payload struct {
		Data struct {
			TermLabel     string `json:"termLabel"`
			Days          []string
			AcademicHours []struct {
				StartTime string `json:"startTime"`
				EndTime   string `json:"endTime"`
			} `json:"academicHours"`
			Courses []struct {
				ID     int    `json:"id"`
				Type   string `json:"type"`
				Groups []struct {
					Name     string `json:"name"`
					Sessions []struct {
						Day               string `json:"day"`
						StartHourAcademic int16  `json:"startHourAcademic"`
					} `json:"sessions"`
				} `json:"groups"`
			} `json:"courses"`
		} `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if len(payload.Data.AcademicHours) != 15 {
		t.Fatalf("academic hours = %d, want 15", len(payload.Data.AcademicHours))
	}
	if payload.Data.AcademicHours[1].EndTime != "08:40" {
		t.Fatalf("hora 2 corregida = %s, want 08:40", payload.Data.AcademicHours[1].EndTime)
	}
	if len(payload.Data.Courses) != 2 {
		t.Fatalf("courses = %+v", payload.Data.Courses)
	}
	if got := payload.Data.Days; fmt.Sprint(got) != "[MONDAY WEDNESDAY]" {
		t.Fatalf("days derivados = %v", got)
	}
	var theoryGroups = payload.Data.Courses[0].Groups
	if payload.Data.Courses[0].Type != "THEORY" {
		theoryGroups = payload.Data.Courses[1].Groups
	}
	if len(theoryGroups) != 1 {
		t.Fatalf("theory groups = %+v", theoryGroups)
	}
	group := theoryGroups[0]
	if group.Name != "A" || len(group.Sessions) != 2 {
		t.Fatalf("group = %+v", group)
	}

	// Compartir: flujo completo ida y vuelta.
	response = doJSON(handler, "POST", "/shared", `{"selection":{"1":1}}`, nil)
	if response.Code != http.StatusCreated {
		t.Fatalf("create share: %d %s", response.Code, response.Body.String())
	}
	var sharedPayload struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &sharedPayload); err != nil {
		t.Fatal(err)
	}
	shareID := sharedPayload.Data.ID
	if len(shareID) != 10 {
		t.Fatalf("share id = %q", shareID)
	}

	response = doJSON(handler, "GET", "/shared/"+shareID, "", nil)
	if response.Code != http.StatusOK {
		t.Fatalf("get share: %d %s", response.Code, response.Body.String())
	}
	var sharedView struct {
		Data struct {
			Selection map[string]int `json:"selection"`
		} `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &sharedView); err != nil {
		t.Fatal(err)
	}
	if sharedView.Data.Selection["1"] != 1 {
		t.Fatalf("selection = %v", sharedView.Data.Selection)
	}
	response = doJSON(handler, "GET", "/shared/ZZZZZZZZZZ", "", nil)
	if response.Code != http.StatusNotFound {
		t.Fatalf("unknown share: %d, want 404", response.Code)
	}
	response = doJSON(handler, "POST", "/shared", `{"selection":{"1":"hack"}}`, nil)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("bad share payload: %d, want 400", response.Code)
	}

	// Eliminaciones.
	response = doJSON(handler, "DELETE", "/groups/1", "", cookie)
	if response.Code != http.StatusNoContent {
		t.Fatalf("delete group: %d", response.Code)
	}
	response = doJSON(handler, "DELETE", "/groups/1", "", cookie)
	if response.Code != http.StatusNotFound {
		t.Fatalf("delete missing group: %d, want 404", response.Code)
	}
	response = doJSON(handler, "DELETE", "/teachers/1", "", cookie)
	if response.Code != http.StatusConflict {
		t.Fatalf("teacher in use: %d, want 409", response.Code)
	}
	response = doJSON(handler, "DELETE", "/classrooms/1", "", cookie)
	if response.Code != http.StatusNoContent {
		t.Fatalf("delete classroom: %d %s", response.Code, response.Body.String())
	}
	response = doJSON(handler, "DELETE", "/courses/1", "", cookie)
	if response.Code != http.StatusConflict {
		t.Fatalf("course with lab dependent: %d, want 409", response.Code)
	}
	response = doJSON(handler, "DELETE", "/courses/2", "", cookie)
	if response.Code != http.StatusNoContent {
		t.Fatalf("delete lab course: %d", response.Code)
	}

	// Un rol ADMIN obsoleto en cookie no autoriza a un correo sin registro.
	ephemeral := authCookie(t, auth, AuthUser{ID: 1, Email: "ghost@unsa.edu.pe", Role: "ADMIN", EmailVerified: true})
	response = doJSON(handler, "POST", "/classrooms", `{"code":"Aula 1000","type":"THEORY"}`, ephemeral)
	if response.Code != http.StatusForbidden {
		t.Fatalf("ephemeral user: %d, want 403", response.Code)
	}
}
