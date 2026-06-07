package db

import (
	"database/sql"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	_ "modernc.org/sqlite"

	"quicklens/backend/auth"
)

var DB *sql.DB

func InitDB() {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = os.Getenv("SQLITE_DB_PATH")
	}
	if dbPath == "" {
		dbPath = "quicklens.db"
	}

	var err error
	DB, err = sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}

	// Configure SQLite
	pragmas := []string{
		"PRAGMA journal_mode=WAL",
		"PRAGMA synchronous=NORMAL",
		"PRAGMA cache_size=-64000",
		"PRAGMA busy_timeout=5000",
		"PRAGMA foreign_keys=ON",
		"PRAGMA temp_store=MEMORY",
	}
	for _, p := range pragmas {
		if _, err := DB.Exec(p); err != nil {
			log.Printf("Warning: failed to set %s: %v", p, err)
		}
	}

	DB.SetMaxOpenConns(10)
	DB.SetMaxIdleConns(10)

	createTables()
	seedAdmin()

	log.Println("Database initialized successfully")
}

func createTables() {
	tables := `
	CREATE TABLE IF NOT EXISTS users (
		id TEXT PRIMARY KEY,
		email TEXT UNIQUE NOT NULL,
		hashed_password TEXT NOT NULL,
		role TEXT NOT NULL DEFAULT 'user',
		is_active INTEGER NOT NULL DEFAULT 1,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS sessions (
		id TEXT PRIMARY KEY,
		user_id TEXT NOT NULL,
		refresh_token TEXT NOT NULL,
		expires_at DATETIME NOT NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS models (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		provider TEXT NOT NULL,
		model_id TEXT NOT NULL DEFAULT '',
		endpoint TEXT NOT NULL DEFAULT '',
		status TEXT NOT NULL DEFAULT 'unknown',
		quantization TEXT NOT NULL DEFAULT '',
		size_bytes INTEGER NOT NULL DEFAULT 0,
		context_length INTEGER NOT NULL DEFAULT 0,
		last_seen_at DATETIME,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS traces (
		id TEXT PRIMARY KEY,
		trace_id TEXT NOT NULL DEFAULT '',
		session_id TEXT NOT NULL DEFAULT '',
		name TEXT NOT NULL DEFAULT '',
		status TEXT NOT NULL DEFAULT 'ok',
		total_duration_ms INTEGER NOT NULL DEFAULT 0,
		total_tokens INTEGER NOT NULL DEFAULT 0,
		prompt_tokens INTEGER NOT NULL DEFAULT 0,
		completion_tokens INTEGER NOT NULL DEFAULT 0,
		total_cost REAL NOT NULL DEFAULT 0.0,
		input_preview TEXT NOT NULL DEFAULT '',
		output_preview TEXT NOT NULL DEFAULT '',
		metadata TEXT NOT NULL DEFAULT '{}',
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS spans (
		id TEXT PRIMARY KEY,
		trace_id TEXT NOT NULL,
		parent_span_id TEXT NOT NULL DEFAULT '',
		name TEXT NOT NULL DEFAULT '',
		type TEXT NOT NULL DEFAULT 'llm',
		model_id TEXT NOT NULL DEFAULT '',
		provider TEXT NOT NULL DEFAULT '',
		input TEXT NOT NULL DEFAULT '',
		output TEXT NOT NULL DEFAULT '',
		prompt_tokens INTEGER NOT NULL DEFAULT 0,
		completion_tokens INTEGER NOT NULL DEFAULT 0,
		total_tokens INTEGER NOT NULL DEFAULT 0,
		cost REAL NOT NULL DEFAULT 0.0,
		duration_ms INTEGER NOT NULL DEFAULT 0,
		status TEXT NOT NULL DEFAULT 'ok',
		error_message TEXT NOT NULL DEFAULT '',
		metadata TEXT NOT NULL DEFAULT '{}',
		started_at DATETIME,
		ended_at DATETIME,
		FOREIGN KEY (trace_id) REFERENCES traces(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS prompts (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		content TEXT NOT NULL DEFAULT '',
		model_id TEXT NOT NULL DEFAULT '',
		version INTEGER NOT NULL DEFAULT 1,
		parent_id TEXT NOT NULL DEFAULT '',
		tags TEXT NOT NULL DEFAULT '[]',
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS evaluations (
		id TEXT PRIMARY KEY,
		span_id TEXT NOT NULL,
		score_type TEXT NOT NULL DEFAULT '',
		score_value REAL NOT NULL DEFAULT 0.0,
		feedback_text TEXT NOT NULL DEFAULT '',
		evaluator TEXT NOT NULL DEFAULT '',
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS alert_rules (
		id TEXT PRIMARY KEY,
		metric_type TEXT NOT NULL,
		threshold REAL NOT NULL DEFAULT 0.0,
		operator TEXT NOT NULL DEFAULT '>',
		window_seconds INTEGER NOT NULL DEFAULT 300,
		enabled INTEGER NOT NULL DEFAULT 1,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS alerts (
		id TEXT PRIMARY KEY,
		rule_id TEXT,
		severity TEXT NOT NULL DEFAULT 'warning',
		message TEXT NOT NULL DEFAULT '',
		acknowledged INTEGER NOT NULL DEFAULT 0,
		status TEXT NOT NULL DEFAULT 'open',
		owner_id TEXT NOT NULL DEFAULT '',
		service_id TEXT NOT NULL DEFAULT '',
		model_id TEXT NOT NULL DEFAULT '',
		incident_id TEXT NOT NULL DEFAULT '',
		dedupe_key TEXT NOT NULL DEFAULT '',
		runbook_url TEXT NOT NULL DEFAULT '',
		last_seen_at DATETIME,
		resolved_at DATETIME,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS incidents (
		id TEXT PRIMARY KEY,
		title TEXT NOT NULL,
		severity TEXT NOT NULL DEFAULT 'warning',
		status TEXT NOT NULL DEFAULT 'investigating',
		owner_id TEXT NOT NULL DEFAULT '',
		service_id TEXT NOT NULL DEFAULT '',
		model_id TEXT NOT NULL DEFAULT '',
		alert_id TEXT NOT NULL DEFAULT '',
		summary TEXT NOT NULL DEFAULT '',
		runbook_url TEXT NOT NULL DEFAULT '',
		started_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		resolved_at DATETIME,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS incident_events (
		id TEXT PRIMARY KEY,
		incident_id TEXT NOT NULL,
		event_type TEXT NOT NULL,
		message TEXT NOT NULL DEFAULT '',
		actor_id TEXT NOT NULL DEFAULT '',
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (incident_id) REFERENCES incidents(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS comments (
		id TEXT PRIMARY KEY,
		resource_type TEXT NOT NULL,
		resource_id TEXT NOT NULL,
		author_id TEXT NOT NULL DEFAULT '',
		body TEXT NOT NULL DEFAULT '',
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS audit_logs (
		id TEXT PRIMARY KEY,
		actor_id TEXT NOT NULL DEFAULT '',
		action TEXT NOT NULL,
		resource TEXT NOT NULL,
		resource_id TEXT NOT NULL DEFAULT '',
		metadata TEXT NOT NULL DEFAULT '{}',
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS saved_views (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		scope TEXT NOT NULL,
		filters TEXT NOT NULL DEFAULT '{}',
		is_shared INTEGER NOT NULL DEFAULT 0,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS slo_definitions (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		service_id TEXT NOT NULL DEFAULT '',
		target_percent REAL NOT NULL DEFAULT 99.0,
		period_days INTEGER NOT NULL DEFAULT 28,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS slo_windows (
		id TEXT PRIMARY KEY,
		slo_id TEXT NOT NULL,
		good_events INTEGER NOT NULL DEFAULT 0,
		total_events INTEGER NOT NULL DEFAULT 0,
		window_start DATETIME NOT NULL,
		window_end DATETIME NOT NULL,
		FOREIGN KEY (slo_id) REFERENCES slo_definitions(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS notification_rules (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		channel TEXT NOT NULL DEFAULT 'email',
		target TEXT NOT NULL DEFAULT '',
		enabled INTEGER NOT NULL DEFAULT 1,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS dashboard_layouts (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		layout_json TEXT NOT NULL DEFAULT '{}',
		is_default INTEGER NOT NULL DEFAULT 0,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS model_prices (
		id TEXT PRIMARY KEY,
		provider TEXT NOT NULL,
		model_id TEXT NOT NULL,
		prompt_price_per_1k REAL NOT NULL DEFAULT 0.0,
		completion_price_per_1k REAL NOT NULL DEFAULT 0.0,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS llm_sessions (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL DEFAULT '',
		metadata TEXT NOT NULL DEFAULT '{}',
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	-- Indexes
	CREATE INDEX IF NOT EXISTS idx_traces_created_at ON traces(created_at DESC);
	CREATE INDEX IF NOT EXISTS idx_spans_trace_id ON spans(trace_id);
	CREATE INDEX IF NOT EXISTS idx_spans_started_at ON spans(started_at DESC);
	CREATE INDEX IF NOT EXISTS idx_models_provider_model_id ON models(provider, model_id);
	CREATE INDEX IF NOT EXISTS idx_alerts_created_at ON alerts(created_at DESC);
	CREATE INDEX IF NOT EXISTS idx_alerts_status_severity ON alerts(status, severity);
	CREATE INDEX IF NOT EXISTS idx_incidents_status_created_at ON incidents(status, created_at DESC);
	CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON audit_logs(created_at DESC);
	CREATE INDEX IF NOT EXISTS idx_saved_views_scope ON saved_views(scope);
	CREATE INDEX IF NOT EXISTS idx_prompts_name ON prompts(name);
	`

	if _, err := DB.Exec(tables); err != nil {
		log.Fatalf("Failed to create tables: %v", err)
	}

	additiveMigrations()
}

func additiveMigrations() {
	columns := []struct {
		name string
		stmt string
	}{
		{"status", "ALTER TABLE alerts ADD COLUMN status TEXT NOT NULL DEFAULT 'open'"},
		{"owner_id", "ALTER TABLE alerts ADD COLUMN owner_id TEXT NOT NULL DEFAULT ''"},
		{"service_id", "ALTER TABLE alerts ADD COLUMN service_id TEXT NOT NULL DEFAULT ''"},
		{"model_id", "ALTER TABLE alerts ADD COLUMN model_id TEXT NOT NULL DEFAULT ''"},
		{"incident_id", "ALTER TABLE alerts ADD COLUMN incident_id TEXT NOT NULL DEFAULT ''"},
		{"dedupe_key", "ALTER TABLE alerts ADD COLUMN dedupe_key TEXT NOT NULL DEFAULT ''"},
		{"runbook_url", "ALTER TABLE alerts ADD COLUMN runbook_url TEXT NOT NULL DEFAULT ''"},
		{"last_seen_at", "ALTER TABLE alerts ADD COLUMN last_seen_at DATETIME"},
		{"resolved_at", "ALTER TABLE alerts ADD COLUMN resolved_at DATETIME"},
	}
	for _, column := range columns {
		if columnExists("alerts", column.name) {
			continue
		}
		if _, err := DB.Exec(column.stmt); err != nil {
			log.Printf("Migration failed: %s (%v)", column.stmt, err)
		}
	}
}

func columnExists(table, column string) bool {
	rows, err := DB.Query("PRAGMA table_info(" + table + ")")
	if err != nil {
		return false
	}
	defer rows.Close()

	for rows.Next() {
		var cid int
		var name, dataType string
		var notNull int
		var defaultValue interface{}
		var pk int
		if err := rows.Scan(&cid, &name, &dataType, &notNull, &defaultValue, &pk); err != nil {
			continue
		}
		if name == column {
			return true
		}
	}
	return false
}

func seedAdmin() {
	email := os.Getenv("DEFAULT_ADMIN_EMAIL")
	password := os.Getenv("DEFAULT_ADMIN_PASSWORD")
	if email == "" {
		email = "admin@quicklens.dev"
	}
	if password == "" {
		password = "admin123"
	}

	var exists int
	err := DB.QueryRow("SELECT COUNT(*) FROM users WHERE email = ?", email).Scan(&exists)
	if err != nil {
		log.Printf("Warning: failed to check admin user: %v", err)
		return
	}
	if exists > 0 {
		return
	}

	hashedPassword, err := auth.HashPassword(password)
	if err != nil {
		log.Printf("Warning: failed to hash admin password: %v", err)
		return
	}

	id := uuid.New().String()
	now := time.Now().UTC()
	_, err = DB.Exec(
		"INSERT INTO users (id, email, hashed_password, role, is_active, created_at, updated_at) VALUES (?, ?, ?, 'admin', 1, ?, ?)",
		id, email, hashedPassword, now, now,
	)
	if err != nil {
		log.Printf("Warning: failed to create admin user: %v", err)
		return
	}
	log.Printf("Admin user created: %s", email)
}
