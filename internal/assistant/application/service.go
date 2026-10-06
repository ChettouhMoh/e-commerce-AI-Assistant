package application

import (
	"context"
	"fmt"
	"strings"
	"time"

	"ecommerce-ai-assistant/internal/assistant/domain"
	"ecommerce-ai-assistant/internal/assistant/ports"
)

// AssistantService orchestrates the LLM tool-call loop with bounded iterations.
type AssistantService struct {
	llm      ports.LLMProvider
	tools    ports.ToolRegistry
	maxCalls int
}

// NewAssistantService creates a new AssistantService.
// maxCalls bounds the number of consecutive LLM tool-call iterations (default 5).
func NewAssistantService(llm ports.LLMProvider, tools ports.ToolRegistry, maxCalls int) *AssistantService {
	if maxCalls <= 0 {
		maxCalls = 5
	}
	return &AssistantService{llm: llm, tools: tools, maxCalls: maxCalls}
}

// Run satisfies the channels.AssistantUseCase interface. It loads or creates
// a conversation, delegates to HandleMessage, and persists the conversation.
func (s *AssistantService) Run(ctx context.Context, conversationID, userID, message string) (*domain.AssistantResponse, error) {
	if conversationID == "" || userID == "" {
		return nil, fmt.Errorf("conversationID and userID are required")
	}
	if message == "" {
		return nil, fmt.Errorf("message is required")
	}

	conv := &domain.Conversation{
		ID:        conversationID,
		ChannelID: userID,
		Turns:     []domain.Turn{},
		Metadata:  make(map[string]string),
	}

	resp, err := s.HandleMessage(ctx, conv, message)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// HandleMessage processes a user message and returns the assistant response.
// It runs the bounded LLM tool-call loop: at each iteration the LLM may return
// tool calls; those are executed and fed back as observation messages until
// the LLM returns a final text response or maxCalls is reached.
func (s *AssistantService) HandleMessage(ctx context.Context, conv *domain.Conversation, message string) (*domain.AssistantResponse, error) {
	conv.Turns = append(conv.Turns, domain.Turn{Role: domain.RoleUser, Content: message, CreatedAt: time.Now()})

	messages := s.buildMessages(conv)
	var response *domain.AssistantResponse
	var err error

	for i := 0; i < s.maxCalls; i++ {
		toolDefs := s.tools.List()
		response, err = s.llm.Complete(ctx, messages, toolDefs)
		if err != nil {
			return nil, fmt.Errorf("LLM completion failed: %w", err)
		}

		if len(response.ToolCalls) == 0 {
			break
		}

		toolResults, err := s.executeTools(ctx, response.ToolCalls)
		if err != nil {
			return nil, err
		}

		for _, tc := range response.ToolCalls {
			content := toolResults[tc.ID]
			messages = append(messages, domain.Message{Role: domain.RoleUser, Content: fmt.Sprintf("Tool %s returned: %s", tc.Name, content)})
		}
	}

	conv.Turns = append(conv.Turns, domain.Turn{Role: domain.RoleAssistant, Content: strings.Join(s.extractTexts(response.Messages), "\n"), CreatedAt: time.Now()})
	return response, nil
}

// buildMessages converts conversation turns into the flat message list for the LLM.
func (s *AssistantService) buildMessages(conv *domain.Conversation) []domain.Message {
	var msgs []domain.Message
	for _, t := range conv.Turns {
		if t.Content != "" {
			msgs = append(msgs, domain.Message{Role: t.Role, Content: t.Content})
		}
		for _, tr := range t.ToolResults {
			msgs = append(msgs, domain.Message{Role: domain.RoleUser, Content: tr.Content})
		}
	}
	return msgs
}

// executeTools runs all tool calls from an LLM response and returns their results.
func (s *AssistantService) executeTools(ctx context.Context, calls []domain.ToolCall) (map[string]string, error) {
	results := make(map[string]string)
	for _, call := range calls {
		result, err := s.tools.Execute(ctx, domain.ToolCallRequest{CallID: call.ID, Tool: call.Name, Args: call.Arguments})
		if err != nil {
			results[call.ID] = fmt.Sprintf("Error: %v", err)
			continue
		}
		results[call.ID] = result.Content
	}
	return results, nil
}

// extractTexts collects non-empty text from assistant messages.
func (s *AssistantService) extractTexts(msgs []domain.Message) []string {
	var texts []string
	for _, m := range msgs {
		if m.Content != "" {
			texts = append(texts, m.Content)
		}
	}
	return texts
}
