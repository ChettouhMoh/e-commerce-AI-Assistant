// Package llmhttp provides a fake LLM provider for testing.
package llmhttp

import (
	"context"
	"encoding/json"
	"sync"

	"ecommerce-ai-assistant/internal/assistant/domain"
)

// FakeResponse is a pre-programmed response returned by the fake provider.
type FakeResponse struct {
	ToolCalls []domain.ToolCall
	Content   string
	Error     error
	Usage     domain.Usage
}

// FakeProvider is a fake implementation of domain.LLMProvider for testing.
type FakeProvider struct {
	mu           sync.Mutex
	responses    []FakeResponse
	nextIdx      int
	calls        int
	defaultResp  FakeResponse
	recorded     []FakeResponse
	lastMessages []domain.Message
	lastTools    []domain.ToolDefinition
}

// NewFakeProvider creates a new fake LLM provider.
// If no responses are provided, it returns a default non-tool-calling response.
func NewFakeProvider(responses []FakeResponse, defaultResponse FakeResponse) *FakeProvider {
	if len(responses) == 0 && defaultResponse.Content == "" && len(defaultResponse.ToolCalls) == 0 {
		defaultResponse = FakeResponse{
			Content: "I can help you with that.",
			Usage:   domain.Usage{InputTokens: 10, OutputTokens: 5},
		}
	}
	return &FakeProvider{
		responses:   responses,
		defaultResp: defaultResponse,
		recorded:    make([]FakeResponse, 0),
	}
}

// Complete returns the next pre-programmed response, recording the call.
func (f *FakeProvider) Complete(ctx context.Context, messages []domain.Message, tools []domain.ToolDefinition) (*domain.AssistantResponse, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.calls++
	f.lastMessages = messages
	f.lastTools = tools

	var resp FakeResponse
	if f.nextIdx < len(f.responses) {
		resp = f.responses[f.nextIdx]
		f.nextIdx++
	} else {
		resp = f.defaultResp
	}

	f.recorded = append(f.recorded, resp)

	if resp.Error != nil {
		return nil, resp.Error
	}

	response := &domain.AssistantResponse{
		Messages:   []domain.Message{{Role: domain.RoleAssistant, Content: resp.Content}},
		ToolCalls:  resp.ToolCalls,
		StopReason: "end_turn",
		Usage:      resp.Usage,
	}
	if len(resp.ToolCalls) > 0 {
		response.StopReason = "tool_calls"
		response.Messages = nil
	}

	return response, nil
}

// CallCount returns the number of times Complete was called.
func (f *FakeProvider) CallCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.calls
}

// LastMessages returns the messages from the most recent Complete call.
func (f *FakeProvider) LastMessages() []domain.Message {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.lastMessages == nil {
		return nil
	}
	out := make([]domain.Message, len(f.lastMessages))
	copy(out, f.lastMessages)
	return out
}

// LastTools returns the tool definitions from the most recent Complete call.
func (f *FakeProvider) LastTools() []domain.ToolDefinition {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.lastTools == nil {
		return nil
	}
	out := make([]domain.ToolDefinition, len(f.lastTools))
	copy(out, f.lastTools)
	return out
}

// Reset clears the recorded calls and resets the response index.
func (f *FakeProvider) Reset() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls = 0
	f.nextIdx = 0
	f.recorded = make([]FakeResponse, 0)
	f.lastMessages = nil
	f.lastTools = nil
}

// RecordedResponses returns all responses that were returned so far.
func (f *FakeProvider) RecordedResponses() []FakeResponse {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]FakeResponse, len(f.recorded))
	copy(out, f.recorded)
	return out
}

// SetResponses replaces the list of pre-programmed responses and resets the index.
func (f *FakeProvider) SetResponses(responses []FakeResponse) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.responses = responses
	f.nextIdx = 0
}

// MustMarshalJSON is a test helper to marshal a value to JSON string.
func MustMarshalJSON(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return string(b)
}

// --- Test fixture helpers ---

// WithToolCall creates a FakeResponse that triggers a tool call.
func WithToolCall(callID, toolName string, args map[string]any) FakeResponse {
	if args == nil {
		args = make(map[string]any)
	}
	return FakeResponse{
		ToolCalls: []domain.ToolCall{
			{ID: callID, Name: toolName, Arguments: args},
		},
		Usage: domain.Usage{InputTokens: 20, OutputTokens: 10},
	}
}

// WithTextResponse creates a FakeResponse with a text message.
func WithTextResponse(text string, usage domain.Usage) FakeResponse {
	if usage.InputTokens == 0 && usage.OutputTokens == 0 {
		usage = domain.Usage{InputTokens: 10, OutputTokens: 5}
	}
	return FakeResponse{
		Content: text,
		Usage:   usage,
	}
}

// WithError creates a FakeResponse that returns an error.
func WithError(err error) FakeResponse {
	return FakeResponse{Error: err}
}

// Sequence returns the given responses in order.
func Sequence(responses ...FakeResponse) []FakeResponse {
	return responses
}
