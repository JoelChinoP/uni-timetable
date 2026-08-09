package main

import (
	"context"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/JoelChinoP/uni-timetable/backend/internal/database"
)

const defaultCareerCode = "SIS"

var plannerDayOrder = []string{"MONDAY", "TUESDAY", "WEDNESDAY", "THURSDAY", "FRIDAY", "SATURDAY"}

type plannerHandler struct {
	queries   *database.Queries
	termLabel string
}

type academicHourJSON struct {
	HourNumber int16  `json:"hourNumber"`
	StartTime  string `json:"startTime"`
	EndTime    string `json:"endTime"`
}

type teacherJSON struct {
	ID       int32  `json:"id"`
	FullName string `json:"fullName"`
}

type sessionJSON struct {
	ID                int32  `json:"id"`
	Title             string `json:"title"`
	Day               string `json:"day"`
	StartHourAcademic int16  `json:"startHourAcademic"`
	DurationHours     int16  `json:"durationHours"`
	Mode              string `json:"mode"`
	ClassroomLabel    string `json:"classroomLabel"`
}

type groupJSON struct {
	ID             int32         `json:"id"`
	Name           string        `json:"name"`
	ClassroomID    *int32        `json:"classroomId"`
	ClassroomLabel string        `json:"classroomLabel"`
	Sessions       []sessionJSON `json:"sessions"`
}

type courseJSON struct {
	ID           int32        `json:"id"`
	Code         string       `json:"code"`
	Name         string       `json:"name"`
	Abbreviation string       `json:"abbreviation"`
	Summary      string       `json:"summary"`
	Credits      *int16       `json:"credits"`
	Color        string       `json:"color"`
	Type         string       `json:"type"`
	AcademicYear int16        `json:"academicYear"`
	TheoryID     *int32       `json:"theoryCourseId"`
	Teacher      *teacherJSON `json:"teacher"`
	Groups       []groupJSON  `json:"groups"`
}

type dashboardJSON struct {
	TermLabel     string             `json:"termLabel"`
	Days          []string           `json:"days"`
	AcademicHours []academicHourJSON `json:"academicHours"`
	Courses       []courseJSON       `json:"courses"`
}

func (handler *plannerHandler) dashboard(w http.ResponseWriter, r *http.Request) {
	if handler.queries == nil {
		writeError(w, http.StatusServiceUnavailable, "database is not configured")
		return
	}
	career := r.URL.Query().Get("career")
	if career == "" {
		career = defaultCareerCode
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	hours, err := handler.queries.ListAcademicHours(ctx)
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, "database unavailable")
		return
	}
	courseRows, err := handler.queries.ListPlannerCourses(ctx, career)
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, "database unavailable")
		return
	}
	groupRows, err := handler.queries.ListPlannerGroups(ctx, career)
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, "database unavailable")
		return
	}
	sessionRows, err := handler.queries.ListPlannerSessions(ctx, career)
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, "database unavailable")
		return
	}

	classroomByGroup := make(map[int32]string, len(groupRows))
	groupsByCourse := make(map[int32][]groupJSON)
	for _, row := range groupRows {
		label := ""
		if row.ClassroomCode.Valid {
			label = row.ClassroomCode.String
		}
		classroomByGroup[row.ID] = label
		group := groupJSON{
			ID: row.ID, Name: row.Name, ClassroomLabel: label, Sessions: []sessionJSON{},
		}
		if row.IDClassroom.Valid {
			group.ClassroomID = &row.IDClassroom.Int32
		}
		groupsByCourse[row.IDCourse] = append(groupsByCourse[row.IDCourse], group)
	}

	sessionsByGroup := make(map[int32][]sessionJSON)
	daysSeen := map[string]bool{}
	for _, row := range sessionRows {
		daysSeen[string(row.Day)] = true
		sessionsByGroup[row.IDGroup] = append(sessionsByGroup[row.IDGroup], sessionJSON{
			ID:                row.ID,
			Title:             row.CourseName,
			Day:               string(row.Day),
			StartHourAcademic: row.StartHourAcademic,
			DurationHours:     row.DurationHours,
			Mode:              string(row.Type),
			ClassroomLabel:    classroomByGroup[row.IDGroup],
		})
	}

	courses := make([]courseJSON, 0, len(courseRows))
	for _, row := range courseRows {
		course := courseJSON{
			ID:           row.ID,
			Code:         row.Code,
			Name:         row.Name,
			Abbreviation: row.Abbreviation,
			Summary:      row.Summary,
			Color:        row.Color,
			Type:         string(row.Type),
			AcademicYear: row.AcademicYear,
			Groups:       []groupJSON{},
		}
		if row.Credits.Valid {
			course.Credits = &row.Credits.Int16
		}
		if row.IDCourseTheory.Valid {
			course.TheoryID = &row.IDCourseTheory.Int32
		}
		if row.TeacherID.Valid {
			course.Teacher = &teacherJSON{ID: row.TeacherID.Int32, FullName: row.TeacherName}
		}
		for _, group := range groupsByCourse[row.ID] {
			group.Sessions = sessionsByGroup[group.ID]
			if group.Sessions == nil {
				group.Sessions = []sessionJSON{}
			}
			course.Groups = append(course.Groups, group)
		}
		courses = append(courses, course)
	}

	days := make([]string, 0, len(plannerDayOrder))
	for _, day := range plannerDayOrder {
		if daysSeen[day] {
			days = append(days, day)
		}
	}
	if len(days) == 0 {
		days = plannerDayOrder[:5]
	}

	writeJSON(w, http.StatusOK, dashboardJSON{
		TermLabel:     handler.termLabel,
		Days:          days,
		AcademicHours: mapAcademicHours(hours),
		Courses:       courses,
	})
}

func mapAcademicHours(rows []database.AppAcademicHour) []academicHourJSON {
	hours := make([]academicHourJSON, 0, len(rows))
	for _, row := range rows {
		hours = append(hours, academicHourJSON{
			HourNumber: row.HourNumber,
			StartTime:  formatPGTime(row.StartTime),
			EndTime:    formatPGTime(row.EndTime),
		})
	}
	return hours
}

func formatPGTime(value pgtype.Time) string {
	return time.UnixMicro(value.Microseconds).UTC().Format("15:04")
}
