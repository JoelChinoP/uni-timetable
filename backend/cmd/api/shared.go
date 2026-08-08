package main

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/JoelChinoP/uni-timetable/backend/internal/database"
)

const (
	shareIDLength = 10
	maxShareItems = 100
)

// ponytail: base62 without lookalikes handled by alphabet choice? No — full base62, randomness is what matters.
const shareAlphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"

var shareIDPattern = regexp.MustCompile(`^[A-Za-z0-9]{10}$`)

type sharedHandler struct {
	queries *database.Queries
}

func generateShareID() (string, error) {
	buffer := make([]byte, shareIDLength)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}
	for index, value := range buffer {
		buffer[index] = shareAlphabet[int(value)%len(shareAlphabet)]
	}
	return string(buffer), nil
}

func normalizeSelection(raw map[string]json.RawMessage) ([]byte, error) {
	if len(raw) == 0 || len(raw) > maxShareItems {
		return nil, fmt.Errorf("selection debe tener entre 1 y %d elementos", maxShareItems)
	}
	selection := make(map[int32]int32, len(raw))
	for key, value := range raw {
		courseID, err := strconv.ParseInt(key, 10, 32)
		if err != nil || courseID <= 0 {
			return nil, fmt.Errorf("selection contiene un curso inválido")
		}
		var groupID int32
		if err := json.Unmarshal(value, &groupID); err != nil || groupID <= 0 {
			return nil, fmt.Errorf("selection contiene un grupo inválido")
		}
		selection[int32(courseID)] = groupID
	}
	return json.Marshal(selection)
}

func (handler *sharedHandler) shared(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		handler.createShared(w, r)
	default:
		w.Header().Set("Allow", "POST")
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (handler *sharedHandler) createShared(w http.ResponseWriter, r *http.Request) {
	if handler.queries == nil {
		writeError(w, http.StatusServiceUnavailable, "database is not configured")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 16<<10)
	var request struct {
		Selection map[string]json.RawMessage `json:"selection"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid share payload")
		return
	}

	selection, err := normalizeSelection(request.Selection)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	var pairs map[int32]int32
	if err := json.Unmarshal(selection, &pairs); err != nil {
		writeError(w, http.StatusBadRequest, "invalid share payload")
		return
	}
	for courseID, groupID := range pairs {
		valid, err := handler.queries.GroupBelongsToCourse(ctx, database.GroupBelongsToCourseParams{
			ID: groupID, IDCourse: courseID,
		})
		if err != nil {
			writeError(w, http.StatusServiceUnavailable, "database unavailable")
			return
		}
		if !valid {
			writeError(w, http.StatusBadRequest, "selection contiene un grupo ajeno al curso")
			return
		}
	}

	// ponytail: collision probability is ~0; one retry covers the freak case.
	for range 2 {
		id, err := generateShareID()
		if err != nil {
			writeError(w, http.StatusInternalServerError, "share unavailable")
			return
		}
		err = handler.queries.CreateSharedTimetable(ctx, database.CreateSharedTimetableParams{
			ID:      id,
			Column2: string(selection),
		})
		if err == nil {
			writeJSON(w, http.StatusCreated, map[string]string{"id": id})
			return
		}
		var pgError *pgconn.PgError
		if !errors.As(err, &pgError) || pgError.Code != "23505" {
			writeError(w, http.StatusServiceUnavailable, "database unavailable")
			return
		}
	}
	writeError(w, http.StatusInternalServerError, "share unavailable")
}

func (handler *sharedHandler) getShared(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if !shareIDPattern.MatchString(id) {
		writeError(w, http.StatusNotFound, "enlace no encontrado")
		return
	}
	if handler.queries == nil {
		writeError(w, http.StatusServiceUnavailable, "database is not configured")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	selection, err := handler.queries.GetSharedTimetable(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "enlace no encontrado")
			return
		}
		writeError(w, http.StatusServiceUnavailable, "database unavailable")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	// selection already is a validated JSON object; echo it verbatim.
	fmt.Fprintf(w, "{\"data\":{\"id\":%q,\"selection\":%s}}\n", id, selection)
}
