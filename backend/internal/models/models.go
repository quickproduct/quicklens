package models

import "time"

// ─── Auth ───────────────────────────────────────────────────────────────────────

type User struct {
	ID             string    `json:"id"`
	Email          string    `json:"email"`
	HashedPassword string    `json:"-"`
	Role           string    `json:"role"`
	IsActive       bool      `json:"is_active"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type UserResponse struct {
	ID        string     `json:"id"`
	Email     string     `json:"email"`
	Role      string     `json:"role"`
	IsActive  bool       `json:"is_active"`
	CreatedAt *time.Time `json:"created_at"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// ─── Models ─────────────────────────────────────────────────────────────────────

type ModelResponse struct {
	ID            string     `json:"id"`
	Name          string     `json:"name"`
	Provider      string     `json:"provider"`
	ModelID       string     `json:"model_id"`
	Endpoint      string     `json:"endpoint"`
	Status        string     `json:"status"`
	Quantization  string     `json:"quantization"`
	SizeBytes     int64      `json:"size_bytes"`
	ContextLength int        `json:"context_length"`
	LastSeenAt    *time.Time `json:"last_seen_at"`
	CreatedAt     *time.Time `json:"created_at"`
	TotalRequests int64      `json:"total_requests"`
	AvgLatencyMs  float64    `json:"avg_latency_ms"`
	TotalTokens   int64      `json:"total_tokens"`
}

type ModelCreateRequest struct {
	Name          string `json:"name"`
	Provider      string `json:"provider"`
	ModelID       string `json:"model_id"`
	Endpoint      string `json:"endpoint"`
	ContextLength int    `json:"context_length"`
}

// ─── Traces ─────────────────────────────────────────────────────────────────────

type TraceResponse struct {
	ID               string     `json:"id"`
	TraceID          string     `json:"trace_id"`
	SessionID        string     `json:"session_id"`
	Name             string     `json:"name"`
	Status           string     `json:"status"`
	TotalDurationMs  int64      `json:"total_duration_ms"`
	DurationMs       int64      `json:"duration_ms"` // Alias for frontend compatibility
	TotalTokens      int64      `json:"total_tokens"`
	PromptTokens     int64      `json:"prompt_tokens"`
	CompletionTokens int64      `json:"completion_tokens"`
	TotalCost        float64    `json:"total_cost"`
	Cost             float64    `json:"cost"` // Alias for frontend compatibility
	InputPreview     string     `json:"input_preview"`
	OutputPreview    string     `json:"output_preview"`
	ModelName        string     `json:"model_name"`
	Model            string     `json:"model"` // Alias for frontend compatibility
	Provider         string     `json:"provider"`
	SpanCount        int        `json:"span_count"`
	CreatedAt        *time.Time `json:"created_at"`
}

type TraceDetailResponse struct {
	TraceResponse
	Spans []SpanResponse `json:"spans"`
}

type SpanResponse struct {
	ID               string                 `json:"id"`
	TraceID          string                 `json:"trace_id"`
	ParentSpanID     string                 `json:"parent_span_id"`
	Name             string                 `json:"name"`
	Type             string                 `json:"type"`
	ModelID          string                 `json:"model_id"`
	Provider         string                 `json:"provider"`
	Input            string                 `json:"input"`
	Output           string                 `json:"output"`
	PromptTokens     int64                  `json:"prompt_tokens"`
	CompletionTokens int64                  `json:"completion_tokens"`
	TotalTokens      int64                  `json:"total_tokens"`
	Cost             float64                `json:"cost"`
	DurationMs       int64                  `json:"duration_ms"`
	Status           string                 `json:"status"`
	ErrorMessage     string                 `json:"error_message"`
	Metadata         map[string]interface{} `json:"metadata"`
	StartedAt        *time.Time             `json:"started_at"`
	EndedAt          *time.Time             `json:"ended_at"`
	Children         []SpanResponse         `json:"children"`
}

type TraceIngestRequest struct {
	TraceID   string              `json:"trace_id"`
	SessionID string              `json:"session_id"`
	Name      string              `json:"name"`
	Spans     []SpanIngestRequest `json:"spans"`
}

type SpanIngestRequest struct {
	SpanID           string                 `json:"span_id"`
	ParentSpanID     string                 `json:"parent_span_id"`
	Name             string                 `json:"name"`
	Type             string                 `json:"type"`
	ModelID          string                 `json:"model_id"`
	Provider         string                 `json:"provider"`
	Input            string                 `json:"input"`
	Output           string                 `json:"output"`
	PromptTokens     int64                  `json:"prompt_tokens"`
	CompletionTokens int64                  `json:"completion_tokens"`
	DurationMs       int64                  `json:"duration_ms"`
	Status           string                 `json:"status"`
	ErrorMessage     string                 `json:"error_message"`
	Metadata         map[string]interface{} `json:"metadata"`
	StartedAt        *time.Time             `json:"started_at"`
	EndedAt          *time.Time             `json:"ended_at"`
}

// ─── Prompts ────────────────────────────────────────────────────────────────────

type PromptResponse struct {
	ID        string                 `json:"id"`
	Name      string                 `json:"name"`
	Content   string                 `json:"content"`
	ModelID   string                 `json:"model_id"`
	Version   int                    `json:"version"`
	Tags      []string               `json:"tags"`
	CreatedAt *time.Time             `json:"created_at"`
	Versions  []PromptVersionSummary `json:"versions"`
}

type PromptVersionSummary struct {
	ID        string     `json:"id"`
	Version   int        `json:"version"`
	CreatedAt *time.Time `json:"created_at"`
}

type PromptCreateRequest struct {
	Name    string   `json:"name"`
	Content string   `json:"content"`
	ModelID string   `json:"model_id"`
	Tags    []string `json:"tags"`
}

type PromptDiffResponse struct {
	VersionA int    `json:"version_a"`
	VersionB int    `json:"version_b"`
	ContentA string `json:"content_a"`
	ContentB string `json:"content_b"`
}

// ─── Evaluations ────────────────────────────────────────────────────────────────

type EvalResponse struct {
	ID           string     `json:"id"`
	SpanID       string     `json:"span_id"`
	ScoreType    string     `json:"score_type"`
	ScoreValue   float64    `json:"score_value"`
	FeedbackText string     `json:"feedback_text"`
	Evaluator    string     `json:"evaluator"`
	CreatedAt    *time.Time `json:"created_at"`
}

type EvalCreateRequest struct {
	SpanID       string  `json:"span_id"`
	ScoreType    string  `json:"score_type"`
	ScoreValue   float64 `json:"score_value"`
	FeedbackText string  `json:"feedback_text"`
	Evaluator    string  `json:"evaluator"`
}

// ─── Dashboard ──────────────────────────────────────────────────────────────────

type DashboardResponse struct {
	TotalTracesToday     int64               `json:"total_traces_today"`
	TotalTokensToday     int64               `json:"total_tokens_today"`
	TotalCostToday       float64             `json:"total_cost_today"`
	AvgLatencyMs         float64             `json:"avg_latency_ms"`
	SuccessRateToday     float64             `json:"success_rate_today"`
	ModelsOnline         int                 `json:"models_online"`
	ModelsTotal          int                 `json:"models_total"`
	LastUpdatedAt        string              `json:"last_updated_at"`
	DataFreshnessSeconds int                 `json:"data_freshness_seconds"`
	TokenTimeSeries      []TimeSeriesPoint   `json:"token_time_series"`
	CostTimeSeries       []TimeSeriesPoint   `json:"cost_time_series"`
	TopModels            []ModelUsageSummary `json:"top_models"`
	RecentTraces         []TraceResponse     `json:"recent_traces"`
}

type TimeSeriesPoint struct {
	Time  string  `json:"time"`
	Value float64 `json:"value"`
}

type ModelUsageSummary struct {
	ModelName    string `json:"model_name"`
	Provider     string `json:"provider"`
	RequestCount int64  `json:"request_count"`
	TokenCount   int64  `json:"token_count"`
}

// ─── Logs ───────────────────────────────────────────────────────────────────────

type LogEntry struct {
	ID               string     `json:"id"`
	TraceID          string     `json:"trace_id"`
	SpanID           string     `json:"span_id"`
	ModelName        string     `json:"model_name"`
	Provider         string     `json:"provider"`
	Status           string     `json:"status"`
	DurationMs       int64      `json:"duration_ms"`
	PromptTokens     int64      `json:"prompt_tokens"`
	CompletionTokens int64      `json:"completion_tokens"`
	InputPreview     string     `json:"input_preview"`
	OutputPreview    string     `json:"output_preview"`
	ErrorMessage     string     `json:"error_message"`
	CreatedAt        *time.Time `json:"created_at"`
}

// ─── Sessions ───────────────────────────────────────────────────────────────────

type SessionResponse struct {
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	TraceCount int        `json:"trace_count"`
	CreatedAt  *time.Time `json:"created_at"`
}
