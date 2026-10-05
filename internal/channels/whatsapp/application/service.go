package application

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"ecommerce-ai-assistant/internal/channels/domain"
	"ecommerce-ai-assistant/internal/channels/ports"
)

type Deduplicator struct {
	mu       sync.Mutex
	seen     map[string]time.Time
	window   time.Duration
	stopOnce sync.Once
	stopCh   chan struct{}
}

func NewDeduplicator(window time.Duration) *Deduplicator {
	if window <= 0 {
		window = 5 * time.Minute
	}
	d := &Deduplicator{
		seen:   make(map[string]time.Time),
		window: window,
		stopCh: make(chan struct{}),
	}
	go d.reapLoop()
	return d
}

func (d *Deduplicator) Seen(id string) bool {
	if id == "" {
		return false
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	if t, ok := d.seen[id]; ok && time.Since(t) < d.window {
		return true
	}
	d.seen[id] = time.Now()
	return false
}

func (d *Deduplicator) reapLoop() {
	ticker := time.NewTicker(d.window)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			d.mu.Lock()
			cutoff := time.Now().Add(-d.window)
			for k, v := range d.seen {
				if v.Before(cutoff) {
					delete(d.seen, k)
				}
			}
			d.mu.Unlock()
		case <-d.stopCh:
			return
		}
	}
}

func (d *Deduplicator) Stop() {
	d.stopOnce.Do(func() { close(d.stopCh) })
}

type WhatsAppService struct {
	verifyToken     string
	appSecret       string
	sigRequired     bool
	accessToken     string
	phoneNumberID   string
	graphAPIVersion string
	httpClient      ports.HTTPClient
	sender          ports.OutboundSender
	dedup           *Deduplicator
}

func NewWhatsAppService(verifyToken, appSecret string, sigRequired bool, accessToken, phoneNumberID, graphAPIVersion string, httpClient ports.HTTPClient, sender ports.OutboundSender) *WhatsAppService {
	if graphAPIVersion == "" {
		graphAPIVersion = "v18.0"
	}
	return &WhatsAppService{
		verifyToken:     verifyToken,
		appSecret:       appSecret,
		sigRequired:     sigRequired,
		accessToken:     accessToken,
		phoneNumberID:   phoneNumberID,
		graphAPIVersion: graphAPIVersion,
		httpClient:      httpClient,
		sender:          sender,
		dedup:           NewDeduplicator(5 * time.Minute),
	}
}

func (s *WhatsAppService) VerifyToken(token, challenge string) (string, bool) {
	if token != s.verifyToken {
		return "", false
	}
	return challenge, true
}

func (s *WhatsAppService) ValidateSignature(rawPayload []byte, signatureHeader string) error {
	if !s.sigRequired || s.appSecret == "" {
		if !s.sigRequired {
			return nil
		}
		return fmt.Errorf("app secret not configured")
	}

	if signatureHeader == "" {
		return fmt.Errorf("missing X-Hub-Signature-256")
	}

	if !strings.HasPrefix(signatureHeader, "sha256=") {
		return fmt.Errorf("invalid signature format")
	}

	sigHex := strings.TrimPrefix(signatureHeader, "sha256=")
	if len(sigHex) != 64 {
		return fmt.Errorf("invalid signature length")
	}

	mac := hmac.New(sha256.New, []byte(s.appSecret))
	mac.Write(rawPayload)
	expectedMAC := hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(sigHex), []byte(expectedMAC)) {
		return fmt.Errorf("signature mismatch")
	}
	return nil
}

func (s *WhatsAppService) IsDuplicate(msgID string) bool {
	return s.dedup.Seen(msgID)
}

func (s *WhatsAppService) SendTextMessage(ctx context.Context, recipient, text string) error {
	if s.sender != nil {
		return s.sender.SendText(ctx, recipient, text)
	}

	url := fmt.Sprintf("https://graph.facebook.com/%s/%s/messages", s.graphAPIVersion, s.phoneNumberID)
	payload := map[string]interface{}{
		"messaging_product": "whatsapp",
		"to":                recipient,
		"type":              "text",
		"text": map[string]interface{}{
			"preview_url": false,
			"body":        text,
		},
	}

	b, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal message: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(string(b)))
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.accessToken)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("http error: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var apiErr struct {
			Error struct {
				Message string `json:"message"`
			} `json:"error"`
		}
		if json.Unmarshal(respBody, &apiErr) == nil && apiErr.Error.Message != "" {
			return fmt.Errorf("whatsapp API %d: %s", resp.StatusCode, apiErr.Error.Message)
		}
		return fmt.Errorf("whatsapp API status %d", resp.StatusCode)
	}
	return nil
}

func (s *WhatsAppService) SendTemplateMessage(ctx context.Context, recipient, templateName string, components []map[string]interface{}) error {
	if s.sender != nil {
		return s.sender.SendTemplate(ctx, recipient, templateName, components)
	}

	url := fmt.Sprintf("https://graph.facebook.com/%s/%s/messages", s.graphAPIVersion, s.phoneNumberID)

	tmplParams := make([]map[string]interface{}, 0, len(components))
	for _, comp := range components {
		if comp != nil {
			tmplParams = append(tmplParams, map[string]interface{}{
				"type": "text",
				"text": fmt.Sprintf("%v", comp),
			})
		}
	}

	payload := map[string]interface{}{
		"messaging_product": "whatsapp",
		"to":                recipient,
		"type":              "template",
		"template": map[string]interface{}{
			"name": templateName,
			"language": map[string]string{
				"code": "en_US",
			},
			"components": []map[string]interface{}{
				{
					"type":       "body",
					"parameters": tmplParams,
				},
			},
		},
	}

	b, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal template message: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(string(b)))
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.accessToken)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("http error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("whatsapp template API status %d", resp.StatusCode)
	}
	return nil
}

func (s *WhatsAppService) Normalize(event domain.ChannelEvent) (*domain.InboundMessage, error) {
	content, _ := event.Payload["content"].(string)
	userID, _ := event.Payload["user_id"].(string)
	msgID, _ := event.Payload["message_id"].(string)

	if content == "" {
		return nil, fmt.Errorf("no message content in event payload")
	}

	now := time.Now()

	return &domain.InboundMessage{
		ID:          msgID,
		ChannelID:   event.ChannelID,
		ChannelType: event.ChannelType,
		UserID:      userID,
		Content:     strings.TrimSpace(content),
		Metadata:    stringMapFromPayload(event.Payload),
		ReceivedAt:  now,
	}, nil
}

func stringMapFromPayload(payload map[string]interface{}) map[string]string {
	if payload == nil {
		return nil
	}
	result := make(map[string]string, len(payload))
	for k, v := range payload {
		if s, ok := v.(string); ok {
			result[k] = s
		}
	}
	return result
}

func (s *WhatsAppService) Stop() {
	s.dedup.Stop()
}
