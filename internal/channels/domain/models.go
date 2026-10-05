// Package channels contains the core domain models for messaging channel events.
package domain

import (
	"time"
)

// InboundMessage represents a message received from a channel.
type InboundMessage struct {
	ID          string
	ChannelID   string
	ChannelType string
	UserID      string
	Username    string
	Content     string
	Attachments []Attachment
	Metadata    map[string]string
	ReceivedAt  time.Time
}

// Attachment represents a file attachment in a message.
type Attachment struct {
	ID        string
	Filename  string
	MimeType  string
	Size      int64
	URL       string
	Thumbnail string
}

// OutboundMessage represents a message to be sent to a channel.
type OutboundMessage struct {
	ID          string
	ChannelID   string
	ChannelType string
	UserID      string
	Content     string
	Attachments []Attachment
	Metadata    map[string]string
	Priority    int
}

// ChannelEvent represents a lifecycle event for a messaging channel.
type ChannelEvent struct {
	ID          string
	ChannelID   string
	ChannelType string
	EventType   string
	UserID      string
	Payload     map[string]interface{}
	Timestamp   time.Time
}
