package handlers

import (
	"github.com/go-chi/chi/v5"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"go.uber.org/zap"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"

	"quicklens/backend/internal/db"
	"quicklens/backend/internal/models"
)

func ListModelsHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.DB.Query(`
		SELECT m.id, m.name, m.provider, m.model_id, m.endpoint, m.status,
		       m.quantization, m.size_bytes, m.context_length, m.last_seen_at, m.created_at,
		       COALESCE(stats.total_requests, 0),
		       COALESCE(stats.avg_latency, 0),
		       COALESCE(stats.total_tokens, 0)
		FROM models m
		LEFT JOIN (
			SELECT model_id, provider,
			       COUNT(*) as total_requests,
			       AVG(duration_ms) as avg_latency,
			       SUM(total_tokens) as total_tokens
			FROM spans
			WHERE type = 'llm'
			GROUP BY model_id, provider
		) stats ON m.model_id = stats.model_id AND m.provider = stats.provider
		ORDER BY m.created_at DESC
	`)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to query models")
		return
	}
	defer rows.Close()

	result := make([]models.ModelResponse, 0)
	for rows.Next() {
		var m models.ModelResponse
		var lastSeenAt, createdAt sql.NullTime
		err := rows.Scan(
			&m.ID, &m.Name, &m.Provider, &m.ModelID, &m.Endpoint, &m.Status,
			&m.Quantization, &m.SizeBytes, &m.ContextLength, &lastSeenAt, &createdAt,
			&m.TotalRequests, &m.AvgLatencyMs, &m.TotalTokens,
		)
		if err != nil {
			zap.L().Sugar().Infof("Error scanning model row: %v", err)
			continue
		}
		if lastSeenAt.Valid {
			m.LastSeenAt = &lastSeenAt.Time
		}
		if createdAt.Valid {
			m.CreatedAt = &createdAt.Time
		}
		result = append(result, m)
	}

	WriteJSON(w, http.StatusOK, result)
}

func CreateModelHandler(w http.ResponseWriter, r *http.Request) {
	var req models.ModelCreateRequest
	if err := ParseJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Name == "" || req.Provider == "" {
		WriteError(w, http.StatusBadRequest, "Name and provider are required")
		return
	}

	id := uuid.New().String()
	now := time.Now().UTC()
	modelID := req.ModelID
	if modelID == "" {
		modelID = req.Name
	}

	_, err := db.DB.Exec(
		`INSERT INTO models (id, name, provider, model_id, endpoint, status, context_length, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, 'online', ?, ?, ?)`,
		id, req.Name, req.Provider, modelID, req.Endpoint, req.ContextLength, now, now,
	)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to create model")
		return
	}

	WriteJSON(w, http.StatusCreated, models.ModelResponse{
		ID:            id,
		Name:          req.Name,
		Provider:      req.Provider,
		ModelID:       modelID,
		Endpoint:      req.Endpoint,
		Status:        "online",
		ContextLength: req.ContextLength,
		CreatedAt:     &now,
	})
}

func GetModelHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		WriteError(w, http.StatusBadRequest, "Model ID is required")
		return
	}

	var m models.ModelResponse
	var lastSeenAt, createdAt sql.NullTime
	err := db.DB.QueryRow(`
		SELECT m.id, m.name, m.provider, m.model_id, m.endpoint, m.status,
		       m.quantization, m.size_bytes, m.context_length, m.last_seen_at, m.created_at,
		       COALESCE(stats.total_requests, 0),
		       COALESCE(stats.avg_latency, 0),
		       COALESCE(stats.total_tokens, 0)
		FROM models m
		LEFT JOIN (
			SELECT model_id, provider,
			       COUNT(*) as total_requests,
			       AVG(duration_ms) as avg_latency,
			       SUM(total_tokens) as total_tokens
			FROM spans
			WHERE type = 'llm'
			GROUP BY model_id, provider
		) stats ON m.model_id = stats.model_id AND m.provider = stats.provider
		WHERE m.id = ?
	`, id).Scan(
		&m.ID, &m.Name, &m.Provider, &m.ModelID, &m.Endpoint, &m.Status,
		&m.Quantization, &m.SizeBytes, &m.ContextLength, &lastSeenAt, &createdAt,
		&m.TotalRequests, &m.AvgLatencyMs, &m.TotalTokens,
	)
	if err != nil {
		WriteError(w, http.StatusNotFound, "Model not found")
		return
	}
	if lastSeenAt.Valid {
		m.LastSeenAt = &lastSeenAt.Time
	}
	if createdAt.Valid {
		m.CreatedAt = &createdAt.Time
	}

	WriteJSON(w, http.StatusOK, m)
}

func DeleteModelHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		WriteError(w, http.StatusBadRequest, "Model ID is required")
		return
	}

	result, err := db.DB.Exec("DELETE FROM models WHERE id = ?", id)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to delete model")
		return
	}

	affected, _ := result.RowsAffected()
	if affected == 0 {
		WriteError(w, http.StatusNotFound, "Model not found")
		return
	}

	WriteJSON(w, http.StatusOK, map[string]string{"message": "Model deleted"})
}

func DiscoverModelsHandler(w http.ResponseWriter, r *http.Request) {
	ollamaHost := os.Getenv("OLLAMA_HOST")
	if ollamaHost == "" {
		ollamaHost = "localhost:11434"
	}

	resp, err := http.Get(fmt.Sprintf("http://%s/api/tags", ollamaHost))
	if err != nil {
		WriteError(w, http.StatusBadGateway, "Failed to connect to Ollama: "+err.Error())
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to read Ollama response")
		return
	}

	var ollamaResp struct {
		Models []struct {
			Name       string `json:"name"`
			Model      string `json:"model"`
			ModifiedAt string `json:"modified_at"`
			Size       int64  `json:"size"`
			Details    struct {
				Format            string `json:"format"`
				Family            string `json:"family"`
				ParameterSize     string `json:"parameter_size"`
				QuantizationLevel string `json:"quantization_level"`
			} `json:"details"`
		} `json:"models"`
	}

	if err := json.Unmarshal(body, &ollamaResp); err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to parse Ollama response")
		return
	}

	now := time.Now().UTC()
	discovered := make([]models.ModelResponse, 0)

	for _, om := range ollamaResp.Models {
		modelName := om.Name
		modelID := om.Model
		if modelID == "" {
			modelID = modelName
		}

		// Upsert: check if exists
		var existingID string
		err := db.DB.QueryRow(
			"SELECT id FROM models WHERE provider = 'ollama' AND model_id = ?", modelID,
		).Scan(&existingID)

		if err == sql.ErrNoRows {
			// Insert new model
			id := uuid.New().String()
			_, err = db.DB.Exec(
				`INSERT INTO models (id, name, provider, model_id, endpoint, status, quantization, size_bytes, last_seen_at, created_at, updated_at)
				 VALUES (?, ?, 'ollama', ?, ?, 'online', ?, ?, ?, ?, ?)`,
				id, modelName, modelID, fmt.Sprintf("http://%s", ollamaHost),
				om.Details.QuantizationLevel, om.Size, now, now, now,
			)
			if err != nil {
				zap.L().Sugar().Infof("Failed to insert discovered model %s: %v", modelName, err)
				continue
			}
			discovered = append(discovered, models.ModelResponse{
				ID:        id,
				Name:      modelName,
				Provider:  "ollama",
				ModelID:   modelID,
				Status:    "online",
				SizeBytes: om.Size,
				CreatedAt: &now,
			})
		} else if err == nil {
			// Update existing
			_, _ = db.DB.Exec(
				"UPDATE models SET status = 'online', last_seen_at = ?, size_bytes = ?, quantization = ?, updated_at = ? WHERE id = ?",
				now, om.Size, om.Details.QuantizationLevel, now, existingID,
			)
		}
	}

	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"message":    fmt.Sprintf("Discovered %d new models", len(discovered)),
		"discovered": discovered,
	})
}
