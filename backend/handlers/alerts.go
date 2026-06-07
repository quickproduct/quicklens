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

func ListAlertsHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.DB.Query(`
		SELECT id, rule_id, severity, message, acknowledged, status, owner_id, service_id,
		       model_id, incident_id, dedupe_key, runbook_url, last_seen_at, resolved_at, created_at
		FROM alerts
		ORDER BY CASE status WHEN 'open' THEN 0 WHEN 'acknowledged' THEN 1 WHEN 'resolved' THEN 2 ELSE 3 END,
		         CASE severity WHEN 'critical' THEN 0 WHEN 'warning' THEN 1 ELSE 2 END,
		         created_at DESC
	`)
	if err != nil {
		log.Printf("Failed to query alerts: %v", err)
		WriteError(w, http.StatusInternalServerError, "Failed to query alerts")
		return
	}
	defer rows.Close()

	result := make([]models.AlertResponse, 0)
	for rows.Next() {
		var a models.AlertResponse
		var ruleID sql.NullString
		var lastSeenAt, resolvedAt sql.NullTime
		var createdAt time.Time
		err := rows.Scan(
			&a.ID, &ruleID, &a.Severity, &a.Message, &a.Acknowledged, &a.Status,
			&a.OwnerID, &a.ServiceID, &a.ModelID, &a.IncidentID, &a.DedupeKey,
			&a.RunbookURL, &lastSeenAt, &resolvedAt, &createdAt,
		)
		if err != nil {
			log.Printf("Failed to scan alert: %v", err)
			continue
		}
		if ruleID.Valid {
			a.RuleID = &ruleID.String
		}
		if lastSeenAt.Valid {
			a.LastSeenAt = &lastSeenAt.Time
		}
		if resolvedAt.Valid {
			a.ResolvedAt = &resolvedAt.Time
		}
		a.CreatedAt = &createdAt
		result = append(result, a)
	}

	WriteJSON(w, http.StatusOK, result)
}

func AcknowledgeAlertHandler(w http.ResponseWriter, r *http.Request) {
	alertID := r.PathValue("alert_id")
	if alertID == "" {
		WriteError(w, http.StatusBadRequest, "Alert ID is required")
		return
	}

	_, err := db.DB.Exec("UPDATE alerts SET acknowledged = 1, status = 'acknowledged' WHERE id = ?", alertID)
	if err != nil {
		log.Printf("Failed to acknowledge alert: %v", err)
		WriteError(w, http.StatusInternalServerError, "Failed to acknowledge alert")
		return
	}

	WriteJSON(w, http.StatusOK, map[string]string{"status": "success"})
}

func ListRulesHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.DB.Query(`
		SELECT id, metric_type, threshold, operator, window_seconds, enabled, created_at
		FROM alert_rules
		ORDER BY created_at DESC
	`)
	if err != nil {
		log.Printf("Failed to query alert rules: %v", err)
		WriteError(w, http.StatusInternalServerError, "Failed to query alert rules")
		return
	}
	defer rows.Close()

	result := make([]models.AlertRuleResponse, 0)
	for rows.Next() {
		var rule models.AlertRuleResponse
		var createdAt time.Time
		err := rows.Scan(
			&rule.ID, &rule.MetricType, &rule.Threshold, &rule.Operator,
			&rule.WindowSeconds, &rule.Enabled, &createdAt,
		)
		if err != nil {
			log.Printf("Failed to scan alert rule: %v", err)
			continue
		}
		rule.CreatedAt = &createdAt
		result = append(result, rule)
	}

	WriteJSON(w, http.StatusOK, result)
}

func CreateRuleHandler(w http.ResponseWriter, r *http.Request) {
	var req models.AlertRuleCreate
	if err := ParseJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.MetricType == "" || req.Operator == "" || req.WindowSeconds <= 0 {
		WriteError(w, http.StatusBadRequest, "MetricType, Operator, and valid WindowSeconds are required")
		return
	}

	id := uuid.New().String()
	now := time.Now().UTC()

	_, err := db.DB.Exec(`
		INSERT INTO alert_rules (id, metric_type, threshold, operator, window_seconds, enabled, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, id, req.MetricType, req.Threshold, req.Operator, req.WindowSeconds, req.Enabled, now)
	if err != nil {
		log.Printf("Failed to create alert rule: %v", err)
		WriteError(w, http.StatusInternalServerError, "Failed to create alert rule")
		return
	}

	resp := models.AlertRuleResponse{
		ID:            id,
		MetricType:    req.MetricType,
		Threshold:     req.Threshold,
		Operator:      req.Operator,
		WindowSeconds: req.WindowSeconds,
		Enabled:       req.Enabled,
		CreatedAt:     &now,
	}

	WriteJSON(w, http.StatusCreated, resp)
}

func UpdateRuleHandler(w http.ResponseWriter, r *http.Request) {
	ruleID := r.PathValue("rule_id")
	if ruleID == "" {
		WriteError(w, http.StatusBadRequest, "Rule ID is required")
		return
	}

	var req models.AlertRuleUpdate
	if err := ParseJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Read existing rule
	var threshold float64
	var operator string
	var windowSeconds int
	var enabled bool
	err := db.DB.QueryRow(`
		SELECT threshold, operator, window_seconds, enabled
		FROM alert_rules WHERE id = ?
	`, ruleID).Scan(&threshold, &operator, &windowSeconds, &enabled)
	if err != nil {
		if err == sql.ErrNoRows {
			WriteError(w, http.StatusNotFound, "Rule not found")
		} else {
			WriteError(w, http.StatusInternalServerError, "Failed to fetch rule")
		}
		return
	}

	// Update fields if provided
	if req.Threshold != nil {
		threshold = *req.Threshold
	}
	if req.Operator != nil {
		operator = *req.Operator
	}
	if req.WindowSeconds != nil {
		windowSeconds = *req.WindowSeconds
	}
	if req.Enabled != nil {
		enabled = *req.Enabled
	}

	_, err = db.DB.Exec(`
		UPDATE alert_rules
		SET threshold = ?, operator = ?, window_seconds = ?, enabled = ?
		WHERE id = ?
	`, threshold, operator, windowSeconds, enabled, ruleID)
	if err != nil {
		log.Printf("Failed to update alert rule: %v", err)
		WriteError(w, http.StatusInternalServerError, "Failed to update alert rule")
		return
	}

	// Fetch updated rule
	var updated models.AlertRuleResponse
	var createdAt time.Time
	err = db.DB.QueryRow(`
		SELECT id, metric_type, threshold, operator, window_seconds, enabled, created_at
		FROM alert_rules WHERE id = ?
	`, ruleID).Scan(&updated.ID, &updated.MetricType, &updated.Threshold, &updated.Operator, &updated.WindowSeconds, &updated.Enabled, &createdAt)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to fetch updated rule")
		return
	}
	updated.CreatedAt = &createdAt

	WriteJSON(w, http.StatusOK, updated)
}

func DeleteRuleHandler(w http.ResponseWriter, r *http.Request) {
	ruleID := r.PathValue("rule_id")
	if ruleID == "" {
		WriteError(w, http.StatusBadRequest, "Rule ID is required")
		return
	}

	_, err := db.DB.Exec("DELETE FROM alert_rules WHERE id = ?", ruleID)
	if err != nil {
		log.Printf("Failed to delete alert rule: %v", err)
		WriteError(w, http.StatusInternalServerError, "Failed to delete alert rule")
		return
	}

	WriteJSON(w, http.StatusOK, map[string]string{"status": "success"})
}
