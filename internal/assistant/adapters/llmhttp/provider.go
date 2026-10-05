// Package llmhttp provides an OpenAI-compatible HTTP LLM provider with tool calling support.
package llmhttp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"ecommerce-ai-assistant/internal/assistant/domain"
	"ecommerce-ai-assistant/internal/platform/httpx"
)

// Config holds the configuration for the HTTP LLM provider.
type Config struct {
	BaseURL    string
	APIKey     string
	Model      string
	TimeoutSec int
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig(baseURL, apiKey, model string) Config {
	return Config{
		BaseURL:    baseURL,
		APIKey:     apiKey,
		Model:      model,
		TimeoutSec: 60,
	}
}

// Client is an OpenAI-compatible LLM provider that communicates over HTTP.
type Client struct {
	config     Config
	httpClient *httpx.Client
}

// NewClient creates a new LLM HTTP client.
func NewClient(cfg Config) *Client {
	timeout := time.Duration(cfg.TimeoutSec) * time.Second
	if timeout == 0 {
		timeout = 60 * time.Second
	}
	return &Client{
		config:     cfg,
		httpClient: httpx.NewClient(cfg.BaseURL, cfg.APIKey, timeout),
	}
}

// Complete sends a chat completion request with tool definitions and returns the response.
func (c *Client) Complete(ctx context.Context, messages []domain.Message, tools []domain.ToolDefinition) (*domain.AssistantResponse, error) {
	openaiMessages := buildOpenAIMessages(messages)
	openaiTools := buildOpenAITools(tools)

	body := map[string]any{
		"model":       c.config.Model,
		"messages":    openaiMessages,
		"tools":       openaiTools,
		"temperature": 0.7,
	}
	if len(openaiTools) == 0 {
		body["tools"] = nil
	}

	respBody, err := c.doRequest(ctx, "/chat/completions", body)
	if err != nil {
		return nil, fmt.Errorf("LLM request failed: %w", err)
	}

	return parseResponse(respBody, c.config.Model)
}

func (c *Client) doRequest(ctx context.Context, path string, body any) (map[string]any, error) {
	b, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.httpClient.BuildURL(path), strings.NewReader(string(b)))
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}

	resp, err := c.httpClient.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("http error: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("LLM API status %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	var result map[string]any
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	return result, nil
}

func buildOpenAIMessages(messages []domain.Message) []map[string]any {
	out := make([]map[string]any, 0, len(messages))
	for _, m := range messages {
		msg := map[string]any{
			"role":    string(m.Role),
			"content": m.Content,
		}
		out = append(out, msg)
	}
	return out
}

func buildOpenAITools(tools []domain.ToolDefinition) []map[string]any {
	out := make([]map[string]any, 0, len(tools))
	for _, t := range tools {
		out = append(out, map[string]any{
			"type": "function",
			"function": map[string]any{
				"name":        t.Name,
				"description": t.Description,
				"parameters":  t.JSONSchema(),
			},
		})
	}
	return out
}

func parseResponse(body map[string]any, model string) (*domain.AssistantResponse, error) {
	choices, ok := body["choices"].([]any)
	if !ok || len(choices) == 0 {
		return nil, fmt.Errorf("unexpected response format: no choices")
	}

	choice, ok := choices[0].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("unexpected choice format")
	}

	msg, ok := choice["message"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("unexpected message format")
	}

	response := &domain.AssistantResponse{
		Model:      model,
		StopReason: "end_turn",
	}

	if stopReason, ok := choice["finish_reason"].(string); ok {
		response.StopReason = mapFinishReason(stopReason)
	}

	if content, ok := msg["content"].(string); ok && content != "" {
		response.Messages = append(response.Messages, domain.Message{
			Role:    domain.RoleAssistant,
			Content: content,
		})
	}

	if toolCalls, ok := msg["tool_calls"].([]any); ok {
		for _, tc := range toolCalls {
			tcMap, ok := tc.(map[string]any)
			if !ok {
				continue
			}
			fn, ok := tcMap["function"].(map[string]any)
			if !ok {
				continue
			}
			name, _ := fn["name"].(string)
			argsStr, _ := fn["arguments"].(string)
			var args map[string]any
			if argsStr != "" {
				_ = json.Unmarshal([]byte(argsStr), &args)
			}
			response.ToolCalls = append(response.ToolCalls, domain.ToolCall{
				ID:        getString(tcMap, "id"),
				Name:      name,
				Arguments: args,
			})
		}
	}

	if usage, ok := body["usage"].(map[string]any); ok {
		response.Usage = domain.Usage{
			InputTokens:  getInt(usage, "prompt_tokens"),
			OutputTokens: getInt(usage, "completion_tokens"),
		}
	}

	return response, nil
}

func mapFinishReason(openaiReason string) string {
	switch strings.ToLower(openaiReason) {
	case "stop", "end_turn":
		return "end_turn"
	case "length":
		return "length"
	case "tool_calls", "function_call":
		return "tool_calls"
	case "content_filter":
		return "content_filter"
	default:
		return openaiReason
	}
}

func getString(m map[string]any, key string) string {
	v, ok := m[key]
	if !ok {
		return ""
	}
	s, ok := v.(string)
	if !ok {
		return ""
	}
	return s
}

func getInt(m map[string]any, key string) int {
	v, ok := m[key]
	if !ok {
		return 0
	}
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	case int64:
		return int(n)
	default:
		return 0
	}
}
