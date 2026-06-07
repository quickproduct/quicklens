package workers

import (
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"

	"quicklens/backend/db"
	"quicklens/backend/ws"
)

func StartAlertsWorker() {
	go func() {
		ticker := time.NewTicker(60 * time.Second)
		log.Println("Alerts worker started, running every 60 seconds")

		// Run once immediately
		evaluateAlertRules()

		for range ticker.C {
			evaluateAlertRules()
		}
	}()
}

type alertRule struct {
	id            string
	metricType    string
	threshold     float64
	operator      string
	windowSeconds int
}

func evaluateAlertRules() {
	rows, err := db.DB.Query(`
		SELECT id, metric_type, threshold, operator, window_seconds
		FROM alert_rules
		WHERE enabled = 1
	`)
	if err != nil {
		log.Printf("Alerts worker: failed to query enabled rules: %v", err)
		return
	}

	var rules []alertRule
	for rows.Next() {
		var r alertRule
		if err := rows.Scan(&r.id, &r.metricType, &r.threshold, &r.operator, &r.windowSeconds); err != nil {
			log.Printf("Alerts worker: error scanning rule: %v", err)
			continue
		}
		rules = append(rules, r)
	}
	rows.Close()

	now := time.Now().UTC()

	for _, r := range rules {
		ruleID := r.id
		metricType := r.metricType
		operator := r.operator
		threshold := r.threshold
		windowSeconds := r.windowSeconds

		// Check if an unacknowledged alert already exists for this rule to avoid alert spamming
		var existingCount int
		err := db.DB.QueryRow(`
			SELECT COUNT(*) FROM alerts WHERE rule_id = ? AND acknowledged = 0
		`, ruleID).Scan(&existingCount)
		if err != nil {
			log.Printf("Alerts worker: failed to query existing alerts: %v", err)
			continue
		}
		if existingCount > 0 {
			// Alert already active, skip triggering another
			continue
		}

		windowStart := now.Add(-time.Duration(windowSeconds) * time.Second)
		var currentValue float64
		var triggerAlert bool

		switch metricType {
		case "cost":
			db.DB.QueryRow(`
				SELECT COALESCE(SUM(total_cost), 0) FROM traces WHERE created_at >= ?
			`, windowStart).Scan(&currentValue)
		case "error_rate":
			var total, errors float64
			db.DB.QueryRow(`
				SELECT COUNT(*), COALESCE(SUM(CASE WHEN status = 'error' THEN 1 ELSE 0 END), 0)
				FROM traces WHERE created_at >= ?
			`, windowStart).Scan(&total, &errors)
			if total > 0 {
				currentValue = errors / total
			}
		case "latency":
			db.DB.QueryRow(`
				SELECT COALESCE(AVG(total_duration_ms), 0) FROM traces WHERE created_at >= ?
			`, windowStart).Scan(&currentValue)
		case "tokens":
			db.DB.QueryRow(`
				SELECT COALESCE(SUM(total_tokens), 0) FROM traces WHERE created_at >= ?
			`, windowStart).Scan(&currentValue)
		default:
			log.Printf("Alerts worker: unknown metric type: %s", metricType)
			continue
		}

		switch operator {
		case ">":
			triggerAlert = currentValue > threshold
		case ">=":
			triggerAlert = currentValue >= threshold
		case "<":
			triggerAlert = currentValue < threshold
		case "<=":
			triggerAlert = currentValue <= threshold
		default:
			log.Printf("Alerts worker: unknown operator: %s", operator)
			continue
		}

		if triggerAlert {
			alertID := uuid.New().String()
			severity := "warning"
			if metricType == "error_rate" && currentValue > 0.5 {
				severity = "critical"
			}

			var message string
			switch metricType {
			case "cost":
				message = fmt.Sprintf("Cost threshold breached: $%.4f (Threshold: $%.4f)", currentValue, threshold)
			case "error_rate":
				message = fmt.Sprintf("Error rate threshold breached: %.1f%% (Threshold: %.1f%%)", currentValue*100.0, threshold*100.0)
			case "latency":
				message = fmt.Sprintf("Average latency threshold breached: %.2fms (Threshold: %.2fms)", currentValue, threshold)
			case "tokens":
				message = fmt.Sprintf("Token budget threshold breached: %.0f tokens (Threshold: %.0f)", currentValue, threshold)
			}

			_, err = db.DB.Exec(`
				INSERT INTO alerts (id, rule_id, severity, message, acknowledged, created_at)
				VALUES (?, ?, ?, ?, 0, ?)
			`, alertID, ruleID, severity, message, now)
			if err != nil {
				log.Printf("Alerts worker: failed to create alert: %v", err)
				continue
			}

			log.Printf("Alert triggered! Severity: %s, Message: %s", severity, message)

			// Broadcast the new alert
			ws.Manager.Broadcast("alerts", map[string]interface{}{
				"type": "new_alert",
				"alert": map[string]interface{}{
					"id":           alertID,
					"rule_id":      ruleID,
					"severity":     severity,
					"message":      message,
					"acknowledged": false,
					"created_at":   now.Format(time.RFC3339),
				},
			})
		}
	}
}
