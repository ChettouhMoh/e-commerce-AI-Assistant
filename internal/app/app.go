package app

import (
	"fmt"
	"net/http"

	"ecommerce-ai-assistant/internal/assistant/adapters/llmhttp"
	"ecommerce-ai-assistant/internal/assistant/application"
	catalogmemory "ecommerce-ai-assistant/internal/catalog/adapters/memory"
	devchathandler "ecommerce-ai-assistant/internal/channels/devchat"
	conversationmemory "ecommerce-ai-assistant/internal/conversation/adapters/memory"
	"ecommerce-ai-assistant/internal/knowledge/adapters/fake"
	hf "ecommerce-ai-assistant/internal/knowledge/adapters/huggingface"
	knowledgememory "ecommerce-ai-assistant/internal/knowledge/adapters/memory"
	knowledgeports "ecommerce-ai-assistant/internal/knowledge/ports"
	ordermemory "ecommerce-ai-assistant/internal/orders/adapters/memory"
	"ecommerce-ai-assistant/internal/platform/config"
	toolsapplication "ecommerce-ai-assistant/internal/tools/application"
)

// Dependencies holds all wired dependencies for the application.
type Dependencies struct {
	Config       *config.Config
	Embedder     knowledgeports.EmbeddingProvider
	VectorRepo   *knowledgememory.VectorRepository
	LLMProvider  *llmhttp.Client
	Assistant    *application.AssistantService
	DevChat      *devchathandler.Handler
}

// Wire sets up all application dependencies based on configuration.
func Wire(cfg *config.Config) (*Dependencies, error) {
	deps := &Dependencies{Config: cfg}

	// Wire embedding provider based on config
	switch cfg.EmbeddingProvider {
	case "huggingface":
		if cfg.HuggingFaceAPIKey == "" {
			return nil, fmt.Errorf("HUGGINGFACE_API_KEY is required when EMBEDDING_PROVIDER=huggingface")
		}
		deps.Embedder = hf.NewClient(hf.DefaultConfig(
			cfg.HuggingFaceEmbeddingModel,
			cfg.HuggingFaceAPIKey,
		))
	case "fake":
		deps.Embedder = fake.NewFakeEmbeddingProvider()
	default:
		return nil, fmt.Errorf("unsupported embedding provider: %s", cfg.EmbeddingProvider)
	}

	// Wire repositories
	catalogRepo := catalogmemory.NewProductRepository()
	orderRepo := ordermemory.NewOrderRepository()
	vectorRepo := knowledgememory.NewVectorRepository()
	_ = conversationmemory.NewConversationRepository()

	deps.VectorRepo = vectorRepo

	// Wire LLM provider (OpenAI-compatible, works with Groq, Ollama, etc.)
	llmCfg := llmhttp.DefaultConfig(cfg.LLMBaseURL, cfg.LLMApiKey, cfg.LLMModel)
	llmCfg.TimeoutSec = cfg.LLMTimeout
	deps.LLMProvider = llmhttp.NewClient(llmCfg)

	// Wire tool registry with repositories injected via context
	toolRegistry := toolsapplication.NewToolRegistry(
		catalogRepo,
		orderRepo,
		vectorRepo,
	)

	// Wire assistant service
	deps.Assistant = application.NewAssistantService(
		deps.LLMProvider,
		toolRegistry,
		cfg.MaxToolCalls,
	)

	// Wire dev chat handler
	deps.DevChat = devchathandler.NewHandler(deps.Assistant, devchathandler.DefaultConfig())

	return deps, nil
}

// BuildHandler builds the HTTP handler with all routes.
func BuildHandler(deps *Dependencies) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "ok")
	})

	mux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "ready")
	})

	// Dev chat endpoint
	if deps.DevChat != nil {
		mux.HandleFunc("/api/v1/chat", deps.DevChat.ServeHTTP)
	}

	// TODO: Add WhatsApp webhook routes here when ready

	return mux
}
