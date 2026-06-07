package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"

	"quicklens/backend/db"
	"quicklens/backend/models"
	"quicklens/backend/ws"
)

var ingestChan = make(chan models.TraceIngestRequest, 100)

func StartIngestWorker() {
	go func() {
		log.Println("Ingest worker started")
		for req := range ingestChan {
			persistTrace(req)
		}
	}()
}

func persistTrace(req models.TraceIngestRequest) {
	traceID := req.TraceID
	if traceID == "" {
		traceID = uuid.New().String()
	}

	id := uuid.New().String()
	now := time.Now().UTC()

	// Compute aggregates from spans
	var totalTokens, promptTokens, completionTokens int64
	var totalDurationMs int64
	var totalCost float64
	traceStatus := "ok"
	var inputPreview, outputPreview string

	for i, s := range req.Spans {
		sTotalTokens := s.PromptTokens + s.CompletionTokens
		totalTokens += sTotalTokens
		promptTokens += s.PromptTokens
		completionTokens += s.CompletionTokens
		totalDurationMs += s.DurationMs

		if s.Status == "error" {
			traceStatus = "error"
		}

		// Compute cost from model_prices
		var promptPrice, completionPrice float64
		db.DB.QueryRow(
			"SELECT prompt_price_per_1k, completion_price_per_1k FROM model_prices WHERE provider = ? AND model_id = ?",
			s.Provider, s.ModelID,
		).Scan(&promptPrice, &completionPrice)

		spanCost := (float64(s.PromptTokens) / 1000.0 * promptPrice) +
			(float64(s.CompletionTokens) / 1000.0 * completionPrice)
		totalCost += spanCost

		// Get input/output preview from first span
		if i == 0 {
			inputPreview = truncate(s.Input, 200)
			outputPreview = truncate(s.Output, 200)
		}
	}

	// Ensure session exists if provided
	if req.SessionID != "" {
		var exists int
		err := db.DB.QueryRow("SELECT COUNT(*) FROM llm_sessions WHERE id = ?", req.SessionID).Scan(&exists)
		if err == nil && exists == 0 {
			_, _ = db.DB.Exec(
				"INSERT INTO llm_sessions (id, name, metadata, created_at) VALUES (?, ?, '{}', ?)",
				req.SessionID, req.Name, now,
			)
		}
	}

	// Insert trace
	_, err := db.DB.Exec(
		`INSERT INTO traces (id, trace_id, session_id, name, status, total_duration_ms,
		 total_tokens, prompt_tokens, completion_tokens, total_cost, input_preview, output_preview, metadata, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, '{}', ?)`,
		id, traceID, req.SessionID, req.Name, traceStatus, totalDurationMs,
		totalTokens, promptTokens, completionTokens, totalCost, inputPreview, outputPreview, now,
	)
	if err != nil {
		log.Printf("Failed to insert trace: %v", err)
		return
	}

	// Insert spans
	for _, s := range req.Spans {
		spanID := s.SpanID
		if spanID == "" {
			spanID = uuid.New().String()
		}

		sTotalTokens := s.PromptTokens + s.CompletionTokens

		// Compute span cost
		var promptPrice, completionPrice float64
		db.DB.QueryRow(
			"SELECT prompt_price_per_1k, completion_price_per_1k FROM model_prices WHERE provider = ? AND model_id = ?",
			s.Provider, s.ModelID,
		).Scan(&promptPrice, &completionPrice)
		spanCost := (float64(s.PromptTokens) / 1000.0 * promptPrice) +
			(float64(s.CompletionTokens) / 1000.0 * completionPrice)

		metadataJSON := "{}"
		if s.Metadata != nil {
			if b, err := json.Marshal(s.Metadata); err == nil {
				metadataJSON = string(b)
			}
		}

		startedAt := s.StartedAt
		endedAt := s.EndedAt
		if startedAt == nil {
			n := now
			startedAt = &n
		}
		if endedAt == nil {
			n := now.Add(time.Duration(s.DurationMs) * time.Millisecond)
			endedAt = &n
		}

		_, err := db.DB.Exec(
			`INSERT INTO spans (id, trace_id, parent_span_id, name, type, model_id, provider,
			 input, output, prompt_tokens, completion_tokens, total_tokens, cost, duration_ms,
			 status, error_message, metadata, started_at, ended_at)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			spanID, id, s.ParentSpanID, s.Name, s.Type, s.ModelID, s.Provider,
			s.Input, s.Output, s.PromptTokens, s.CompletionTokens, sTotalTokens, spanCost, s.DurationMs,
			s.Status, s.ErrorMessage, metadataJSON, startedAt, endedAt,
		)
		if err != nil {
			log.Printf("Failed to insert span: %v", err)
		}
	}

	// Broadcast new trace via WebSocket
	ws.Manager.Broadcast("traces", map[string]interface{}{
		"type":    "new_trace",
		"traceId": id,
		"name":    req.Name,
		"status":  traceStatus,
		"tokens":  totalTokens,
		"cost":    totalCost,
	})

	log.Printf("Ingested trace %s with %d spans", traceID, len(req.Spans))
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

func IngestTraceHandler(w http.ResponseWriter, r *http.Request) {
	var req models.TraceIngestRequest
	if err := ParseJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Name == "" {
		req.Name = "Unnamed trace"
	}

	select {
	case ingestChan <- req:
		WriteJSON(w, http.StatusAccepted, map[string]string{
			"message":  "Trace accepted for processing",
			"trace_id": req.TraceID,
		})
	default:
		WriteError(w, http.StatusServiceUnavailable, "Ingest queue is full, try again later")
	}
}

func IngestOTelHandler(w http.ResponseWriter, r *http.Request) {
	// Parse OTLP JSON format
	var otlpPayload struct {
		ResourceSpans []struct {
			Resource struct {
				Attributes []struct {
					Key   string      `json:"key"`
					Value interface{} `json:"value"`
				} `json:"attributes"`
			} `json:"resource"`
			ScopeSpans []struct {
				Spans []struct {
					TraceID           string `json:"traceId"`
					SpanID            string `json:"spanId"`
					ParentSpanID      string `json:"parentSpanId"`
					Name              string `json:"name"`
					Kind              int    `json:"kind"`
					StartTimeUnixNano string `json:"startTimeUnixNano"`
					EndTimeUnixNano   string `json:"endTimeUnixNano"`
					Status            struct {
						Code    int    `json:"code"`
						Message string `json:"message"`
					} `json:"status"`
					Attributes []struct {
						Key   string      `json:"key"`
						Value interface{} `json:"value"`
					} `json:"attributes"`
				} `json:"spans"`
			} `json:"scopeSpans"`
		} `json:"resourceSpans"`
	}

	if err := ParseJSON(r, &otlpPayload); err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid OTLP JSON payload")
		return
	}

	for _, rs := range otlpPayload.ResourceSpans {
		for _, ss := range rs.ScopeSpans {
			if len(ss.Spans) == 0 {
				continue
			}

			traceIDStr := ss.Spans[0].TraceID
			ingestReq := models.TraceIngestRequest{
				TraceID: traceIDStr,
				Name:    ss.Spans[0].Name,
				Spans:   make([]models.SpanIngestRequest, 0, len(ss.Spans)),
			}

			for _, otelSpan := range ss.Spans {
				status := "ok"
				errMsg := ""
				if otelSpan.Status.Code == 2 {
					status = "error"
					errMsg = otelSpan.Status.Message
				}

				// Extract attributes
				var modelID, provider, input, output, spanType string
				var promptTokens, completionTokens int64
				spanType = "llm"

				for _, attr := range otelSpan.Attributes {
					switch attr.Key {
					case "llm.model", "gen_ai.request.model":
						if v, ok := extractStringValue(attr.Value); ok {
							modelID = v
						}
					case "llm.provider", "gen_ai.system":
						if v, ok := extractStringValue(attr.Value); ok {
							provider = v
						}
					case "llm.input", "gen_ai.prompt":
						if v, ok := extractStringValue(attr.Value); ok {
							input = v
						}
					case "llm.output", "gen_ai.completion":
						if v, ok := extractStringValue(attr.Value); ok {
							output = v
						}
					case "llm.usage.prompt_tokens", "gen_ai.usage.prompt_tokens":
						if v, ok := extractIntValue(attr.Value); ok {
							promptTokens = v
						}
					case "llm.usage.completion_tokens", "gen_ai.usage.completion_tokens":
						if v, ok := extractIntValue(attr.Value); ok {
							completionTokens = v
						}
					case "span.type":
						if v, ok := extractStringValue(attr.Value); ok {
							spanType = v
						}
					}
				}

				span := models.SpanIngestRequest{
					SpanID:           otelSpan.SpanID,
					ParentSpanID:     otelSpan.ParentSpanID,
					Name:             otelSpan.Name,
					Type:             spanType,
					ModelID:          modelID,
					Provider:         provider,
					Input:            input,
					Output:           output,
					PromptTokens:     promptTokens,
					CompletionTokens: completionTokens,
					Status:           status,
					ErrorMessage:     errMsg,
				}
				ingestReq.Spans = append(ingestReq.Spans, span)
			}

			select {
			case ingestChan <- ingestReq:
			default:
				log.Println("Ingest queue full, dropping OTLP trace")
			}
		}
	}

	WriteJSON(w, http.StatusAccepted, map[string]string{"message": "OTLP data accepted"})
}

func extractStringValue(v interface{}) (string, bool) {
	switch val := v.(type) {
	case string:
		return val, true
	case map[string]interface{}:
		if sv, ok := val["stringValue"]; ok {
			if s, ok := sv.(string); ok {
				return s, true
			}
		}
	}
	return "", false
}

func extractIntValue(v interface{}) (int64, bool) {
	switch val := v.(type) {
	case float64:
		return int64(val), true
	case int64:
		return val, true
	case map[string]interface{}:
		if iv, ok := val["intValue"]; ok {
			switch n := iv.(type) {
			case float64:
				return int64(n), true
			case string:
				// OTLP sometimes sends ints as strings
				var i int64
				if _, err := json.Number(n).Int64(); err == nil {
					return i, true
				}
			}
		}
	}
	return 0, false
}
