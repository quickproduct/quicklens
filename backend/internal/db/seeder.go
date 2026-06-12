package db

import (
	"encoding/json"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

func SeedMockDataIfEmpty(logger *zap.Logger) {
	var count int64
	if err := Gorm.Model(&TraceEntity{}).Count(&count).Error; err != nil {
		logger.Error("seeder: failed to count traces", zap.Error(err))
		return
	}
	if count > 0 {
		logger.Info("seeder: database already contains data, skipping mock seeding")
		return
	}

	logger.Info("seeder: seeding mock data for LLM monitoring MVP...")

	now := time.Now().UTC()

	// 1. Seed Models
	mockModels := []ModelEntity{
		{
			ID:            uuid.New().String(),
			Name:          "GPT-4o",
			Provider:      "openai",
			ModelID:       "gpt-4o",
			Endpoint:      "https://api.openai.com",
			Status:        "online",
			ContextLength: 128000,
			CreatedAt:     now.Add(-48 * time.Hour),
			UpdatedAt:     now.Add(-48 * time.Hour),
		},
		{
			ID:            uuid.New().String(),
			Name:          "Claude 3.5 Sonnet",
			Provider:      "anthropic",
			ModelID:       "claude-3-5-sonnet",
			Endpoint:      "https://api.anthropic.com",
			Status:        "online",
			ContextLength: 200000,
			CreatedAt:     now.Add(-48 * time.Hour),
			UpdatedAt:     now.Add(-48 * time.Hour),
		},
		{
			ID:            uuid.New().String(),
			Name:          "Gemini 1.5 Pro",
			Provider:      "google",
			ModelID:       "gemini-1.5-pro",
			Endpoint:      "https://generativelanguage.googleapis.com",
			Status:        "online",
			ContextLength: 1048576,
			CreatedAt:     now.Add(-48 * time.Hour),
			UpdatedAt:     now.Add(-48 * time.Hour),
		},
		{
			ID:            uuid.New().String(),
			Name:          "Llama 3 (8B)",
			Provider:      "ollama",
			ModelID:       "llama3:8b",
			Endpoint:      "http://localhost:11434",
			Status:        "online",
			Quantization:  "Q4_K_M",
			SizeBytes:     4700000000,
			ContextLength: 8192,
			CreatedAt:     now.Add(-48 * time.Hour),
			UpdatedAt:     now.Add(-48 * time.Hour),
		},
	}

	for _, m := range mockModels {
		if err := Gorm.Create(&m).Error; err != nil {
			logger.Error("seeder: failed to seed model", zap.String("name", m.Name), zap.Error(err))
		}
	}

	// 2. Seed Pricing
	prices := []ModelPriceEntity{
		{ID: uuid.New().String(), Provider: "openai", ModelID: "gpt-4o", PromptPricePer1k: 0.005, CompletionPricePer1k: 0.015, UpdatedAt: now},
		{ID: uuid.New().String(), Provider: "anthropic", ModelID: "claude-3-5-sonnet", PromptPricePer1k: 0.003, CompletionPricePer1k: 0.015, UpdatedAt: now},
		{ID: uuid.New().String(), Provider: "google", ModelID: "gemini-1.5-pro", PromptPricePer1k: 0.007, CompletionPricePer1k: 0.021, UpdatedAt: now},
		{ID: uuid.New().String(), Provider: "ollama", ModelID: "llama3:8b", PromptPricePer1k: 0.0, CompletionPricePer1k: 0.0, UpdatedAt: now},
	}

	for _, p := range prices {
		if err := Gorm.Create(&p).Error; err != nil {
			logger.Error("seeder: failed to seed price", zap.String("model", p.ModelID), zap.Error(err))
		}
	}

	// Price map for seeder calculations
	priceMap := map[string]struct{ PromptPrice, CompletionPrice float64 }{
		"gpt-4o":            {0.005, 0.015},
		"claude-3-5-sonnet": {0.003, 0.015},
		"gemini-1.5-pro":    {0.007, 0.021},
		"llama3:8b":         {0.0, 0.0},
	}

	// 3. Seed Traces and Spans
	traceNames := []string{
		"Chat Assistant",
		"Document Summarizer",
		"SQL Code Generator",
		"Semantic Search Q&A",
		"Agent Plan-Execute",
		"Translation Engine",
		"Sentiment Analyzer",
	}

	promptTemplates := []string{
		"You are a helpful coding assistant. Write a function to reverse a linked list.",
		"Summarize the following contract and list the top 3 liabilities: %s",
		"Convert this natural language prompt into PostgreSQL: 'Show me users registered in last 30 days'",
		"Read the context below and answer the question: 'What is the refund policy?'\nContext: %s",
		"Create an agent plan to schedule a meeting with Alice at 3pm tomorrow.",
		"Translate this email into Spanish:\n%s",
		"Analyze the sentiment of this review: 'The service was sluggish but the food was absolutely spectacular.'",
	}

	responses := []string{
		"Here is the code to reverse a singly linked list in Python:\n```python\ndef reverse_list(head):\n    prev = None\n    curr = head\n    while curr:\n        next_node = curr.next\n        curr.next = prev\n        prev = curr\n        curr = next_node\n    return prev\n```",
		"SUMMARY OF CONTRACT LIABILITIES:\n1. Late delivery penalty: $5,000/day.\n2. Unlimited liability for breach of intellectual property rights.\n3. Early termination requires 90 days notice or payment in lieu.",
		"```sql\nSELECT * FROM users WHERE registered_at >= NOW() - INTERVAL '30 days';\n```",
		"Based on the provided context, the refund policy allows returns within 30 days of purchase with a valid receipt. Shipping costs are non-refundable.",
		"PLAN:\n1. Search calendar for Alice's email address.\n2. Fetch availability for Alice and current user.\n3. Identify overlap at 3:00 PM tomorrow.\n4. Send calendar invite with agenda.",
		"Aquí está la traducción del correo electrónico:\nEstimado equipo, me gustaría programar una reunión de seguimiento para el próximo lunes a las 10:00 AM.",
		"The sentiment is MIXED-POSITIVE. The reviewer expresses frustration with service speed ('sluggish') but high satisfaction with food quality ('absolutely spectacular').",
	}

	metaSamples := []map[string]any{
		{"temperature": 0.2, "top_p": 0.9, "max_tokens": 1024},
		{"temperature": 0.7, "top_p": 0.95, "max_tokens": 2048},
		{"temperature": 0.0, "top_p": 1.0, "max_tokens": 512},
		{"temperature": 0.3, "top_p": 0.9, "max_tokens": 1000},
		{"temperature": 0.5, "top_p": 0.95, "max_tokens": 4096},
		{"temperature": 0.3, "top_p": 0.85, "max_tokens": 2048},
		{"temperature": 0.1, "top_p": 0.9, "max_tokens": 256},
	}

	r := rand.New(rand.NewSource(42))

	// Seed 60 traces across the past 24 hours
	for i := 0; i < 60; i++ {
		// Random time over the last 24h
		minsAgo := r.Intn(1440)
		traceTime := now.Add(-time.Duration(minsAgo) * time.Minute)

		idx := r.Intn(len(traceNames))
		traceName := traceNames[idx]
		promptText := promptTemplates[idx]
		responseText := responses[idx]
		metadataVal := metaSamples[idx]

		// Choose random model
		modelChoice := mockModels[r.Intn(len(mockModels))]

		// Trace details
		traceID := uuid.New().String()
		dbTraceID := uuid.New().String()
		sessionID := uuid.New().String()

		status := "ok"
		errMsg := ""
		// 8% error rate
		if r.Float32() < 0.08 {
			status = "error"
			errMsg = "API Rate limit exceeded (code: 429)"
			responseText = ""
		}

		durationMs := int64(120 + r.Intn(1800))
		if modelChoice.Provider == "ollama" {
			durationMs = int64(600 + r.Intn(2500)) // Ollama is slightly slower locally
		}

		promptTokens := int64(50 + r.Intn(300))
		completionTokens := int64(30 + r.Intn(400))
		if status == "error" {
			completionTokens = 0
		}

		// Calculate cost
		prices := priceMap[modelChoice.ModelID]
		cost := (float64(promptTokens)/1000.0)*prices.PromptPrice + (float64(completionTokens)/1000.0)*prices.CompletionPrice

		metaJSON, _ := json.Marshal(metadataVal)

		// Create parent Trace
		tEntity := TraceEntity{
			ID:               dbTraceID,
			TraceID:          traceID,
			SessionID:        sessionID,
			Name:             traceName,
			Status:           status,
			TotalDurationMs:  durationMs,
			TotalTokens:      promptTokens + completionTokens,
			PromptTokens:     promptTokens,
			CompletionTokens: completionTokens,
			TotalCost:        cost,
			InputPreview:     promptText[:20],
			OutputPreview:    responseText,
			Metadata:         "{}",
			CreatedAt:        traceTime,
		}
		if len(tEntity.OutputPreview) > 200 {
			tEntity.OutputPreview = tEntity.OutputPreview[:200]
		}
		if err := Gorm.Create(&tEntity).Error; err != nil {
			logger.Error("seeder: failed to save trace", zap.Error(err))
			continue
		}

		// Create primary LLM span
		spanTimeEnd := traceTime.Add(time.Duration(durationMs) * time.Millisecond)
		spanEntity := SpanEntity{
			ID:               uuid.New().String(),
			TraceID:          dbTraceID,
			ParentSpanID:     "",
			Name:             traceName + " Completion Call",
			Type:             "llm",
			ModelID:          modelChoice.ModelID,
			Provider:         modelChoice.Provider,
			Input:            promptText,
			Output:           responseText,
			PromptTokens:     promptTokens,
			CompletionTokens: completionTokens,
			TotalTokens:      promptTokens + completionTokens,
			Cost:             cost,
			DurationMs:       durationMs,
			Status:           status,
			ErrorMessage:     errMsg,
			Metadata:         string(metaJSON),
			StartedAt:        &traceTime,
			EndedAt:          &spanTimeEnd,
		}
		if err := Gorm.Create(&spanEntity).Error; err != nil {
			logger.Error("seeder: failed to save span", zap.Error(err))
		}

		// Add custom secondary span for certain runs to simulate chain-of-thought/RAG agents
		if idx == 3 && status == "ok" { // Semantic Search Q&A (RAG)
			Gorm.Create(&SpanEntity{
				ID:           uuid.New().String(),
				TraceID:      dbTraceID,
				ParentSpanID: spanEntity.ID,
				Name:         "Vector DB Search",
				Type:         "retrieval",
				Input:        "What is the refund policy?",
				Output:       "Found 1 document matches in vector space: refund_policy_doc_2026.txt",
				DurationMs:   int64(45 + r.Intn(90)),
				Status:       "ok",
				Metadata:     "{}",
				StartedAt:    &traceTime,
				EndedAt:      &spanTimeEnd,
			})
		} else if idx == 4 && status == "ok" { // Agent Plan-Execute
			Gorm.Create(&SpanEntity{
				ID:           uuid.New().String(),
				TraceID:      dbTraceID,
				ParentSpanID: spanEntity.ID,
				Name:         "Search Calendar Tool",
				Type:         "tool",
				Input:        "Search Alice availability",
				Output:       "Alice calendar fetched. Available slot detected at 3:00 PM tomorrow.",
				DurationMs:   int64(100 + r.Intn(200)),
				Status:       "ok",
				Metadata:     "{}",
				StartedAt:    &traceTime,
				EndedAt:      &spanTimeEnd,
			})
		}
	}

	logger.Info("seeder: mock database seeding completed successfully!")
}
