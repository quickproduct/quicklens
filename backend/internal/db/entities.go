package db

import "time"

// GORM entity models owning the schema via AutoMigrate. Column tags are pinned
// to the exact names the existing handler/worker SQL expects, so the raw-SQL
// path (running on the same glebarez/modernc connection) sees an identical
// schema.

type UserEntity struct {
	ID             string    `gorm:"column:id;primaryKey"`
	Email          string    `gorm:"column:email;uniqueIndex;not null"`
	HashedPassword string    `gorm:"column:hashed_password;not null"`
	Role           string    `gorm:"column:role;not null;default:user"`
	IsActive       bool      `gorm:"column:is_active;not null;default:true"`
	CreatedAt      time.Time `gorm:"column:created_at;type:datetime;default:CURRENT_TIMESTAMP"`
	UpdatedAt      time.Time `gorm:"column:updated_at;type:datetime;default:CURRENT_TIMESTAMP"`
}

func (UserEntity) TableName() string { return "users" }

type SessionEntity struct {
	ID           string    `gorm:"column:id;primaryKey"`
	UserID       string    `gorm:"column:user_id;not null"`
	RefreshToken string    `gorm:"column:refresh_token;not null"`
	ExpiresAt    time.Time `gorm:"column:expires_at;not null"`
	CreatedAt    time.Time `gorm:"column:created_at;type:datetime;default:CURRENT_TIMESTAMP"`
}

func (SessionEntity) TableName() string { return "sessions" }

type ModelEntity struct {
	ID            string     `gorm:"column:id;primaryKey"`
	Name          string     `gorm:"column:name;not null"`
	Provider      string     `gorm:"column:provider;not null;index:idx_models_provider_model_id"`
	ModelID       string     `gorm:"column:model_id;not null;default:'';index:idx_models_provider_model_id"`
	Endpoint      string     `gorm:"column:endpoint;not null;default:''"`
	Status        string     `gorm:"column:status;not null;default:unknown"`
	Quantization  string     `gorm:"column:quantization;not null;default:''"`
	SizeBytes     int64      `gorm:"column:size_bytes;not null;default:0"`
	ContextLength int        `gorm:"column:context_length;not null;default:0"`
	LastSeenAt    *time.Time `gorm:"column:last_seen_at"`
	CreatedAt     time.Time  `gorm:"column:created_at;type:datetime;default:CURRENT_TIMESTAMP"`
	UpdatedAt     time.Time  `gorm:"column:updated_at;type:datetime;default:CURRENT_TIMESTAMP"`
}

func (ModelEntity) TableName() string { return "models" }

type TraceEntity struct {
	ID               string    `gorm:"column:id;primaryKey"`
	TraceID          string    `gorm:"column:trace_id;not null;default:''"`
	SessionID        string    `gorm:"column:session_id;not null;default:''"`
	Name             string    `gorm:"column:name;not null;default:''"`
	Status           string    `gorm:"column:status;not null;default:ok"`
	TotalDurationMs  int64     `gorm:"column:total_duration_ms;not null;default:0"`
	TotalTokens      int64     `gorm:"column:total_tokens;not null;default:0"`
	PromptTokens     int64     `gorm:"column:prompt_tokens;not null;default:0"`
	CompletionTokens int64     `gorm:"column:completion_tokens;not null;default:0"`
	TotalCost        float64   `gorm:"column:total_cost;not null;default:0"`
	InputPreview     string    `gorm:"column:input_preview;not null;default:''"`
	OutputPreview    string    `gorm:"column:output_preview;not null;default:''"`
	Metadata         string    `gorm:"column:metadata;not null;default:'{}'"`
	CreatedAt        time.Time `gorm:"column:created_at;type:datetime;default:CURRENT_TIMESTAMP;index:idx_traces_created_at,sort:desc"`
}

func (TraceEntity) TableName() string { return "traces" }

type SpanEntity struct {
	ID               string     `gorm:"column:id;primaryKey"`
	TraceID          string     `gorm:"column:trace_id;not null;index:idx_spans_trace_id"`
	ParentSpanID     string     `gorm:"column:parent_span_id;not null;default:''"`
	Name             string     `gorm:"column:name;not null;default:''"`
	Type             string     `gorm:"column:type;not null;default:llm"`
	ModelID          string     `gorm:"column:model_id;not null;default:''"`
	Provider         string     `gorm:"column:provider;not null;default:''"`
	Input            string     `gorm:"column:input;not null;default:''"`
	Output           string     `gorm:"column:output;not null;default:''"`
	PromptTokens     int64      `gorm:"column:prompt_tokens;not null;default:0"`
	CompletionTokens int64      `gorm:"column:completion_tokens;not null;default:0"`
	TotalTokens      int64      `gorm:"column:total_tokens;not null;default:0"`
	Cost             float64    `gorm:"column:cost;not null;default:0"`
	DurationMs       int64      `gorm:"column:duration_ms;not null;default:0"`
	Status           string     `gorm:"column:status;not null;default:ok"`
	ErrorMessage     string     `gorm:"column:error_message;not null;default:''"`
	Metadata         string     `gorm:"column:metadata;not null;default:'{}'"`
	StartedAt        *time.Time `gorm:"column:started_at;index:idx_spans_started_at,sort:desc"`
	EndedAt          *time.Time `gorm:"column:ended_at"`
}

func (SpanEntity) TableName() string { return "spans" }

type EvaluationEntity struct {
	ID           string    `gorm:"column:id;primaryKey"`
	SpanID       string    `gorm:"column:span_id;not null"`
	ScoreType    string    `gorm:"column:score_type;not null;default:''"`
	ScoreValue   float64   `gorm:"column:score_value;not null;default:0"`
	FeedbackText string    `gorm:"column:feedback_text;not null;default:''"`
	Evaluator    string    `gorm:"column:evaluator;not null;default:''"`
	CreatedAt    time.Time `gorm:"column:created_at;type:datetime;default:CURRENT_TIMESTAMP"`
}

func (EvaluationEntity) TableName() string { return "evaluations" }

type ModelPriceEntity struct {
	ID                   string    `gorm:"column:id;primaryKey"`
	Provider             string    `gorm:"column:provider;not null"`
	ModelID              string    `gorm:"column:model_id;not null"`
	PromptPricePer1k     float64   `gorm:"column:prompt_price_per_1k;not null;default:0"`
	CompletionPricePer1k float64   `gorm:"column:completion_price_per_1k;not null;default:0"`
	UpdatedAt            time.Time `gorm:"column:updated_at;type:datetime;default:CURRENT_TIMESTAMP"`
}

func (ModelPriceEntity) TableName() string { return "model_prices" }

type LLMSessionEntity struct {
	ID        string    `gorm:"column:id;primaryKey"`
	Name      string    `gorm:"column:name;not null;default:''"`
	Metadata  string    `gorm:"column:metadata;not null;default:'{}'"`
	CreatedAt time.Time `gorm:"column:created_at;type:datetime;default:CURRENT_TIMESTAMP"`
}

func (LLMSessionEntity) TableName() string { return "llm_sessions" }

func entityModels() []any {
	return []any{
		&UserEntity{}, &SessionEntity{}, &ModelEntity{}, &TraceEntity{}, &SpanEntity{},
		&EvaluationEntity{}, &ModelPriceEntity{}, &LLMSessionEntity{},
	}
}
