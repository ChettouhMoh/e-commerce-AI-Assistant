package ollama

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Config struct {
	BaseURL            string
	EmbeddingModel     string
	RequestTimeoutSecs int
}

func DefaultConfig(baseURL, model string) Config {
	return Config{
		BaseURL:            baseURL,
		EmbeddingModel:     model,
		RequestTimeoutSecs: 60,
	}
}

type Client struct {
	baseURL    string
	model      string
	httpClient *http.Client
}

func NewClient(cfg Config) *Client {
	return &Client{
		baseURL:    strings.TrimRight(cfg.BaseURL, "/"),
		model:      cfg.EmbeddingModel,
		httpClient: &http.Client{Timeout: time.Duration(cfg.RequestTimeoutSecs) * time.Second},
	}
}

type embeddingRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

type embeddingResponse struct {
	Embedding []float64 `json:"embedding"`
	Model     string    `json:"model"`
}

func (c *Client) embedOne(ctx context.Context, text string) ([]float64, error) {
	body := embeddingRequest{Model: c.model, Prompt: text}
	b, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/embeddings", strings.NewReader(string(b)))
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
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
		return nil, fmt.Errorf("ollama API status %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	var embResp embeddingResponse
	if err := json.Unmarshal(respBody, &embResp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	if embResp.Embedding == nil {
		return nil, fmt.Errorf("empty embedding returned")
	}

	return embResp.Embedding, nil
}

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
				firstErr = err
			}
			results[i] = make([]float64, c.dimension())
			continue
		}
		results[i] = emb
	}

	return results, firstErr
}

func (c *Client) Dimension() int {
	return c.dimension()
}

func (c *Client) dimension() int {
	if c == nil || c.model == "" {
		return 384
	}
	switch c.model {
	case "nomic-embed-text":
		return 768
	case "mxbai-embed-large":
		return 1024
	case "all-minilm":
		return 384
	default:
		return 384
	}
}
