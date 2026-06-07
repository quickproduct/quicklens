package handlers

import (
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"

	"quicklens/backend/db"
	"quicklens/backend/models"
)

func CreateEvaluationHandler(w http.ResponseWriter, r *http.Request) {
	var req models.EvalCreateRequest
	if err := ParseJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.SpanID == "" || req.ScoreType == "" {
		WriteError(w, http.StatusBadRequest, "SpanID and ScoreType are required")
		return
	}

	// Verify span exists
	var exists int
	err := db.DB.QueryRow("SELECT COUNT(*) FROM spans WHERE id = ?", req.SpanID).Scan(&exists)
	if err != nil || exists == 0 {
		WriteError(w, http.StatusNotFound, "Span not found")
		return
	}

	id := uuid.New().String()
	now := time.Now().UTC()

	_, err = db.DB.Exec(`
		INSERT INTO evaluations (id, span_id, score_type, score_value, feedback_text, evaluator, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, id, req.SpanID, req.ScoreType, req.ScoreValue, req.FeedbackText, req.Evaluator, now)
	if err != nil {
		log.Printf("Failed to insert evaluation: %v", err)
		WriteError(w, http.StatusInternalServerError, "Failed to store evaluation")
		return
	}

	resp := models.EvalResponse{
		ID:           id,
		SpanID:       req.SpanID,
		ScoreType:    req.ScoreType,
		ScoreValue:   req.ScoreValue,
		FeedbackText: req.FeedbackText,
		Evaluator:    req.Evaluator,
		CreatedAt:    &now,
	}

	WriteJSON(w, http.StatusCreated, resp)
}

func ListEvaluationsHandler(w http.ResponseWriter, r *http.Request) {
	query := `
		SELECT id, span_id, score_type, score_value, feedback_text, evaluator, created_at
		FROM evaluations
		ORDER BY created_at DESC
	`
	rows, err := db.DB.Query(query)
	if err != nil {
		log.Printf("Failed to query evaluations: %v", err)
		WriteError(w, http.StatusInternalServerError, "Failed to query evaluations")
		return
	}
	defer rows.Close()

	result := make([]models.EvalResponse, 0)
	for rows.Next() {
		var ev models.EvalResponse
		var createdAt time.Time
		err := rows.Scan(
			&ev.ID, &ev.SpanID, &ev.ScoreType, &ev.ScoreValue, &ev.FeedbackText, &ev.Evaluator, &createdAt,
		)
		if err != nil {
			log.Printf("Failed to scan evaluation: %v", err)
			continue
		}
		ev.CreatedAt = &createdAt
		result = append(result, ev)
	}

	WriteJSON(w, http.StatusOK, result)
}

func GetSpanEvaluationsHandler(w http.ResponseWriter, r *http.Request) {
	spanID := r.PathValue("id")
	if spanID == "" {
		WriteError(w, http.StatusBadRequest, "Span ID is required")
		return
	}

	rows, err := db.DB.Query(`
		SELECT id, span_id, score_type, score_value, feedback_text, evaluator, created_at
		FROM evaluations
		WHERE span_id = ?
		ORDER BY created_at DESC
	`, spanID)
	if err != nil {
		log.Printf("Failed to query span evaluations: %v", err)
		WriteError(w, http.StatusInternalServerError, "Failed to query evaluations")
		return
	}
	defer rows.Close()

	result := make([]models.EvalResponse, 0)
	for rows.Next() {
		var ev models.EvalResponse
		var createdAt time.Time
		err := rows.Scan(
			&ev.ID, &ev.SpanID, &ev.ScoreType, &ev.ScoreValue, &ev.FeedbackText, &ev.Evaluator, &createdAt,
		)
		if err != nil {
			log.Printf("Failed to scan evaluation: %v", err)
			continue
		}
		ev.CreatedAt = &createdAt
		result = append(result, ev)
	}

	WriteJSON(w, http.StatusOK, result)
}
