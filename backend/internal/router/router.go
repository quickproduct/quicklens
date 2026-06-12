// Package router assembles the chi router with middleware and all routes.
package router

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"

	"quicklens/backend/internal/handlers"
	"quicklens/backend/internal/httpx"
	"quicklens/backend/internal/static"
)

// New builds the application router.
func New(logger *zap.Logger) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(httpx.CORS)
	r.Use(httpx.RequestLogger(logger))

	auth := handlers.AuthMiddleware

	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	// ── Auth ───────────────────────────────────────────────────────────────
	r.Post("/api/v1/auth/login", handlers.LoginHandler)
	r.Post("/api/v1/auth/logout", handlers.LogoutHandler)
	r.Post("/api/v1/auth/refresh", handlers.RefreshHandler)
	r.Post("/api/v1/auth/register", handlers.RegisterHandler)
	r.Get("/api/v1/me", auth(handlers.MeHandler))
	r.Get("/api/v1/auth/me", auth(handlers.MeHandler))
	r.Put("/api/v1/auth/password", auth(handlers.ChangePasswordHandler))

	// ── Dashboard ──────────────────────────────────────────────────────────
	r.Get("/api/v1/dashboard", auth(handlers.GetDashboardHandler))

	// ── Models ─────────────────────────────────────────────────────────────
	r.Get("/api/v1/models", auth(handlers.ListModelsHandler))
	r.Post("/api/v1/models", auth(handlers.CreateModelHandler))
	r.Get("/api/v1/models/{id}", auth(handlers.GetModelHandler))
	r.Delete("/api/v1/models/{id}", auth(handlers.DeleteModelHandler))
	r.Post("/api/v1/models/discover", auth(handlers.DiscoverModelsHandler))

	// ── Traces ─────────────────────────────────────────────────────────────
	r.Get("/api/v1/traces", auth(handlers.ListTracesHandler))
	r.Get("/api/v1/traces/{id}", auth(handlers.GetTraceHandler))
	r.Delete("/api/v1/traces/{id}", auth(handlers.DeleteTraceHandler))

	// ── Sessions ───────────────────────────────────────────────────────────
	r.Get("/api/v1/sessions", auth(handlers.ListSessionsHandler))
	r.Get("/api/v1/sessions/{id}/traces", auth(handlers.GetSessionTracesHandler))

	// ── Ingestion (no auth, SDK/proxy facing) ──────────────────────────────
	r.Post("/api/v1/ingest/traces", handlers.IngestTraceHandler)
	r.Post("/api/v1/ingest/otel", handlers.IngestOTelHandler)

	// ── Evaluations ────────────────────────────────────────────────────────
	r.Post("/api/v1/evaluations", auth(handlers.CreateEvaluationHandler))
	r.Get("/api/v1/spans/{id}/evaluations", auth(handlers.GetSpanEvaluationsHandler))

	// ── Logs ───────────────────────────────────────────────────────────────
	r.Get("/api/v1/logs", auth(handlers.SearchLogsHandler))

	// ── Transparent proxy (no auth) ────────────────────────────────────────
	r.Post("/v1/chat/completions", handlers.ProxyChatCompletions)
	r.Post("/v1/completions", handlers.ProxyCompletions)
	r.Get("/v1/models", handlers.ProxyModels)
	r.Post("/ollama/api/generate", handlers.ProxyOllamaGenerate)
	r.Post("/ollama/api/chat", handlers.ProxyOllamaChat)
	r.Get("/ollama/api/tags", handlers.ProxyOllamaTags)

	// ── WebSockets ─────────────────────────────────────────────────────────
	r.Get("/ws/logs", handlers.HandleWSChannel("logs"))
	r.Get("/ws/models", handlers.HandleWSChannel("models"))
	r.Get("/ws/traces", handlers.HandleWSChannel("traces"))
	r.Get("/ws/alerts", handlers.HandleWSChannel("alerts"))

	// ── SPA fallback ───────────────────────────────────────────────────────
	spa := static.Handler()
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		p := r.URL.Path
		if strings.HasPrefix(p, "/api/") || strings.HasPrefix(p, "/ws/") ||
			strings.HasPrefix(p, "/v1/") || strings.HasPrefix(p, "/ollama/") {
			http.NotFound(w, r)
			return
		}
		spa.ServeHTTP(w, r)
	})

	return r
}
