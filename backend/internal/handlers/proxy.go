package handlers

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"go.uber.org/zap"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"

	"quicklens/backend/internal/db"
	"quicklens/backend/internal/models"
)

func getOllamaHost() string {
	host := os.Getenv("OLLAMA_HOST")
	if host == "" {
		host = "host.docker.internal:11434"
	}
	return host
}

func ProxyChatCompletions(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "Failed to read request body")
		return
	}
	defer r.Body.Close()

	var reqBody map[string]interface{}
	if err := json.Unmarshal(body, &reqBody); err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}

	modelName, _ := reqBody["model"].(string)
	stream, _ := reqBody["stream"].(bool)

	// Determine upstream URL
	upstreamURL := getUpstreamURL(modelName, "/v1/chat/completions")

	startTime := time.Now()

	proxyReq, err := http.NewRequest("POST", upstreamURL, bytes.NewReader(body))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to create proxy request")
		return
	}
	proxyReq.Header.Set("Content-Type", "application/json")

	// Forward API key if present
	if apiKey := r.Header.Get("Authorization"); apiKey != "" {
		proxyReq.Header.Set("Authorization", apiKey)
	} else if envKey := os.Getenv("OPENAI_API_KEY"); envKey != "" {
		proxyReq.Header.Set("Authorization", "Bearer "+envKey)
	}

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(proxyReq)
	if err != nil {
		WriteError(w, http.StatusBadGateway, "Failed to reach upstream: "+err.Error())
		return
	}
	defer resp.Body.Close()

	durationMs := time.Since(startTime).Milliseconds()

	if stream {
		// SSE passthrough
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.WriteHeader(resp.StatusCode)

		flusher, ok := w.(http.Flusher)
		if !ok {
			WriteError(w, http.StatusInternalServerError, "Streaming not supported")
			return
		}

		var fullOutput bytes.Buffer
		var promptTokens, completionTokens int64

		scanner := bufio.NewScanner(resp.Body)
		scanner.Buffer(make([]byte, 1024*1024), 1024*1024)
		for scanner.Scan() {
			line := scanner.Text()
			fmt.Fprintf(w, "%s\n", line)
			flusher.Flush()

			// Try to capture usage from final chunk
			if len(line) > 6 && line[:6] == "data: " {
				data := line[6:]
				if data == "[DONE]" {
					continue
				}
				var chunk map[string]interface{}
				if json.Unmarshal([]byte(data), &chunk) == nil {
					if usage, ok := chunk["usage"].(map[string]interface{}); ok {
						if pt, ok := usage["prompt_tokens"].(float64); ok {
							promptTokens = int64(pt)
						}
						if ct, ok := usage["completion_tokens"].(float64); ok {
							completionTokens = int64(ct)
						}
					}
					if choices, ok := chunk["choices"].([]interface{}); ok && len(choices) > 0 {
						if choice, ok := choices[0].(map[string]interface{}); ok {
							if delta, ok := choice["delta"].(map[string]interface{}); ok {
								if content, ok := delta["content"].(string); ok {
									fullOutput.WriteString(content)
								}
							}
						}
					}
				}
			}
		}

		// Log the trace
		go logProxyTrace(modelName, string(body), fullOutput.String(), promptTokens, completionTokens, durationMs, "ok", "")
	} else {
		// Non-streaming response
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			WriteError(w, http.StatusInternalServerError, "Failed to read upstream response")
			return
		}

		// Copy response headers
		for k, v := range resp.Header {
			for _, vv := range v {
				w.Header().Set(k, vv)
			}
		}
		w.WriteHeader(resp.StatusCode)
		w.Write(respBody)

		// Parse response for logging
		var respData map[string]interface{}
		var output string
		var promptTokens, completionTokens int64
		status := "ok"
		errMsg := ""

		if resp.StatusCode >= 400 {
			status = "error"
			errMsg = string(respBody)
		}

		if json.Unmarshal(respBody, &respData) == nil {
			if usage, ok := respData["usage"].(map[string]interface{}); ok {
				if pt, ok := usage["prompt_tokens"].(float64); ok {
					promptTokens = int64(pt)
				}
				if ct, ok := usage["completion_tokens"].(float64); ok {
					completionTokens = int64(ct)
				}
			}
			if choices, ok := respData["choices"].([]interface{}); ok && len(choices) > 0 {
				if choice, ok := choices[0].(map[string]interface{}); ok {
					if msg, ok := choice["message"].(map[string]interface{}); ok {
						if content, ok := msg["content"].(string); ok {
							output = content
						}
					}
				}
			}
		}

		go logProxyTrace(modelName, string(body), output, promptTokens, completionTokens, durationMs, status, errMsg)
	}
}

func ProxyCompletions(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "Failed to read request body")
		return
	}
	defer r.Body.Close()

	var reqBody map[string]interface{}
	json.Unmarshal(body, &reqBody)
	modelName, _ := reqBody["model"].(string)

	upstreamURL := getUpstreamURL(modelName, "/v1/completions")
	startTime := time.Now()

	proxyReq, err := http.NewRequest("POST", upstreamURL, bytes.NewReader(body))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to create proxy request")
		return
	}
	proxyReq.Header.Set("Content-Type", "application/json")
	if apiKey := r.Header.Get("Authorization"); apiKey != "" {
		proxyReq.Header.Set("Authorization", apiKey)
	} else if envKey := os.Getenv("OPENAI_API_KEY"); envKey != "" {
		proxyReq.Header.Set("Authorization", "Bearer "+envKey)
	}

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(proxyReq)
	if err != nil {
		WriteError(w, http.StatusBadGateway, "Failed to reach upstream: "+err.Error())
		return
	}
	defer resp.Body.Close()

	durationMs := time.Since(startTime).Milliseconds()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to read upstream response")
		return
	}

	for k, v := range resp.Header {
		for _, vv := range v {
			w.Header().Set(k, vv)
		}
	}
	w.WriteHeader(resp.StatusCode)
	w.Write(respBody)

	status := "ok"
	errMsg := ""
	if resp.StatusCode >= 400 {
		status = "error"
		errMsg = string(respBody)
	}

	go logProxyTrace(modelName, string(body), string(respBody), 0, 0, durationMs, status, errMsg)
}

func ProxyModels(w http.ResponseWriter, r *http.Request) {
	rows, err := db.DB.Query("SELECT id, name, provider, model_id, created_at FROM models WHERE status = 'online'")
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to query models")
		return
	}
	defer rows.Close()

	type openAIModel struct {
		ID      string `json:"id"`
		Object  string `json:"object"`
		Created int64  `json:"created"`
		OwnedBy string `json:"owned_by"`
	}

	modelList := make([]openAIModel, 0)
	for rows.Next() {
		var id, name, provider, modelID string
		var createdAt time.Time
		if rows.Scan(&id, &name, &provider, &modelID, &createdAt) == nil {
			mid := modelID
			if mid == "" {
				mid = name
			}
			modelList = append(modelList, openAIModel{
				ID:      mid,
				Object:  "model",
				Created: createdAt.Unix(),
				OwnedBy: provider,
			})
		}
	}

	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"object": "list",
		"data":   modelList,
	})
}

func ProxyOllamaGenerate(w http.ResponseWriter, r *http.Request) {
	proxyToOllama(w, r, "/api/generate", "POST")
}

func ProxyOllamaChat(w http.ResponseWriter, r *http.Request) {
	proxyToOllama(w, r, "/api/chat", "POST")
}

func ProxyOllamaTags(w http.ResponseWriter, r *http.Request) {
	proxyToOllama(w, r, "/api/tags", "GET")
}

func proxyToOllama(w http.ResponseWriter, r *http.Request, path string, method string) {
	ollamaHost := getOllamaHost()
	upstreamURL := fmt.Sprintf("http://%s%s", ollamaHost, path)

	var bodyReader io.Reader
	if method == "POST" {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			WriteError(w, http.StatusBadRequest, "Failed to read request body")
			return
		}
		defer r.Body.Close()
		bodyReader = bytes.NewReader(body)
	}

	startTime := time.Now()

	proxyReq, err := http.NewRequest(method, upstreamURL, bodyReader)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to create proxy request")
		return
	}
	if method == "POST" {
		proxyReq.Header.Set("Content-Type", "application/json")
	}

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(proxyReq)
	if err != nil {
		WriteError(w, http.StatusBadGateway, "Failed to reach Ollama: "+err.Error())
		return
	}
	defer resp.Body.Close()

	durationMs := time.Since(startTime).Milliseconds()
	_ = durationMs

	// Stream response back
	for k, v := range resp.Header {
		for _, vv := range v {
			w.Header().Set(k, vv)
		}
	}
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func getUpstreamURL(modelName string, path string) string {
	// Check if model has a configured endpoint
	var endpoint string
	err := db.DB.QueryRow(
		"SELECT endpoint FROM models WHERE (model_id = ? OR name = ?) AND endpoint != ''",
		modelName, modelName,
	).Scan(&endpoint)
	if err == nil && endpoint != "" {
		return endpoint + path
	}

	// Check if it's an Ollama model
	var provider string
	db.DB.QueryRow(
		"SELECT provider FROM models WHERE model_id = ? OR name = ?",
		modelName, modelName,
	).Scan(&provider)
	if provider == "ollama" {
		ollamaHost := getOllamaHost()
		return fmt.Sprintf("http://%s%s", ollamaHost, path)
	}

	// Default to OpenAI
	baseURL := os.Getenv("OPENAI_BASE_URL")
	if baseURL == "" {
		baseURL = "https://api.openai.com"
	}
	return baseURL + path
}

func logProxyTrace(modelName, input, output string, promptTokens, completionTokens, durationMs int64, status, errMsg string) {
	provider := detectProvider(modelName)

	ingestReq := models.TraceIngestRequest{
		TraceID: uuid.New().String(),
		Name:    fmt.Sprintf("Proxy: %s", modelName),
		Spans: []models.SpanIngestRequest{
			{
				SpanID:           uuid.New().String(),
				Name:             fmt.Sprintf("chat.completions.%s", modelName),
				Type:             "llm",
				ModelID:          modelName,
				Provider:         provider,
				Input:            truncate(input, 5000),
				Output:           truncate(output, 5000),
				PromptTokens:     promptTokens,
				CompletionTokens: completionTokens,
				DurationMs:       durationMs,
				Status:           status,
				ErrorMessage:     errMsg,
			},
		},
	}

	select {
	case ingestChan <- ingestReq:
	default:
		zap.L().Sugar().Info("Ingest queue full, dropping proxy trace")
	}
}

func detectProvider(modelName string) string {
	var provider string
	err := db.DB.QueryRow(
		"SELECT provider FROM models WHERE model_id = ? OR name = ?",
		modelName, modelName,
	).Scan(&provider)
	if err == nil {
		return provider
	}

	// Heuristic detection
	switch {
	case len(modelName) > 3 && modelName[:3] == "gpt":
		return "openai"
	case len(modelName) > 6 && modelName[:6] == "claude":
		return "anthropic"
	case len(modelName) > 5 && modelName[:5] == "llama":
		return "ollama"
	default:
		return "unknown"
	}
}
