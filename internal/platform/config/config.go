package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	AppEnv                           string
	HTTPAddr                         string
	LLMBaseURL                       string
	LLMApiKey                        string
	LLMModel                         string
	EmbeddingProvider                string
	OllamaBaseURL                    string
	OllamaEmbeddingModel             string
	WhatsAppVerifyToken              string
	WhatsAppAppSecret                string
	WhatsAppAccessToken              string
	WhatsAppPhoneNumberID            string
	WhatsAppGraphAPIVersion          string
	WhatsAppWebhookSignatureRequired bool
	MaxToolCalls                     int
	TopK                             int
	ChunkSize                        int
	ChunkOverlap                     int
	RequestTimeout                   int
	LLMTimeout                       int
}

func Default() *Config {
	return &Config{
		AppEnv:                           getEnv("APP_ENV", "local"),
		HTTPAddr:                         getEnv("HTTP_ADDR", ":8080"),
		LLMBaseURL:                       getEnv("LLM_BASE_URL", "http://localhost:11434/v1"),
		LLMApiKey:                        getEnv("LLM_API_KEY", "ollama"),
		LLMModel:                         getEnv("LLM_MODEL", "llama3.2"),
		EmbeddingProvider:                getEnv("EMBEDDING_PROVIDER", "ollama"),
		OllamaBaseURL:                    getEnv("OLLAMA_BASE_URL", "http://localhost:11434"),
		OllamaEmbeddingModel:             getEnv("OLLAMA_EMBEDDING_MODEL", "nomic-embed-text"),
		WhatsAppVerifyToken:              getEnv("WHATSAPP_VERIFY_TOKEN", ""),
		WhatsAppAppSecret:                getEnv("WHATSAPP_APP_SECRET", ""),
		WhatsAppAccessToken:              getEnv("WHATSAPP_ACCESS_TOKEN", ""),
		WhatsAppPhoneNumberID:            getEnv("WHATSAPP_PHONE_NUMBER_ID", ""),
		WhatsAppGraphAPIVersion:          getEnv("WHATSAPP_GRAPH_API_VERSION", "v18.0"),
		WhatsAppWebhookSignatureRequired: getEnvBool("WHATSAPP_WEBHOOK_SIGNATURE_REQUIRED", false),
		MaxToolCalls:                     getEnvInt("MAX_TOOL_CALLS", 5),
		TopK:                             getEnvInt("TOP_K", 5),
		ChunkSize:                        getEnvInt("CHUNK_SIZE", 1000),
		ChunkOverlap:                     getEnvInt("CHUNK_OVERLAP", 200),
		RequestTimeout:                   getEnvInt("REQUEST_TIMEOUT", 30),
		LLMTimeout:                       getEnvInt("LLM_TIMEOUT", 60),
	}
}

func (c *Config) Validate() error {
	if c.HTTPAddr == "" {
		return fmt.Errorf("HTTP_ADDR is required")
	}
	if c.MaxToolCalls <= 0 {
		return fmt.Errorf("MAX_TOOL_CALLS must be positive")
	}
	if c.TopK <= 0 {
		return fmt.Errorf("TOP_K must be positive")
	}
	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if b, err := strconv.ParseBool(value); err == nil {
			return b
		}
	}
	return defaultValue
}
