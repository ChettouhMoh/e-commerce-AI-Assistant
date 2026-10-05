package adapter

import (
	"context"
	"fmt"
	"sync"
	"time"

	"ecommerce-ai-assistant/internal/channels/domain"
	"ecommerce-ai-assistant/internal/channels/ports"
)

type OutboundSender struct {
	mu      sync.Mutex
	sent    []ports.OutboundMessage
	enabled bool
}

func NewOutboundSender(enabled bool) *OutboundSender {
	return &OutboundSender{enabled: enabled}
}

func (f *OutboundSender) SendText(ctx context.Context, recipient, text string) error {
	if !f.enabled {
		return fmt.Errorf("outbound sender disabled")
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.sent = append(f.sent, ports.OutboundMessage{
		Type:      "text",
		Recipient: recipient,
		Content:   map[string]interface{}{"text": text},
		SentAt:    time.Now(),
	})
	return nil
}

func (f *OutboundSender) SendTemplate(ctx context.Context, recipient, templateName string, components []map[string]interface{}) error {
	if !f.enabled {
		return fmt.Errorf("outbound sender disabled")
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.sent = append(f.sent, ports.OutboundMessage{
		Type:      "template",
		Recipient: recipient,
		Content:   map[string]interface{}{"template": templateName, "components": components},
		SentAt:    time.Now(),
	})
	return nil
}

func (f *OutboundSender) SentMessages() []ports.OutboundMessage {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]ports.OutboundMessage, len(f.sent))
	copy(out, f.sent)
	return out
}

func (f *OutboundSender) Reset() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.sent = nil
}

type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error on %s: %s", e.Field, e.Message)
}

func ValidateInboundMessage(msg domain.InboundMessage) error {
	if msg.UserID == "" {
		return &ValidationError{Field: "user_id", Message: "sender identifier is required"}
	}
	if msg.Content == "" {
		return &ValidationError{Field: "content", Message: "message content is required"}
	}
	if len(msg.Content) > 4096 {
		return &ValidationError{Field: "content", Message: "message exceeds 4096 character limit"}
	}
	return nil
}
