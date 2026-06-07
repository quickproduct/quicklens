package workers

import (
	"log"
	"os"
	"strconv"
	"time"

	"quicklens/backend/db"
)

func StartJanitorWorker() {
	go func() {
		// Run once on startup after 5 seconds delay to let initialization settle
		time.Sleep(5 * time.Second)
		runJanitor()

		ticker := time.NewTicker(1 * time.Hour)
		log.Println("Janitor worker started, running every hour")

		for range ticker.C {
			runJanitor()
		}
	}()
}

func runJanitor() {
	log.Println("Janitor worker: starting data cleanup...")

	// Data retention days
	retentionDays := 30
	if envDays := os.Getenv("DATA_RETENTION_DAYS"); envDays != "" {
		if val, err := strconv.Atoi(envDays); err == nil && val > 0 {
			retentionDays = val
		}
	}

	cutoffTime := time.Now().UTC().AddDate(0, 0, -retentionDays)
	now := time.Now().UTC()

	// Delete old traces (cascade delete spans if foreign keys are ON, but let's delete spans explicitly too)
	res, err := db.DB.Exec("DELETE FROM spans WHERE started_at < ?", cutoffTime)
	if err != nil {
		log.Printf("Janitor error deleting old spans: %v", err)
	} else {
		affected, _ := res.RowsAffected()
		if affected > 0 {
			log.Printf("Janitor: deleted %d old spans", affected)
		}
	}

	res, err = db.DB.Exec("DELETE FROM traces WHERE created_at < ?", cutoffTime)
	if err != nil {
		log.Printf("Janitor error deleting old traces: %v", err)
	} else {
		affected, _ := res.RowsAffected()
		if affected > 0 {
			log.Printf("Janitor: deleted %d old traces", affected)
		}
	}

	// Delete expired auth sessions
	res, err = db.DB.Exec("DELETE FROM sessions WHERE expires_at < ?", now)
	if err != nil {
		log.Printf("Janitor error deleting expired sessions: %v", err)
	} else {
		affected, _ := res.RowsAffected()
		if affected > 0 {
			log.Printf("Janitor: deleted %d expired auth sessions", affected)
		}
	}

	// Run vacuum / incremental vacuum to reclaim space in SQLite
	_, err = db.DB.Exec("PRAGMA incremental_vacuum")
	if err != nil {
		log.Printf("Janitor warning running incremental_vacuum: %v", err)
	}

	log.Println("Janitor worker: data cleanup completed")
}
