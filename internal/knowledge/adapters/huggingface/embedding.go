package huggingface

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"ecommerce-ai-assistant/internal/knowledge/ports"
)

// Config holds the configuration for the Hugging Face embedding adapter.
type Config struct {
	BaseURL    string
	Model      string
	APIKey     string
	TimeoutSec int
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig(model, apiKey string) Config {
	return Config{
		BaseURL:    "https://api-inference.huggingface.co",
		Model:      model,
		APIKey:     apiKey,
		TimeoutSec: 60,
	}
}

// Client is a Hugging Face embedding adapter that calls the Inference API.
type Client struct {
	baseURL    string
	model      string
	apiKey     string
	httpClient *http.Client
}

// NewClient creates a new Hugging Face embedding client.
func NewClient(cfg Config) *Client {
	timeout := time.Duration(cfg.TimeoutSec) * time.Second
	if timeout == 0 {
		timeout = 60 * time.Second
	}
	return &Client{
		baseURL:    cfg.BaseURL,
		model:      cfg.Model,
		apiKey:     cfg.APIKey,
		httpClient: &http.Client{Timeout: timeout},
	}
}

// Embed sends texts to the Hugging Face Inference API and returns embeddings.
// It uses the feature-extraction pipeline endpoint.
func (c *Client) Embed(ctx context.Context, texts []string) ([][]float64, error) {
	if len(texts) == 0 {
		return nil, nil
	}

	results := make([][]float64, len(texts))
	var firstErr error

	for i, text := range texts {
		emb, err := c.embedOne(ctx, text)
		if err != nil {
			if firstErr == nil {
				firstErr = fmt.Errorf("embedding failed for text %d: %w", i, err)
			}
			results[i] = make([]float64, c.dimension())
			continue
		}
		results[i] = emb
	}

	return results, firstErr
}

// Dimension returns the expected embedding dimension for the configured model.
func (c *Client) Dimension() int {
	return c.dimension()
}

func (c *Client) embedOne(ctx context.Context, text string) ([]float64, error) {
	url := fmt.Sprintf("%s/models/%s", c.baseURL, c.model)

	body, err := json.Marshal(map[string]string{"inputs": text})
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http error: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("HF API status %d: %s", resp.StatusCode, string(respBody))
	}

	var embResp []float64
	if err := json.Unmarshal(respBody, &embResp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	if len(embResp) == 0 {
		return nil, fmt.Errorf("empty embedding returned")
	}

	return embResp, nil
}

func (c *Client) dimension() int {
	switch c.model {
	case "sentence-transformers/all-MiniLM-L6-v2":
		return 384
	case "sentence-transformers/all-mpnet-base-v2":
		return 768
	case "BAAI/bge-base-en-v1.5":
		return 768
	case "BAAI/bge-large-en-v1.5":
		return 1024
	default:
		return 384
	}
}

// Ensure Client implements the EmbeddingProvider interface at compile time.
var _ ports.EmbeddingProvider = (*Client)(nil)
