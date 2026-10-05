package devchathandler

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"ecommerce-ai-assistant/internal/assistant/domain"
)

type AssistantUseCase interface {
	Run(ctx context.Context, conversationID, userID, message string) (*domain.AssistantResponse, error)
}

type Config struct {
	AllowedOrigins []string
	RateLimitRPS   int
}

func DefaultConfig() Config {
	return Config{
		AllowedOrigins: []string{"*"},
		RateLimitRPS:   30,
	}
}

type Handler struct {
	assistant AssistantUseCase
	config    Config
}

func NewHandler(assistant AssistantUseCase, cfg Config) *Handler {
	return &Handler{assistant: assistant, config: cfg}
}

type chatRequest struct {
	ConversationID string `json:"conversation_id"`
	UserID         string `json:"user_id"`
	Message        string `json:"message"`
}

type chatResponse struct {
	ConversationID string            `json:"conversation_id"`
	Messages       []domain.Message  `json:"messages"`
	ToolCalls      []domain.ToolCall `json:"tool_calls,omitempty"`
	StopReason     string            `json:"stop_reason"`
	Usage          usageResponse     `json:"usage"`
	ProcessedAt    time.Time         `json:"processed_at"`
}

type usageResponse struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if !h.isOriginAllowed(r) {
		w.Header().Set("Allow", "POST, OPTIONS")
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error":"origin not allowed"}`))
		return
	}

	switch r.Method {
	case http.MethodOptions:
		w.Header().Set("Allow", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.WriteHeader(http.StatusNoContent)
		return

	case http.MethodPost:
		var req chatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"error":"invalid JSON body"}`))
			return
		}

		req.Message = strings.TrimSpace(req.Message)
		req.ConversationID = strings.TrimSpace(req.ConversationID)
		req.UserID = strings.TrimSpace(req.UserID)

		if req.ConversationID == "" || req.UserID == "" || req.Message == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"error":"conversation_id, user_id, and message are required"}`))
			return
		}

		resp, err := h.assistant.Run(ctx, req.ConversationID, req.UserID, req.Message)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"error":"` + escapeJSON(err.Error()) + `"}`))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(chatResponse{
			ConversationID: req.ConversationID,
			Messages:       resp.Messages,
			ToolCalls:      resp.ToolCalls,
			StopReason:     resp.StopReason,
			Usage: usageResponse{
				InputTokens:  resp.Usage.InputTokens,
				OutputTokens: resp.Usage.OutputTokens,
			},
			ProcessedAt: time.Now().UTC(),
		})

	default:
		w.Header().Set("Allow", "POST, OPTIONS")
		w.WriteHeader(http.StatusMethodNotAllowed)
		_, _ = w.Write([]byte(`{"error":"method not allowed"}`))
	}
}

func (h *Handler) isOriginAllowed(r *http.Request) bool {
	if len(h.config.AllowedOrigins) == 0 {
		return true
	}
	for _, o := range h.config.AllowedOrigins {
		if o == "*" {
			return true
		}
		if o == r.Header.Get("Origin") {
			return true
		}
	}
	return false
}

func escapeJSON(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}
