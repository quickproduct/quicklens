package workers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"

	"quicklens/backend/internal/db"
	"quicklens/backend/internal/ws"
)

type OllamaTagsResponse struct {
	Models []OllamaModel `json:"models"`
}

type OllamaModel struct {
	Name    string             `json:"name"`
	Model   string             `json:"model"`
	Size    int64              `json:"size"`
	Details OllamaModelDetails `json:"details"`
}

type OllamaModelDetails struct {
	Format            string   `json:"format"`
	Family            string   `json:"family"`
	Families          []string `json:"families"`
	ParameterSize     string   `json:"parameter_size"`
	QuantizationLevel string   `json:"quantization_level"`
}

func StartOllamaWorker() {
	pollIntervalSec := 30
	if intervalStr := os.Getenv("OLLAMA_POLL_INTERVAL"); intervalStr != "" {
		if val, err := strconv.Atoi(intervalStr); err == nil && val > 0 {
			pollIntervalSec = val
		}
	}

	go func() {
		ticker := time.NewTicker(time.Duration(pollIntervalSec) * time.Second)
		zap.L().Sugar().Infof("Ollama worker started, polling every %d seconds", pollIntervalSec)

		// Run once immediately
		pollOllama()

		for range ticker.C {
			pollOllama()
		}
	}()
}

func pollOllama() {
	host := os.Getenv("OLLAMA_HOST")
	if host == "" {
		host = "localhost"
	}
	port := os.Getenv("OLLAMA_PORT")
	if port == "" {
		port = "11434"
	}

	url := fmt.Sprintf("http://%s:%s/api/tags", host, port)
	client := http.Client{Timeout: 5 * time.Second}

	resp, err := client.Get(url)
	if err != nil {
		// Just log warning, mark existing ollama models as offline if we can't reach Ollama
		zap.L().Sugar().Infof("Warning: Failed to poll Ollama at %s: %v", url, err)
		setOllamaModelsOffline()
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		zap.L().Sugar().Infof("Warning: Ollama returned status %d", resp.StatusCode)
		setOllamaModelsOffline()
		return
	}

	var tags OllamaTagsResponse
	if err := json.NewDecoder(resp.Body).Decode(&tags); err != nil {
		zap.L().Sugar().Infof("Warning: Failed to decode Ollama response: %v", err)
		return
	}

	now := time.Now().UTC()
	discoveredModelIDs := make(map[string]bool)

	for _, om := range tags.Models {
		discoveredModelIDs[om.Name] = true

		var id string
		var exists bool

		err := db.DB.QueryRow(`
			SELECT id FROM models WHERE provider = 'ollama' AND model_id = ?
		`, om.Name).Scan(&id)

		if err == sql.ErrNoRows {
			id = uuid.New().String()
			exists = false
		} else if err != nil {
			zap.L().Sugar().Infof("Database query error in Ollama worker: %v", err)
			continue
		} else {
			exists = true
		}

		endpoint := fmt.Sprintf("http://%s:%s", host, port)

		if !exists {
			_, err = db.DB.Exec(`
				INSERT INTO models (id, name, provider, model_id, endpoint, status, quantization, size_bytes, context_length, last_seen_at, created_at, updated_at)
				VALUES (?, ?, 'ollama', ?, ?, 'online', ?, ?, 4096, ?, ?, ?)
			`, id, om.Name, om.Name, endpoint, om.Details.QuantizationLevel, om.Size, now, now, now)
		} else {
			_, err = db.DB.Exec(`
				UPDATE models
				SET status = 'online', last_seen_at = ?, updated_at = ?, size_bytes = ?, quantization = ?
				WHERE id = ?
			`, now, now, om.Size, om.Details.QuantizationLevel, id)
		}

		if err != nil {
			zap.L().Sugar().Infof("Failed to upsert Ollama model %s: %v", om.Name, err)
		}
	}

	// Mark models not present in discovery as offline
	rows, err := db.DB.Query("SELECT id, model_id, status FROM models WHERE provider = 'ollama'")
	if err != nil {
		zap.L().Sugar().Infof("Failed to query models for offline check: %v", err)
		return
	}
	defer rows.Close()

	var modelsToUpdate []string
	for rows.Next() {
		var id, modelID, status string
		if rows.Scan(&id, &modelID, &status) == nil {
			if !discoveredModelIDs[modelID] && status == "online" {
				modelsToUpdate = append(modelsToUpdate, id)
			}
		}
	}

	for _, id := range modelsToUpdate {
		_, err = db.DB.Exec("UPDATE models SET status = 'offline', updated_at = ? WHERE id = ?", now, id)
		if err != nil {
			zap.L().Sugar().Infof("Failed to update status for offline model %s: %v", id, err)
		}
	}

	// Broadcast updates via WS
	ws.Manager.Broadcast("models", map[string]string{"type": "update", "time": now.Format(time.RFC3339)})
}

func setOllamaModelsOffline() {
	now := time.Now().UTC()
	_, err := db.DB.Exec(`
		UPDATE models SET status = 'offline', updated_at = ? WHERE provider = 'ollama' AND status = 'online'
	`, now)
	if err != nil {
		zap.L().Sugar().Infof("Failed to set Ollama models offline: %v", err)
	}
	ws.Manager.Broadcast("models", map[string]string{"type": "update", "time": now.Format(time.RFC3339)})
}
