package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"quicklens/backend/db"
	"quicklens/backend/models"
)

func SearchLogsHandler(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 50
	}
	offset := (page - 1) * perPage

	// Filters
	where := "WHERE type = 'llm'"
	args := make([]interface{}, 0)

	if model := r.URL.Query().Get("model"); model != "" {
		where += " AND model_id = ?"
		args = append(args, model)
	}
	if status := r.URL.Query().Get("status"); status != "" {
		where += " AND status = ?"
		args = append(args, status)
	}
	if minLatencyStr := r.URL.Query().Get("min_latency"); minLatencyStr != "" {
		if minLatency, err := strconv.ParseInt(minLatencyStr, 10, 64); err == nil {
			where += " AND duration_ms >= ?"
			args = append(args, minLatency)
		}
	}
	if maxLatencyStr := r.URL.Query().Get("max_latency"); maxLatencyStr != "" {
		if maxLatency, err := strconv.ParseInt(maxLatencyStr, 10, 64); err == nil {
			where += " AND duration_ms <= ?"
			args = append(args, maxLatency)
		}
	}
	if from := r.URL.Query().Get("from"); from != "" {
		where += " AND started_at >= ?"
		args = append(args, from)
	}
	if to := r.URL.Query().Get("to"); to != "" {
		where += " AND started_at <= ?"
		args = append(args, to)
	}

	query := fmt.Sprintf(`
		SELECT id, trace_id, model_id, provider, status, duration_ms,
		       prompt_tokens, completion_tokens, input, output, error_message, started_at
		FROM spans
		%s
		ORDER BY started_at DESC
		LIMIT ? OFFSET ?
	`, where)

	// Append pagination args
	queryArgs := append(args, perPage, offset)

	rows, err := db.DB.Query(query, queryArgs...)
	if err != nil {
		log.Printf("Failed to query log spans: %v", err)
		WriteError(w, http.StatusInternalServerError, "Failed to query logs")
		return
	}
	defer rows.Close()

	result := make([]models.LogEntry, 0)
	for rows.Next() {
		var entry models.LogEntry
		var input, output string
		var startedAt time.Time

		err := rows.Scan(
			&entry.ID, &entry.TraceID, &entry.ModelName, &entry.Provider, &entry.Status, &entry.DurationMs,
			&entry.PromptTokens, &entry.CompletionTokens, &input, &output, &entry.ErrorMessage, &startedAt,
		)
		if err != nil {
			log.Printf("Failed to scan log span: %v", err)
			continue
		}
		entry.SpanID = entry.ID
		entry.InputPreview = truncate(input, 200)
		entry.OutputPreview = truncate(output, 200)
		entry.CreatedAt = &startedAt
		result = append(result, entry)
	}

	WriteJSON(w, http.StatusOK, result)
}
