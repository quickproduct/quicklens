package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"quicklens/backend/db"
	"quicklens/backend/models"
)

func GetDashboardHandler(w http.ResponseWriter, r *http.Request) {
	now := time.Now().UTC()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	var resp models.DashboardResponse

	// Total traces today
	db.DB.QueryRow(
		"SELECT COALESCE(COUNT(*), 0) FROM traces WHERE created_at >= ?", startOfDay,
	).Scan(&resp.TotalTracesToday)

	// Total tokens today
	db.DB.QueryRow(
		"SELECT COALESCE(SUM(total_tokens), 0) FROM traces WHERE created_at >= ?", startOfDay,
	).Scan(&resp.TotalTokensToday)

	// Total cost today
	db.DB.QueryRow(
		"SELECT COALESCE(SUM(total_cost), 0) FROM traces WHERE created_at >= ?", startOfDay,
	).Scan(&resp.TotalCostToday)

	// Average latency today
	db.DB.QueryRow(
		"SELECT COALESCE(AVG(total_duration_ms), 0) FROM traces WHERE created_at >= ?", startOfDay,
	).Scan(&resp.AvgLatencyMs)

	// Models online / total
	db.DB.QueryRow("SELECT COUNT(*) FROM models WHERE status = 'online'").Scan(&resp.ModelsOnline)
	db.DB.QueryRow("SELECT COUNT(*) FROM models").Scan(&resp.ModelsTotal)
	db.DB.QueryRow("SELECT COUNT(*) FROM alerts WHERE acknowledged = 0 AND severity = 'critical' AND status != 'resolved'").Scan(&resp.CriticalAlertCount)
	db.DB.QueryRow("SELECT COUNT(*) FROM incidents WHERE status != 'resolved'").Scan(&resp.ActiveIncidentCount)

	var latestTraceAt sql.NullTime
	db.DB.QueryRow("SELECT MAX(created_at) FROM traces").Scan(&latestTraceAt)
	resp.LastUpdatedAt = now.Format(time.RFC3339)
	if latestTraceAt.Valid {
		resp.DataFreshnessSeconds = int(now.Sub(latestTraceAt.Time).Seconds())
		if resp.DataFreshnessSeconds < 0 {
			resp.DataFreshnessSeconds = 0
		}
	} else {
		resp.DataFreshnessSeconds = 0
	}

	healthPenalty := resp.CriticalAlertCount*18 + resp.ActiveIncidentCount*12
	if resp.ModelsTotal > 0 {
		offlineModels := resp.ModelsTotal - resp.ModelsOnline
		healthPenalty += offlineModels * 8
	}
	resp.HealthScore = 100 - healthPenalty
	if resp.HealthScore < 0 {
		resp.HealthScore = 0
	}

	var totalToday, okToday int64
	db.DB.QueryRow("SELECT COALESCE(COUNT(*), 0) FROM traces WHERE created_at >= ?", startOfDay).Scan(&totalToday)
	db.DB.QueryRow("SELECT COALESCE(COUNT(*), 0) FROM traces WHERE created_at >= ? AND status != 'error'", startOfDay).Scan(&okToday)
	attainment := 100.0
	if totalToday > 0 {
		attainment = float64(okToday) / float64(totalToday) * 100
	}
	budget := 100.0
	target := 99.0
	if attainment < target {
		budget = 0
	} else if target < 100 {
		budget = (attainment - target) / (100 - target) * 100
	}
	status := "healthy"
	if budget <= 0 || resp.CriticalAlertCount > 0 {
		status = "critical"
	} else if budget < 25 || resp.ActiveIncidentCount > 0 {
		status = "warning"
	}
	resp.SLOBurnSummary = models.SLOBurnSummary{
		TargetPercent:        target,
		AttainmentPercent:    attainment,
		ErrorBudgetRemaining: budget,
		BurnRate:             100 - budget,
		Status:               status,
	}

	// Token time series - hourly buckets last 24h
	resp.TokenTimeSeries = make([]models.TimeSeriesPoint, 0)
	for i := 23; i >= 0; i-- {
		bucketStart := now.Add(-time.Duration(i+1) * time.Hour)
		bucketEnd := now.Add(-time.Duration(i) * time.Hour)
		var val float64
		db.DB.QueryRow(
			"SELECT COALESCE(SUM(total_tokens), 0) FROM traces WHERE created_at >= ? AND created_at < ?",
			bucketStart, bucketEnd,
		).Scan(&val)
		resp.TokenTimeSeries = append(resp.TokenTimeSeries, models.TimeSeriesPoint{
			Time:  bucketStart.Format(time.RFC3339),
			Value: val,
		})
	}

	// Cost time series - daily buckets last 7 days
	resp.CostTimeSeries = make([]models.TimeSeriesPoint, 0)
	for i := 6; i >= 0; i-- {
		dayStart := time.Date(now.Year(), now.Month(), now.Day()-i, 0, 0, 0, 0, time.UTC)
		dayEnd := dayStart.Add(24 * time.Hour)
		var val float64
		db.DB.QueryRow(
			"SELECT COALESCE(SUM(total_cost), 0) FROM traces WHERE created_at >= ? AND created_at < ?",
			dayStart, dayEnd,
		).Scan(&val)
		resp.CostTimeSeries = append(resp.CostTimeSeries, models.TimeSeriesPoint{
			Time:  dayStart.Format("2006-01-02"),
			Value: val,
		})
	}

	// Top 5 models by request count
	resp.TopModels = make([]models.ModelUsageSummary, 0)
	topRows, err := db.DB.Query(`
		SELECT s.model_id, s.provider, COUNT(*) as req_count, COALESCE(SUM(s.total_tokens), 0) as tok_count
		FROM spans s
		WHERE s.type = 'llm' AND s.model_id != ''
		GROUP BY s.model_id, s.provider
		ORDER BY req_count DESC
		LIMIT 5
	`)
	if err == nil {
		defer topRows.Close()
		for topRows.Next() {
			var m models.ModelUsageSummary
			if err := topRows.Scan(&m.ModelName, &m.Provider, &m.RequestCount, &m.TokenCount); err == nil {
				resp.TopModels = append(resp.TopModels, m)
			}
		}
	}

	// Recent 10 traces
	resp.RecentTraces = make([]models.TraceResponse, 0)
	traceRows, err := db.DB.Query(`
		SELECT t.id, t.trace_id, t.session_id, t.name, t.status, t.total_duration_ms,
		       t.total_tokens, t.prompt_tokens, t.completion_tokens, t.total_cost,
		       t.input_preview, t.output_preview, t.created_at,
		       (SELECT COUNT(*) FROM spans WHERE trace_id = t.id) as span_count
		FROM traces t
		ORDER BY t.created_at DESC
		LIMIT 10
	`)
	if err == nil {
		defer traceRows.Close()
		for traceRows.Next() {
			var tr models.TraceResponse
			var createdAt time.Time
			if err := traceRows.Scan(
				&tr.ID, &tr.TraceID, &tr.SessionID, &tr.Name, &tr.Status,
				&tr.TotalDurationMs, &tr.TotalTokens, &tr.PromptTokens,
				&tr.CompletionTokens, &tr.TotalCost, &tr.InputPreview,
				&tr.OutputPreview, &createdAt, &tr.SpanCount,
			); err == nil {
				tr.CreatedAt = &createdAt

				// Get model info from first LLM span
				var modelName, provider string
				db.DB.QueryRow(
					"SELECT model_id, provider FROM spans WHERE trace_id = ? AND type = 'llm' LIMIT 1",
					tr.ID,
				).Scan(&modelName, &provider)
				tr.ModelName = modelName
				tr.Provider = provider

				resp.RecentTraces = append(resp.RecentTraces, tr)
			}
		}
	}

	// Active (unacknowledged) alerts
	resp.ActiveAlerts = make([]models.AlertResponse, 0)
	alertRows, err := db.DB.Query(`
		SELECT id, rule_id, severity, message, acknowledged, status, owner_id, service_id,
		       model_id, incident_id, dedupe_key, runbook_url, last_seen_at, resolved_at, created_at
		FROM alerts
		WHERE acknowledged = 0 AND status != 'resolved'
		ORDER BY CASE severity WHEN 'critical' THEN 0 WHEN 'warning' THEN 1 ELSE 2 END, created_at DESC
		LIMIT 20
	`)
	if err == nil {
		defer alertRows.Close()
		for alertRows.Next() {
			var a models.AlertResponse
			var ruleID sql.NullString
			var lastSeenAt, resolvedAt sql.NullTime
			var createdAt time.Time
			if err := alertRows.Scan(
				&a.ID, &ruleID, &a.Severity, &a.Message, &a.Acknowledged, &a.Status,
				&a.OwnerID, &a.ServiceID, &a.ModelID, &a.IncidentID, &a.DedupeKey,
				&a.RunbookURL, &lastSeenAt, &resolvedAt, &createdAt,
			); err == nil {
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
				resp.ActiveAlerts = append(resp.ActiveAlerts, a)
			}
		}
	}

	_ = fmt.Sprintf("") // avoid unused import
	WriteJSON(w, http.StatusOK, resp)
}
