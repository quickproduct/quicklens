package handlers

import (
	"github.com/go-chi/chi/v5"
	"database/sql"
	"fmt"
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"time"

	"quicklens/backend/internal/db"
	"quicklens/backend/internal/models"
)

func ListTracesHandler(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	offset := (page - 1) * perPage

	// Build filters
	where := "WHERE 1=1"
	args := make([]interface{}, 0)

	if model := r.URL.Query().Get("model"); model != "" {
		where += " AND t.id IN (SELECT DISTINCT trace_id FROM spans WHERE model_id = ?)"
		args = append(args, model)
	}
	if status := r.URL.Query().Get("status"); status != "" {
		where += " AND t.status = ?"
		args = append(args, status)
	}
	if from := r.URL.Query().Get("from"); from != "" {
		where += " AND t.created_at >= ?"
		args = append(args, from)
	}
	if to := r.URL.Query().Get("to"); to != "" {
		where += " AND t.created_at <= ?"
		args = append(args, to)
	}
	if sessionID := r.URL.Query().Get("session_id"); sessionID != "" {
		where += " AND t.session_id = ?"
		args = append(args, sessionID)
	}
	if q := r.URL.Query().Get("q"); q != "" {
		where += " AND t.name LIKE ?"
		args = append(args, "%"+q+"%")
	}

	// Get total count
	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM traces t %s", where)
	db.DB.QueryRow(countQuery, args...).Scan(&total)

	// Get traces
	query := fmt.Sprintf(`
		SELECT t.id, t.trace_id, t.session_id, t.name, t.status, t.total_duration_ms,
		       t.total_tokens, t.prompt_tokens, t.completion_tokens, t.total_cost,
		       t.input_preview, t.output_preview, t.created_at,
		       (SELECT COUNT(*) FROM spans WHERE trace_id = t.id) as span_count
		FROM traces t
		%s
		ORDER BY t.created_at DESC
		LIMIT ? OFFSET ?
	`, where)
	args = append(args, perPage, offset)

	rows, err := db.DB.Query(query, args...)
	if err != nil {
		zap.L().Sugar().Infof("Error querying traces: %v", err)
		WriteError(w, http.StatusInternalServerError, "Failed to query traces")
		return
	}
	defer rows.Close()

	traces := make([]models.TraceResponse, 0)
	for rows.Next() {
		var tr models.TraceResponse
		var createdAt time.Time
		if err := rows.Scan(
			&tr.ID, &tr.TraceID, &tr.SessionID, &tr.Name, &tr.Status,
			&tr.TotalDurationMs, &tr.TotalTokens, &tr.PromptTokens,
			&tr.CompletionTokens, &tr.TotalCost, &tr.InputPreview,
			&tr.OutputPreview, &createdAt, &tr.SpanCount,
		); err != nil {
			zap.L().Sugar().Infof("Error scanning trace: %v", err)
			continue
		}
		tr.CreatedAt = &createdAt

		// Get model info from first LLM span
		var modelName, provider string
		db.DB.QueryRow(
			"SELECT model_id, provider FROM spans WHERE trace_id = ? AND type = 'llm' LIMIT 1",
			tr.ID,
		).Scan(&modelName, &provider)
		tr.ModelName = modelName
		tr.Provider = provider

		// Populate alias fields
		tr.DurationMs = tr.TotalDurationMs
		tr.Cost = tr.TotalCost
		tr.Model = tr.ModelName

		traces = append(traces, tr)
	}

	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"items":    traces,
		"total":    total,
		"page":     page,
		"per_page": perPage,
		"pages":    (total + perPage - 1) / perPage,
	})
}

func GetTraceHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		WriteError(w, http.StatusBadRequest, "Trace ID is required")
		return
	}

	var tr models.TraceDetailResponse
	var createdAt time.Time
	err := db.DB.QueryRow(`
		SELECT t.id, t.trace_id, t.session_id, t.name, t.status, t.total_duration_ms,
		       t.total_tokens, t.prompt_tokens, t.completion_tokens, t.total_cost,
		       t.input_preview, t.output_preview, t.created_at,
		       (SELECT COUNT(*) FROM spans WHERE trace_id = t.id) as span_count
		FROM traces t
		WHERE t.id = ?
	`, id).Scan(
		&tr.ID, &tr.TraceID, &tr.SessionID, &tr.Name, &tr.Status,
		&tr.TotalDurationMs, &tr.TotalTokens, &tr.PromptTokens,
		&tr.CompletionTokens, &tr.TotalCost, &tr.InputPreview,
		&tr.OutputPreview, &createdAt, &tr.SpanCount,
	)
	if err != nil {
		WriteError(w, http.StatusNotFound, "Trace not found")
		return
	}
	tr.CreatedAt = &createdAt

	// Get model info
	var modelName, provider string
	db.DB.QueryRow(
		"SELECT model_id, provider FROM spans WHERE trace_id = ? AND type = 'llm' LIMIT 1",
		tr.ID,
	).Scan(&modelName, &provider)
	tr.ModelName = modelName
	tr.Provider = provider

	// Populate alias fields
	tr.DurationMs = tr.TotalDurationMs
	tr.Cost = tr.TotalCost
	tr.Model = tr.ModelName

	// Get all spans for this trace
	spanRows, err := db.DB.Query(`
		SELECT id, trace_id, parent_span_id, name, type, model_id, provider,
		       input, output, prompt_tokens, completion_tokens, total_tokens,
		       cost, duration_ms, status, error_message, metadata, started_at, ended_at
		FROM spans
		WHERE trace_id = ?
		ORDER BY started_at ASC
	`, id)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to query spans")
		return
	}
	defer spanRows.Close()

	allSpans := make([]models.SpanResponse, 0)
	for spanRows.Next() {
		var s models.SpanResponse
		var metadataStr string
		var startedAt, endedAt sql.NullTime
		if err := spanRows.Scan(
			&s.ID, &s.TraceID, &s.ParentSpanID, &s.Name, &s.Type, &s.ModelID, &s.Provider,
			&s.Input, &s.Output, &s.PromptTokens, &s.CompletionTokens, &s.TotalTokens,
			&s.Cost, &s.DurationMs, &s.Status, &s.ErrorMessage, &metadataStr, &startedAt, &endedAt,
		); err != nil {
			zap.L().Sugar().Infof("Error scanning span: %v", err)
			continue
		}
		if startedAt.Valid {
			s.StartedAt = &startedAt.Time
		}
		if endedAt.Valid {
			s.EndedAt = &endedAt.Time
		}
		s.Children = make([]models.SpanResponse, 0)
		allSpans = append(allSpans, s)
	}

	// Build tree structure
	tr.Spans = buildSpanTree(allSpans)

	WriteJSON(w, http.StatusOK, tr)
}

func buildSpanTree(spans []models.SpanResponse) []models.SpanResponse {
	spanMap := make(map[string]*models.SpanResponse)
	for i := range spans {
		spans[i].Children = make([]models.SpanResponse, 0)
		spanMap[spans[i].ID] = &spans[i]
	}

	roots := make([]models.SpanResponse, 0)
	for i := range spans {
		if spans[i].ParentSpanID == "" {
			roots = append(roots, spans[i])
		} else if parent, ok := spanMap[spans[i].ParentSpanID]; ok {
			parent.Children = append(parent.Children, spans[i])
		} else {
			roots = append(roots, spans[i])
		}
	}

	// Copy children from map back to roots
	var copyChildren func(span *models.SpanResponse)
	copyChildren = func(span *models.SpanResponse) {
		if mapped, ok := spanMap[span.ID]; ok {
			span.Children = mapped.Children
			for i := range span.Children {
				copyChildren(&span.Children[i])
			}
		}
	}
	for i := range roots {
		copyChildren(&roots[i])
	}

	return roots
}

func DeleteTraceHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		WriteError(w, http.StatusBadRequest, "Trace ID is required")
		return
	}

	// Delete spans first (cascade)
	_, _ = db.DB.Exec("DELETE FROM spans WHERE trace_id = ?", id)

	result, err := db.DB.Exec("DELETE FROM traces WHERE id = ?", id)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to delete trace")
		return
	}

	affected, _ := result.RowsAffected()
	if affected == 0 {
		WriteError(w, http.StatusNotFound, "Trace not found")
		return
	}

	WriteJSON(w, http.StatusOK, map[string]string{"message": "Trace deleted"})
}

func ListSessionsHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.DB.Query(`
		SELECT ls.id, ls.name, ls.created_at,
		       (SELECT COUNT(*) FROM traces WHERE session_id = ls.id) as trace_count
		FROM llm_sessions ls
		ORDER BY ls.created_at DESC
	`)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to query sessions")
		return
	}
	defer rows.Close()

	sessions := make([]models.SessionResponse, 0)
	for rows.Next() {
		var s models.SessionResponse
		var createdAt time.Time
		if err := rows.Scan(&s.ID, &s.Name, &createdAt, &s.TraceCount); err != nil {
			zap.L().Sugar().Infof("Error scanning session: %v", err)
			continue
		}
		s.CreatedAt = &createdAt
		sessions = append(sessions, s)
	}

	WriteJSON(w, http.StatusOK, sessions)
}

func GetSessionTracesHandler(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "id")
	if sessionID == "" {
		WriteError(w, http.StatusBadRequest, "Session ID is required")
		return
	}

	rows, err := db.DB.Query(`
		SELECT t.id, t.trace_id, t.session_id, t.name, t.status, t.total_duration_ms,
		       t.total_tokens, t.prompt_tokens, t.completion_tokens, t.total_cost,
		       t.input_preview, t.output_preview, t.created_at,
		       (SELECT COUNT(*) FROM spans WHERE trace_id = t.id) as span_count
		FROM traces t
		WHERE t.session_id = ?
		ORDER BY t.created_at DESC
	`, sessionID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to query session traces")
		return
	}
	defer rows.Close()

	traces := make([]models.TraceResponse, 0)
	for rows.Next() {
		var tr models.TraceResponse
		var createdAt time.Time
		if err := rows.Scan(
			&tr.ID, &tr.TraceID, &tr.SessionID, &tr.Name, &tr.Status,
			&tr.TotalDurationMs, &tr.TotalTokens, &tr.PromptTokens,
			&tr.CompletionTokens, &tr.TotalCost, &tr.InputPreview,
			&tr.OutputPreview, &createdAt, &tr.SpanCount,
		); err != nil {
			continue
		}
		tr.CreatedAt = &createdAt

		var modelName, prov string
		db.DB.QueryRow(
			"SELECT model_id, provider FROM spans WHERE trace_id = ? AND type = 'llm' LIMIT 1",
			tr.ID,
		).Scan(&modelName, &prov)
		tr.ModelName = modelName
		tr.Provider = prov

		// Populate alias fields
		tr.DurationMs = tr.TotalDurationMs
		tr.Cost = tr.TotalCost
		tr.Model = tr.ModelName

		traces = append(traces, tr)
	}

	WriteJSON(w, http.StatusOK, traces)
}
