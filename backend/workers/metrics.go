package workers

import (
	"log"
	"time"

	"quicklens/backend/db"
	"quicklens/backend/ws"
)

func StartMetricsWorker() {
	go func() {
		ticker := time.NewTicker(60 * time.Second)
		log.Println("Metrics worker started, running every 60 seconds")

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
		log.Printf("Metrics worker: failed to query recent token count: %v", err)
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
