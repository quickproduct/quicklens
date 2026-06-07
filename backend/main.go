package main

import (
	"embed"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strings"

	"quicklens/backend/db"
	"quicklens/backend/handlers"
	"quicklens/backend/workers"
	"quicklens/backend/ws"
)

//go:embed all:frontend/build
var staticFS embed.FS

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, HEAD")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		log.Printf("HTTP Request: %s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func main() {
	// Initialize Database
	db.InitDB()

	// Start WebSocket Heartbeat
	ws.Manager.StartHeartbeat()

	// Start Background Tasks / Workers
	handlers.StartIngestWorker()
	workers.StartOllamaWorker()
	workers.StartAlertsWorker()
	workers.StartMetricsWorker()
	workers.StartJanitorWorker()

	mux := http.NewServeMux()

	// Health check
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Auth API
	mux.HandleFunc("POST /api/v1/auth/login", handlers.LoginHandler)
	mux.HandleFunc("POST /api/v1/auth/logout", handlers.LogoutHandler)
	mux.HandleFunc("POST /api/v1/auth/refresh", handlers.RefreshHandler)
	mux.HandleFunc("GET /api/v1/me", handlers.AuthMiddleware(handlers.MeHandler))
	mux.HandleFunc("GET /api/v1/auth/me", handlers.AuthMiddleware(handlers.MeHandler))
	mux.HandleFunc("PUT /api/v1/auth/password", handlers.AuthMiddleware(handlers.ChangePasswordHandler))
	mux.HandleFunc("POST /api/v1/auth/register", handlers.RegisterHandler)

	// Dashboard API
	mux.HandleFunc("GET /api/v1/dashboard", handlers.AuthMiddleware(handlers.GetDashboardHandler))

	// Models API
	mux.HandleFunc("GET /api/v1/models", handlers.AuthMiddleware(handlers.ListModelsHandler))
	mux.HandleFunc("POST /api/v1/models", handlers.AuthMiddleware(handlers.CreateModelHandler))
	mux.HandleFunc("GET /api/v1/models/{id}", handlers.AuthMiddleware(handlers.GetModelHandler))
	mux.HandleFunc("DELETE /api/v1/models/{id}", handlers.AuthMiddleware(handlers.DeleteModelHandler))
	mux.HandleFunc("POST /api/v1/models/discover", handlers.AuthMiddleware(handlers.DiscoverModelsHandler))

	// Traces API
	mux.HandleFunc("GET /api/v1/traces", handlers.AuthMiddleware(handlers.ListTracesHandler))
	mux.HandleFunc("GET /api/v1/traces/{id}", handlers.AuthMiddleware(handlers.GetTraceHandler))
	mux.HandleFunc("DELETE /api/v1/traces/{id}", handlers.AuthMiddleware(handlers.DeleteTraceHandler))

	// Sessions API
	mux.HandleFunc("GET /api/v1/sessions", handlers.AuthMiddleware(handlers.ListSessionsHandler))
	mux.HandleFunc("GET /api/v1/sessions/{id}/traces", handlers.AuthMiddleware(handlers.GetSessionTracesHandler))

	// Ingestion API
	mux.HandleFunc("POST /api/v1/ingest/traces", handlers.IngestTraceHandler)
	mux.HandleFunc("POST /api/v1/ingest/otel", handlers.IngestOTelHandler)

	// Prompts API
	mux.HandleFunc("GET /api/v1/prompts", handlers.AuthMiddleware(handlers.ListPromptsHandler))
	mux.HandleFunc("POST /api/v1/prompts", handlers.AuthMiddleware(handlers.CreatePromptHandler))
	mux.HandleFunc("GET /api/v1/prompts/{id}", handlers.AuthMiddleware(handlers.GetPromptHandler))
	mux.HandleFunc("GET /api/v1/prompts/{id}/diff/{version}", handlers.AuthMiddleware(handlers.DiffPromptVersionsHandler))
	mux.HandleFunc("DELETE /api/v1/prompts/{id}", handlers.AuthMiddleware(handlers.DeletePromptHandler))

	// Evaluations API
	mux.HandleFunc("POST /api/v1/evaluations", handlers.AuthMiddleware(handlers.CreateEvaluationHandler))
	mux.HandleFunc("GET /api/v1/evaluations", handlers.AuthMiddleware(handlers.ListEvaluationsHandler))
	mux.HandleFunc("GET /api/v1/spans/{id}/evaluations", handlers.AuthMiddleware(handlers.GetSpanEvaluationsHandler))

	// Alerts API
	mux.HandleFunc("GET /api/v1/alerts", handlers.AuthMiddleware(handlers.ListAlertsHandler))
	mux.HandleFunc("POST /api/v1/alerts/{alert_id}/acknowledge", handlers.AuthMiddleware(handlers.AcknowledgeAlertHandler))
	mux.HandleFunc("GET /api/v1/alert-rules", handlers.AuthMiddleware(handlers.ListRulesHandler))
	mux.HandleFunc("POST /api/v1/alert-rules", handlers.AuthMiddleware(handlers.CreateRuleHandler))
	mux.HandleFunc("PUT /api/v1/alert-rules/{rule_id}", handlers.AuthMiddleware(handlers.UpdateRuleHandler))
	mux.HandleFunc("DELETE /api/v1/alert-rules/{rule_id}", handlers.AuthMiddleware(handlers.DeleteRuleHandler))

	// Operations API
	mux.HandleFunc("GET /api/v1/incidents", handlers.AuthMiddleware(handlers.ListIncidentsHandler))
	mux.HandleFunc("POST /api/v1/incidents", handlers.AuthMiddleware(handlers.CreateIncidentHandler))
	mux.HandleFunc("PUT /api/v1/incidents/{incident_id}", handlers.AuthMiddleware(handlers.UpdateIncidentHandler))
	mux.HandleFunc("GET /api/v1/incidents/{incident_id}/events", handlers.AuthMiddleware(handlers.ListIncidentEventsHandler))
	mux.HandleFunc("GET /api/v1/audit-logs", handlers.AuthMiddleware(handlers.ListAuditLogsHandler))
	mux.HandleFunc("GET /api/v1/saved-views", handlers.AuthMiddleware(handlers.ListSavedViewsHandler))
	mux.HandleFunc("GET /api/v1/slo-definitions", handlers.AuthMiddleware(handlers.ListSLODefinitionsHandler))
	mux.HandleFunc("GET /api/v1/notification-rules", handlers.AuthMiddleware(handlers.ListNotificationRulesHandler))
	mux.HandleFunc("GET /api/v1/dashboard-layouts", handlers.AuthMiddleware(handlers.ListDashboardLayoutsHandler))

	// Logs API
	mux.HandleFunc("GET /api/v1/logs", handlers.AuthMiddleware(handlers.SearchLogsHandler))

	// Proxy endpoints (transparent)
	mux.HandleFunc("POST /v1/chat/completions", handlers.ProxyChatCompletions)
	mux.HandleFunc("POST /v1/completions", handlers.ProxyCompletions)
	mux.HandleFunc("GET /v1/models", handlers.ProxyModels)
	mux.HandleFunc("POST /ollama/api/generate", handlers.ProxyOllamaGenerate)
	mux.HandleFunc("POST /ollama/api/chat", handlers.ProxyOllamaChat)
	mux.HandleFunc("GET /ollama/api/tags", handlers.ProxyOllamaTags)

	// WebSockets
	mux.HandleFunc("GET /ws/logs", handlers.HandleWSChannel("logs"))
	mux.HandleFunc("GET /ws/models", handlers.HandleWSChannel("models"))
	mux.HandleFunc("GET /ws/traces", handlers.HandleWSChannel("traces"))
	mux.HandleFunc("GET /ws/alerts", handlers.HandleWSChannel("alerts"))

	// Embedded static frontend configuration
	fSys, err := fs.Sub(staticFS, "frontend/build")
	if err != nil {
		log.Fatalf("Failed to initialize embedded static assets: %v", err)
	}

	fileServer := http.FileServer(http.FS(fSys))

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		// Let specific router handles handle their exact paths. If any of these prefixes fall through here, return 404
		if strings.HasPrefix(path, "/api/") || strings.HasPrefix(path, "/ws/") || strings.HasPrefix(path, "/v1/") || strings.HasPrefix(path, "/ollama/") || path == "/health" {
			http.NotFound(w, r)
			return
		}

		// Try opening the file in embedded FS
		f, err := fSys.Open(strings.TrimPrefix(path, "/"))
		if err != nil {
			// Fallback to index.html for clientside SPA routing
			indexFile, err := fSys.Open("index.html")
			if err != nil {
				http.Error(w, "Frontend Build Not Found", http.StatusNotFound)
				return
			}
			defer indexFile.Close()
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			io.Copy(w, indexFile)
			return
		}
		defer f.Close()

		fileServer.ServeHTTP(w, r)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	log.Printf("Starting QuickLens backend on :%s", port)
	if err := http.ListenAndServe(":"+port, corsMiddleware(mux)); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
