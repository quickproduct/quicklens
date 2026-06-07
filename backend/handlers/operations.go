package handlers

import (
	"database/sql"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"

	"quicklens/backend/db"
	"quicklens/backend/models"
)

func ListIncidentsHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.DB.Query(`
		SELECT i.id, i.title, i.severity, i.status, i.owner_id, i.service_id, i.model_id,
		       i.alert_id, i.summary, i.runbook_url, i.started_at, i.resolved_at,
		       i.updated_at, i.created_at,
		       (SELECT COUNT(*) FROM incident_events WHERE incident_id = i.id) AS event_count
		FROM incidents i
		ORDER BY CASE i.status WHEN 'investigating' THEN 0 WHEN 'mitigating' THEN 1 WHEN 'monitoring' THEN 2 WHEN 'resolved' THEN 3 ELSE 4 END,
		         CASE i.severity WHEN 'critical' THEN 0 WHEN 'warning' THEN 1 ELSE 2 END,
		         i.created_at DESC
	`)
	if err != nil {
		log.Printf("Failed to query incidents: %v", err)
		WriteError(w, http.StatusInternalServerError, "Failed to query incidents")
		return
	}
	defer rows.Close()

	result := make([]models.IncidentResponse, 0)
	for rows.Next() {
		var item models.IncidentResponse
		var startedAt, createdAt, updatedAt time.Time
		var resolvedAt sql.NullTime
		if err := rows.Scan(
			&item.ID, &item.Title, &item.Severity, &item.Status, &item.OwnerID,
			&item.ServiceID, &item.ModelID, &item.AlertID, &item.Summary, &item.RunbookURL,
			&startedAt, &resolvedAt, &updatedAt, &createdAt, &item.EventCount,
		); err != nil {
			log.Printf("Failed to scan incident: %v", err)
			continue
		}
		item.StartedAt = &startedAt
		item.UpdatedAt = &updatedAt
		item.CreatedAt = &createdAt
		if resolvedAt.Valid {
			item.ResolvedAt = &resolvedAt.Time
		}
		result = append(result, item)
	}

	WriteJSON(w, http.StatusOK, result)
}

func CreateIncidentHandler(w http.ResponseWriter, r *http.Request) {
	var req models.IncidentCreateRequest
	if err := ParseJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if req.Title == "" {
		req.Title = "Untitled incident"
	}
	if req.Severity == "" {
		req.Severity = "warning"
	}

	id := uuid.New().String()
	now := time.Now().UTC()
	_, err := db.DB.Exec(`
		INSERT INTO incidents (id, title, severity, status, owner_id, service_id, model_id, alert_id, summary, runbook_url, started_at, updated_at, created_at)
		VALUES (?, ?, ?, 'investigating', ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, id, req.Title, req.Severity, req.OwnerID, req.ServiceID, req.ModelID, req.AlertID, req.Summary, req.RunbookURL, now, now, now)
	if err != nil {
		log.Printf("Failed to create incident: %v", err)
		WriteError(w, http.StatusInternalServerError, "Failed to create incident")
		return
	}

	if req.AlertID != "" {
		_, _ = db.DB.Exec("UPDATE alerts SET incident_id = ?, status = 'acknowledged', acknowledged = 1 WHERE id = ?", id, req.AlertID)
	}
	writeIncidentEvent(id, "created", "Incident declared", getUserID(r))
	writeAuditLog(getUserID(r), "incident.created", "incident", id)

	WriteJSON(w, http.StatusCreated, models.IncidentResponse{
		ID: id, Title: req.Title, Severity: req.Severity, Status: "investigating",
		OwnerID: req.OwnerID, ServiceID: req.ServiceID, ModelID: req.ModelID, AlertID: req.AlertID,
		Summary: req.Summary, RunbookURL: req.RunbookURL, StartedAt: &now, UpdatedAt: &now, CreatedAt: &now,
		EventCount: 1,
	})
}

func UpdateIncidentHandler(w http.ResponseWriter, r *http.Request) {
	incidentID := r.PathValue("incident_id")
	if incidentID == "" {
		WriteError(w, http.StatusBadRequest, "Incident ID is required")
		return
	}

	var req models.IncidentUpdateRequest
	if err := ParseJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	var status, ownerID, summary, runbookURL string
	err := db.DB.QueryRow("SELECT status, owner_id, summary, runbook_url FROM incidents WHERE id = ?", incidentID).
		Scan(&status, &ownerID, &summary, &runbookURL)
	if err != nil {
		if err == sql.ErrNoRows {
			WriteError(w, http.StatusNotFound, "Incident not found")
		} else {
			WriteError(w, http.StatusInternalServerError, "Failed to fetch incident")
		}
		return
	}

	if req.Status != nil {
		status = *req.Status
	}
	if req.OwnerID != nil {
		ownerID = *req.OwnerID
	}
	if req.Summary != nil {
		summary = *req.Summary
	}
	if req.RunbookURL != nil {
		runbookURL = *req.RunbookURL
	}

	now := time.Now().UTC()
	if status == "resolved" {
		_, err = db.DB.Exec(`
			UPDATE incidents SET status = ?, owner_id = ?, summary = ?, runbook_url = ?, resolved_at = COALESCE(resolved_at, ?), updated_at = ?
			WHERE id = ?
		`, status, ownerID, summary, runbookURL, now, now, incidentID)
	} else {
		_, err = db.DB.Exec(`
			UPDATE incidents SET status = ?, owner_id = ?, summary = ?, runbook_url = ?, resolved_at = NULL, updated_at = ?
			WHERE id = ?
		`, status, ownerID, summary, runbookURL, now, incidentID)
	}
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to update incident")
		return
	}

	writeIncidentEvent(incidentID, "updated", "Incident updated", getUserID(r))
	writeAuditLog(getUserID(r), "incident.updated", "incident", incidentID)
	WriteJSON(w, http.StatusOK, map[string]string{"status": "success"})
}

func ListIncidentEventsHandler(w http.ResponseWriter, r *http.Request) {
	incidentID := r.PathValue("incident_id")
	rows, err := db.DB.Query(`
		SELECT id, incident_id, event_type, message, actor_id, created_at
		FROM incident_events
		WHERE incident_id = ?
		ORDER BY created_at ASC
	`, incidentID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to query incident events")
		return
	}
	defer rows.Close()

	result := make([]models.IncidentEventResponse, 0)
	for rows.Next() {
		var item models.IncidentEventResponse
		var createdAt time.Time
		if err := rows.Scan(&item.ID, &item.IncidentID, &item.EventType, &item.Message, &item.ActorID, &createdAt); err == nil {
			item.CreatedAt = &createdAt
			result = append(result, item)
		}
	}
	WriteJSON(w, http.StatusOK, result)
}

func ListAuditLogsHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.DB.Query(`
		SELECT id, actor_id, action, resource, resource_id, metadata, created_at
		FROM audit_logs
		ORDER BY created_at DESC
		LIMIT 100
	`)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to query audit logs")
		return
	}
	defer rows.Close()

	result := make([]models.AuditLogResponse, 0)
	for rows.Next() {
		var item models.AuditLogResponse
		var createdAt time.Time
		if err := rows.Scan(&item.ID, &item.ActorID, &item.Action, &item.Resource, &item.ResourceID, &item.Metadata, &createdAt); err == nil {
			item.CreatedAt = &createdAt
			result = append(result, item)
		}
	}
	WriteJSON(w, http.StatusOK, result)
}

func ListSavedViewsHandler(w http.ResponseWriter, r *http.Request) {
	scope := r.URL.Query().Get("scope")
	query := "SELECT id, name, scope, filters, is_shared, created_at FROM saved_views"
	args := []interface{}{}
	if scope != "" {
		query += " WHERE scope = ?"
		args = append(args, scope)
	}
	query += " ORDER BY created_at DESC"

	rows, err := db.DB.Query(query, args...)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to query saved views")
		return
	}
	defer rows.Close()

	result := make([]models.SavedViewResponse, 0)
	for rows.Next() {
		var item models.SavedViewResponse
		var createdAt time.Time
		if err := rows.Scan(&item.ID, &item.Name, &item.Scope, &item.Filters, &item.IsShared, &createdAt); err == nil {
			item.CreatedAt = &createdAt
			result = append(result, item)
		}
	}
	WriteJSON(w, http.StatusOK, result)
}

func ListSLODefinitionsHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.DB.Query(`
		SELECT id, name, service_id, target_percent, period_days, created_at
		FROM slo_definitions
		ORDER BY created_at DESC
	`)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to query SLO definitions")
		return
	}
	defer rows.Close()

	result := make([]models.SLODefinitionResponse, 0)
	for rows.Next() {
		var item models.SLODefinitionResponse
		var createdAt time.Time
		if err := rows.Scan(&item.ID, &item.Name, &item.ServiceID, &item.TargetPercent, &item.PeriodDays, &createdAt); err != nil {
			continue
		}
		item.CreatedAt = &createdAt
		item.GoodEvents, item.TotalEvents = currentSLOEvents()
		item.ErrorBudgetRemaining, item.BurnRate, item.Status = computeSLOState(item.TargetPercent, item.GoodEvents, item.TotalEvents)
		result = append(result, item)
	}
	if len(result) == 0 {
		good, total := currentSLOEvents()
		remaining, burn, status := computeSLOState(99, good, total)
		now := time.Now().UTC()
		result = append(result, models.SLODefinitionResponse{
			ID: "default-availability", Name: "Default trace success SLO", ServiceID: "quicklens",
			TargetPercent: 99, PeriodDays: 28, GoodEvents: good, TotalEvents: total,
			ErrorBudgetRemaining: remaining, BurnRate: burn, Status: status, CreatedAt: &now,
		})
	}
	WriteJSON(w, http.StatusOK, result)
}

func ListNotificationRulesHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.DB.Query("SELECT id, name, channel, target, enabled, created_at FROM notification_rules ORDER BY created_at DESC")
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to query notification rules")
		return
	}
	defer rows.Close()
	WriteJSON(w, http.StatusOK, rowsToMaps(rows))
}

func ListDashboardLayoutsHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.DB.Query("SELECT id, name, layout_json, is_default, created_at FROM dashboard_layouts ORDER BY is_default DESC, created_at DESC")
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to query dashboard layouts")
		return
	}
	defer rows.Close()
	WriteJSON(w, http.StatusOK, rowsToMaps(rows))
}

func writeIncidentEvent(incidentID, eventType, message, actorID string) {
	_, _ = db.DB.Exec(
		"INSERT INTO incident_events (id, incident_id, event_type, message, actor_id, created_at) VALUES (?, ?, ?, ?, ?, ?)",
		uuid.New().String(), incidentID, eventType, message, actorID, time.Now().UTC(),
	)
}

func writeAuditLog(actorID, action, resource, resourceID string) {
	_, _ = db.DB.Exec(
		"INSERT INTO audit_logs (id, actor_id, action, resource, resource_id, metadata, created_at) VALUES (?, ?, ?, ?, ?, '{}', ?)",
		uuid.New().String(), actorID, action, resource, resourceID, time.Now().UTC(),
	)
}

func getUserID(r *http.Request) string {
	if v, ok := r.Context().Value(UserIDKey).(string); ok {
		return v
	}
	return ""
}

func currentSLOEvents() (int64, int64) {
	now := time.Now().UTC()
	start := now.Add(-28 * 24 * time.Hour)
	var good, total int64
	db.DB.QueryRow("SELECT COALESCE(COUNT(*), 0) FROM traces WHERE created_at >= ? AND status != 'error'", start).Scan(&good)
	db.DB.QueryRow("SELECT COALESCE(COUNT(*), 0) FROM traces WHERE created_at >= ?", start).Scan(&total)
	return good, total
}

func computeSLOState(target float64, good, total int64) (float64, float64, string) {
	attainment := 100.0
	if total > 0 {
		attainment = float64(good) / float64(total) * 100
	}
	remaining := 100.0
	if attainment < target {
		remaining = 0
	} else if target < 100 {
		remaining = (attainment - target) / (100 - target) * 100
	}
	status := "healthy"
	if remaining <= 0 {
		status = "critical"
	} else if remaining < 25 {
		status = "warning"
	}
	return remaining, 100 - remaining, status
}

func rowsToMaps(rows *sql.Rows) []map[string]interface{} {
	cols, err := rows.Columns()
	if err != nil {
		return []map[string]interface{}{}
	}
	result := make([]map[string]interface{}, 0)
	for rows.Next() {
		values := make([]interface{}, len(cols))
		valuePtrs := make([]interface{}, len(cols))
		for i := range values {
			valuePtrs[i] = &values[i]
		}
		if err := rows.Scan(valuePtrs...); err != nil {
			continue
		}
		item := map[string]interface{}{}
		for i, col := range cols {
			if b, ok := values[i].([]byte); ok {
				item[col] = string(b)
			} else {
				item[col] = values[i]
			}
		}
		result = append(result, item)
	}
	return result
}
