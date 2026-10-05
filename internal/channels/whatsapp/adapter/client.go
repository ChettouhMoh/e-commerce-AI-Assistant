package adapter

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"ecommerce-ai-assistant/internal/channels/ports"
)

type GraphAPIConfig struct {
	AccessToken        string
	PhoneNumberID      string
	GraphAPIVersion    string
	RequestTimeoutSecs int
}

func DefaultGraphAPIConfig(accessToken, phoneNumberID string) GraphAPIConfig {
	return GraphAPIConfig{
		AccessToken:        accessToken,
		PhoneNumberID:      phoneNumberID,
		GraphAPIVersion:    "v18.0",
		RequestTimeoutSecs: 30,
	}
}

type Client struct {
	baseURL     string
	accessToken string
	httpClient  *http.Client
}

func NewClient(cfg GraphAPIConfig) *Client {
	baseURL := fmt.Sprintf("https://graph.facebook.com/%s/%s", cfg.GraphAPIVersion, cfg.PhoneNumberID)
	return &Client{
		baseURL:     baseURL,
		accessToken: cfg.AccessToken,
		httpClient:  &http.Client{Timeout: time.Duration(cfg.RequestTimeoutSecs) * time.Second},
	}
}

func (c *Client) SendTextMessage(ctx context.Context, recipient, text string) error {
	url := fmt.Sprintf("%s/messages", c.baseURL)
	payload := map[string]interface{}{
		"messaging_product": "whatsapp",
		"to":                recipient,
		"type":              "text",
		"text": map[string]interface{}{
			"preview_url": false,
			"body":        text,
		},
	}

	return c.doPostJSON(ctx, url, payload)
}

func (c *Client) SendTemplateMessage(ctx context.Context, recipient, templateName string, components []map[string]interface{}) error {
	url := fmt.Sprintf("%s/messages", c.baseURL)

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

	return c.doPostJSON(ctx, url, payload)
}

func (c *Client) SendInteractiveMessage(ctx context.Context, recipient string, interactive ports.InteractiveMessage) error {
	url := fmt.Sprintf("%s/messages", c.baseURL)

	buttons := make([]ports.Button, 0, len(interactive.Buttons))
	for _, btn := range interactive.Buttons {
		buttons = append(buttons, ports.Button{
			Type: btn.Type,
			Reply: struct {
				ID    string `json:"id"`
				Title string `json:"title"`
			}{
				ID:    btn.Reply.ID,
				Title: btn.Reply.Title,
			},
		})
	}

	payload := map[string]interface{}{
		"messaging_product": "whatsapp",
		"to":                recipient,
		"type":              "interactive",
		"interactive": map[string]interface{}{
			"type": interactive.Type,
			"header": map[string]interface{}{
				"type": interactive.Header.Type,
				"text": interactive.Header.Text,
			},
			"body": map[string]interface{}{
				"text": interactive.Body.Text,
			},
			"footer": map[string]interface{}{
				"text": interactive.Footer.Text,
			},
			"action": map[string]interface{}{
				"buttons": buttons,
			},
		},
	}

	return c.doPostJSON(ctx, url, payload)
}

func (c *Client) doPostJSON(ctx context.Context, url string, payload map[string]interface{}) error {
	b, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(string(b)))
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.accessToken)

	resp, err := c.httpClient.Do(req)
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
			return &ports.APIError{StatusCode: resp.StatusCode, Message: apiErr.Error.Message}
		}
		return &ports.APIError{StatusCode: resp.StatusCode, Message: string(respBody)}
	}

	return nil
}
