package whatsapptransport

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"
)

const (
	twitterMaxWebhookBodySize = 65536
)

type VerificationHandler struct {
	verifyToken string
}

func NewVerificationHandler(verifyToken string) *VerificationHandler {
	return &VerificationHandler{verifyToken: verifyToken}
}

func (h *VerificationHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	mode := r.URL.Query().Get("hub.mode")
	token := r.URL.Query().Get("hub.verify_token")
	challenge := r.URL.Query().Get("hub.challenge")

	switch r.Method {
	case http.MethodGet:
		if mode == "subscribe" && token == h.verifyToken && challenge != "" {
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(challenge))
			return
		}
		w.WriteHeader(http.StatusForbidden)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

type InboundHandler struct {
	processFn func(ctx context.Context, msg InboundMessage) error
	logger    func(ctx context.Context, msg string, args ...interface{})
}

type InboundMessage struct {
	MessageID   string
	From        string
	To          string
	Timestamp   time.Time
	Text        string
	MessageType string
	RawPayload  []byte
}

func NewInboundHandler(processFn func(ctx context.Context, msg InboundMessage) error, logger func(ctx context.Context, msg string, args ...interface{})) *InboundHandler {
	return &InboundHandler{processFn: processFn, logger: logger}
}

func (h *InboundHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	switch r.Method {
	case http.MethodGet:
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("WhatsApp webhook endpoint"))
		return

	case http.MethodPost:
		r.Body = http.MaxBytesReader(w, r.Body, twitterMaxWebhookBodySize)
		body, err := io.ReadAll(r.Body)
		if err != nil {
			h.log(ctx, "failed to read body", "error", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		var envelope WebhookEnvelope
		if err := json.Unmarshal(body, &envelope); err != nil {
			h.log(ctx, "failed to parse webhook payload", "error", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if h.processFn != nil {
			go func(entries []WebhookEntry) {
				msg := parseInboundMessage(body, entries)
				if err := h.processFn(ctx, msg); err != nil {
					h.log(ctx, "process inbound message failed", "error", err, "msg_id", msg.MessageID)
				}
			}(envelope.Entry)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"accepted"}`))
		return

	default:
		w.Header().Set("Allow", "GET, POST")
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func parseInboundMessage(payload []byte, entries []WebhookEntry) InboundMessage {
	for _, entry := range entries {
		for _, change := range entry.Changes {
			value := change.Value
			for _, msg := range value.Messages {
				if msg.From == "" {
					continue
				}
				return InboundMessage{
					MessageID:   msg.ID,
					From:        msg.From,
					To:          value.Metadata.PhoneNumberID,
					MessageType: msg.Type,
					Text:        msg.Text.Body,
					RawPayload:  payload,
				}
			}
		}
	}
	return InboundMessage{RawPayload: payload}
}

func (h *InboundHandler) log(ctx context.Context, msg string, args ...interface{}) {
	if h.logger != nil {
		h.logger(ctx, msg, args...)
	}
}

type WebhookEnvelope struct {
	Object string         `json:"object"`
	Entry  []WebhookEntry `json:"entry"`
}

type WebhookEntry struct {
	ID      string          `json:"id"`
	Changes []WebhookChange `json:"changes"`
}

type WebhookChange struct {
	Field string       `json:"field"`
	Value WebhookValue `json:"value"`
}

type WebhookValue struct {
	MessagingProduct string           `json:"messaging_product"`
	Metadata         WebhookMetadata  `json:"metadata"`
	Messages         []WebhookMessage `json:"messages"`
}

type WebhookMetadata struct {
	DisplayPhoneNumberID string `json:"display_phone_number_id"`
	PhoneNumberID        string `json:"phone_number_id"`
}

type WebhookMessage struct {
	ID        string `json:"id"`
	From      string `json:"from"`
	Timestamp string `json:"timestamp"`
	Type      string `json:"type"`
	Text      struct {
		Body string `json:"body"`
	} `json:"text"`
}
