package workers

import (
	"go.uber.org/zap"
	"time"

	"quicklens/backend/internal/db"
	"quicklens/backend/internal/ws"
)

func StartMetricsWorker() {
	go func() {
		ticker := time.NewTicker(60 * time.Second)
		zap.L().Sugar().Info("Metrics worker started, running every 60 seconds")

		// Run once immediately
		calculateMetrics()

		for range ticker.C {
			calculateMetrics()
		}
	}()
}

func calculateMetrics() {
	now := time.Now().UTC()
	oneMinuteAgo := now.Add(-60 * time.Second)

	var totalTokens int64
	err := db.DB.QueryRow(`
		SELECT COALESCE(SUM(total_tokens), 0)
		FROM spans
		WHERE started_at >= ?
	`, oneMinuteAgo).Scan(&totalTokens)
	if err != nil {
		zap.L().Sugar().Infof("Metrics worker: failed to query recent token count: %v", err)
		return
	}

	tokensPerSecond := float64(totalTokens) / 60.0

	// Broadcast token throughput metrics
	ws.Manager.Broadcast("traces", map[string]interface{}{
		"type": "metrics",
		"metrics": map[string]interface{}{
			"tokens_per_second": tokensPerSecond,
			"timestamp":         now.Format(time.RFC3339),
		},
	})
}
