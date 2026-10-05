// Package ports defines the messaging channel interface.
package ports

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"ecommerce-ai-assistant/internal/channels/domain"
)

// MessagingChannel defines the interface for sending and receiving messages
// across different communication channels.
type MessagingChannel interface {
	// Send delivers an outbound message to the channel.
	Send(ctx context.Context, msg domain.OutboundMessage) error

	// SendBatch delivers multiple outbound messages to the channel.
	SendBatch(ctx context.Context, msgs []domain.OutboundMessage) error

	// Subscribe registers a handler for inbound messages.
	Subscribe(handler InboundMessageHandler) error

	// Start begins processing messages for this channel.
	Start(ctx context.Context) error

	// Stop gracefully shuts down the channel.
	Stop(ctx context.Context) error
}

// InboundMessageHandler is a function that processes inbound messages.
type InboundMessageHandler func(ctx context.Context, msg domain.InboundMessage) error

// HTTPClient abstracts HTTP calls so the WhatsApp service can be tested without
// depending on the concrete net/http client.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// OutboundSender abstracts the act of sending an outbound WhatsApp message.
type OutboundSender interface {
	SendText(ctx context.Context, recipient, text string) error
	SendTemplate(ctx context.Context, recipient, templateName string, components []map[string]interface{}) error
}

// OutboundMessage is a test-friendly representation of an outbound message.
type OutboundMessage struct {
	Type      string                 `json:"type"`
	Recipient string                 `json:"recipient"`
	Content   map[string]interface{} `json:"content"`
	SentAt    time.Time              `json:"sent_at,omitempty"`
}

// Button represents a WhatsApp interactive button.
type Button struct {
	Type  string      `json:"type"`
	Reply buttonReply `json:"reply"`
}

type buttonReply struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

// InteractiveMessage represents a WhatsApp interactive message.
type InteractiveMessage struct {
	Type   string `json:"type"`
	Header struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"header"`
	Body struct {
		Text string `json:"text"`
	} `json:"body"`
	Footer struct {
		Text string `json:"text"`
	} `json:"footer"`
	Buttons []Button `json:"buttons"`
}

// APIError represents an error returned by the WhatsApp Graph API.
type APIError struct {
	StatusCode int
	Message    string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API error %d: %s", e.StatusCode, e.Message)
}
