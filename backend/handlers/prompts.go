package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"

	"quicklens/backend/db"
	"quicklens/backend/models"
)

func ListPromptsHandler(w http.ResponseWriter, r *http.Request) {
	// Get latest version for each prompt name
	rows, err := db.DB.Query(`
		SELECT p.id, p.name, p.content, p.model_id, p.version, p.tags, p.created_at
		FROM prompts p
		INNER JOIN (
			SELECT name, MAX(version) as max_version
			FROM prompts
			GROUP BY name
		) latest ON p.name = latest.name AND p.version = latest.max_version
		ORDER BY p.created_at DESC
	`)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to query prompts")
		return
	}
	defer rows.Close()

	prompts := make([]models.PromptResponse, 0)
	for rows.Next() {
		var p models.PromptResponse
		var tagsJSON string
		var createdAt time.Time
		if err := rows.Scan(&p.ID, &p.Name, &p.Content, &p.ModelID, &p.Version, &tagsJSON, &createdAt); err != nil {
			log.Printf("Error scanning prompt: %v", err)
			continue
		}
		p.CreatedAt = &createdAt
		p.Tags = make([]string, 0)
		json.Unmarshal([]byte(tagsJSON), &p.Tags)

		// Get version history for this prompt name
		p.Versions = getVersionHistory(p.Name)

		prompts = append(prompts, p)
	}

	WriteJSON(w, http.StatusOK, prompts)
}

func getVersionHistory(name string) []models.PromptVersionSummary {
	vRows, err := db.DB.Query(
		"SELECT id, version, created_at FROM prompts WHERE name = ? ORDER BY version DESC",
		name,
	)
	if err != nil {
		return make([]models.PromptVersionSummary, 0)
	}
	defer vRows.Close()

	versions := make([]models.PromptVersionSummary, 0)
	for vRows.Next() {
		var vs models.PromptVersionSummary
		var ca time.Time
		if vRows.Scan(&vs.ID, &vs.Version, &ca) == nil {
			vs.CreatedAt = &ca
			versions = append(versions, vs)
		}
	}
	return versions
}

func CreatePromptHandler(w http.ResponseWriter, r *http.Request) {
	var req models.PromptCreateRequest
	if err := ParseJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Name == "" || req.Content == "" {
		WriteError(w, http.StatusBadRequest, "Name and content are required")
		return
	}

	// Check if a prompt with this name already exists
	var existingID string
	var maxVersion int
	err := db.DB.QueryRow(
		"SELECT id, MAX(version) FROM prompts WHERE name = ? GROUP BY name",
		req.Name,
	).Scan(&existingID, &maxVersion)

	version := 1
	parentID := ""
	if err == nil {
		// Existing prompt - create new version
		version = maxVersion + 1
		parentID = existingID
	}

	id := uuid.New().String()
	now := time.Now().UTC()

	tagsJSON := "[]"
	if req.Tags != nil {
		if b, err := json.Marshal(req.Tags); err == nil {
			tagsJSON = string(b)
		}
	}

	_, err = db.DB.Exec(
		"INSERT INTO prompts (id, name, content, model_id, version, parent_id, tags, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		id, req.Name, req.Content, req.ModelID, version, parentID, tagsJSON, now,
	)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to create prompt")
		return
	}

	tags := make([]string, 0)
	if req.Tags != nil {
		tags = req.Tags
	}

	WriteJSON(w, http.StatusCreated, models.PromptResponse{
		ID:        id,
		Name:      req.Name,
		Content:   req.Content,
		ModelID:   req.ModelID,
		Version:   version,
		Tags:      tags,
		CreatedAt: &now,
		Versions:  getVersionHistory(req.Name),
	})
}

func GetPromptHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		WriteError(w, http.StatusBadRequest, "Prompt ID is required")
		return
	}

	var p models.PromptResponse
	var tagsJSON string
	var createdAt time.Time
	err := db.DB.QueryRow(
		"SELECT id, name, content, model_id, version, tags, created_at FROM prompts WHERE id = ?",
		id,
	).Scan(&p.ID, &p.Name, &p.Content, &p.ModelID, &p.Version, &tagsJSON, &createdAt)
	if err != nil {
		if err == sql.ErrNoRows {
			WriteError(w, http.StatusNotFound, "Prompt not found")
		} else {
			WriteError(w, http.StatusInternalServerError, "Failed to query prompt")
		}
		return
	}
	p.CreatedAt = &createdAt
	p.Tags = make([]string, 0)
	json.Unmarshal([]byte(tagsJSON), &p.Tags)
	p.Versions = getVersionHistory(p.Name)

	WriteJSON(w, http.StatusOK, p)
}

func DiffPromptVersionsHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	versionStr := r.PathValue("version")
	if id == "" || versionStr == "" {
		WriteError(w, http.StatusBadRequest, "Prompt ID and version are required")
		return
	}

	targetVersion, err := strconv.Atoi(versionStr)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid version number")
		return
	}

	// Get the current prompt (version A)
	var name, contentA string
	var versionA int
	err = db.DB.QueryRow(
		"SELECT name, content, version FROM prompts WHERE id = ?",
		id,
	).Scan(&name, &contentA, &versionA)
	if err != nil {
		WriteError(w, http.StatusNotFound, "Prompt not found")
		return
	}

	// Get the target version (version B)
	var contentB string
	err = db.DB.QueryRow(
		"SELECT content FROM prompts WHERE name = ? AND version = ?",
		name, targetVersion,
	).Scan(&contentB)
	if err != nil {
		WriteError(w, http.StatusNotFound, "Target version not found")
		return
	}

	WriteJSON(w, http.StatusOK, models.PromptDiffResponse{
		VersionA: versionA,
		VersionB: targetVersion,
		ContentA: contentA,
		ContentB: contentB,
	})
}

func DeletePromptHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		WriteError(w, http.StatusBadRequest, "Prompt ID is required")
		return
	}

	result, err := db.DB.Exec("DELETE FROM prompts WHERE id = ?", id)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to delete prompt")
		return
	}

	affected, _ := result.RowsAffected()
	if affected == 0 {
		WriteError(w, http.StatusNotFound, "Prompt not found")
		return
	}

	WriteJSON(w, http.StatusOK, map[string]string{"message": "Prompt deleted"})
}
