package memory

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	convdomain "ecommerce-ai-assistant/internal/conversation/domain"
	"ecommerce-ai-assistant/internal/conversation/ports"
)

type ConversationRepository struct {
	conversations map[string]convdomain.Conversation
	byUser        map[string][]string
	processed     map[string]bool
	mu            sync.RWMutex
}

var _ ports.ConversationRepository = (*ConversationRepository)(nil)

func NewConversationRepository() *ConversationRepository {
	r := &ConversationRepository{
		conversations: make(map[string]convdomain.Conversation),
		byUser:        make(map[string][]string),
		processed:     make(map[string]bool),
	}
	r.Seed()
	return r
}

func (r *ConversationRepository) FindByID(ctx context.Context, id string) (*convdomain.Conversation, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	c, ok := r.conversations[id]
	if !ok {
		return nil, nil
	}
	return convCopy(&c), nil
}

func (r *ConversationRepository) Save(ctx context.Context, conv *convdomain.Conversation) error {
	if conv == nil {
		return fmt.Errorf("conversation is nil")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if conv.CreatedAt.IsZero() {
		conv.CreatedAt = time.Now()
	}
	conv.UpdatedAt = time.Now()
	r.conversations[conv.ID] = *conv
	r.byUser[conv.UserID] = appendUnique(r.byUser[conv.UserID], conv.ID)
	return nil
}

func (r *ConversationRepository) FindByUserID(ctx context.Context, userID string, limit, offset int) ([]convdomain.Conversation, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	ids, ok := r.byUser[userID]
	if !ok {
		return nil, nil
	}
	result := make([]convdomain.Conversation, 0, len(ids))
	for _, id := range ids {
		c := r.conversations[id]
		result = append(result, *convCopy(&c))
	}
	sort.Slice(result, func(i, j int) bool { return result[i].UpdatedAt.After(result[j].UpdatedAt) })
	if offset > 0 && offset < len(result) {
		result = result[offset:]
	}
	if limit > 0 && limit < len(result) {
		result = result[:limit]
	}
	return result, nil
}

func (r *ConversationRepository) FindActiveByUserID(ctx context.Context, userID, channelID string) (*convdomain.Conversation, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	ids, ok := r.byUser[userID]
	if !ok {
		return nil, nil
	}
	var latest *convdomain.Conversation
	for _, id := range ids {
		c := r.conversations[id]
		if c.ChannelID != channelID {
			continue
		}
		if latest == nil || c.UpdatedAt.After(latest.UpdatedAt) {
			cp := c
			latest = &cp
		}
	}
	return latest, nil
}

func (r *ConversationRepository) AppendTurn(ctx context.Context, conversationID string, turn *convdomain.Turn) error {
	if turn == nil {
		return fmt.Errorf("turn is nil")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	c, ok := r.conversations[conversationID]
	if !ok {
		return fmt.Errorf("conversation not found: %s", conversationID)
	}
	if turn.CreatedAt.IsZero() {
		turn.CreatedAt = time.Now()
	}
	turn.ConversationID = conversationID
	c.Turns = append(c.Turns, *turn)
	c.UpdatedAt = time.Now()
	r.conversations[conversationID] = c
	return nil
}

func (r *ConversationRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	c, ok := r.conversations[id]
	if !ok {
		return nil
	}
	delete(r.conversations, id)
	r.byUser[c.UserID] = removeFromIndexStr(r.byUser[c.UserID], id)
	return nil
}

func (r *ConversationRepository) DeleteOlderThan(ctx context.Context, olderThan time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for id, c := range r.conversations {
		if c.UpdatedAt.Before(olderThan) {
			delete(r.conversations, id)
			r.byUser[c.UserID] = removeFromIndexStr(r.byUser[c.UserID], id)
		}
	}
	return nil
}

func (r *ConversationRepository) MarkProcessed(ctx context.Context, messageID string) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.processed[messageID] {
		return false, nil
	}
	r.processed[messageID] = true
	return true, nil
}

func appendUnique(slice []string, val string) []string {
	for _, v := range slice {
		if v == val {
			return slice
		}
	}
	return append(slice, val)
}

func copyStringMap(m map[string]string) map[string]string {
	if m == nil {
		return nil
	}
	out := make(map[string]string, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

func removeFromIndexStr(index []string, id string) []string {
	for i, v := range index {
		if v == id {
			index[i] = index[len(index)-1]
			return index[:len(index)-1]
		}
	}
	return index
}

func convCopy(c *convdomain.Conversation) *convdomain.Conversation {
	if c == nil {
		return nil
	}
	cp := *c
	turns := make([]convdomain.Turn, len(c.Turns))
	for i, t := range c.Turns {
		turns[i] = *turnCopy(&t)
	}
	cp.Turns = turns
	cp.Metadata = copyStringMap(c.Metadata)
	return &cp
}

func turnCopy(t *convdomain.Turn) *convdomain.Turn {
	if t == nil {
		return nil
	}
	cp := *t
	cp.ToolCalls = nil
	if len(t.ToolCalls) > 0 {
		cp.ToolCalls = make([]convdomain.ToolCall, len(t.ToolCalls))
		for i, tc := range t.ToolCalls {
			cp.ToolCalls[i] = *toolCallCopy(&tc)
		}
	}
	return &cp
}

func toolCallCopy(tc *convdomain.ToolCall) *convdomain.ToolCall {
	if tc == nil {
		return nil
	}
	cp := *tc
	cp.Arguments = copyStringInterfaceMap(tc.Arguments)
	return &cp
}

func copyStringInterfaceMap(m map[string]interface{}) map[string]interface{} {
	if m == nil {
		return nil
	}
	out := make(map[string]interface{}, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

func (r *ConversationRepository) Seed() {
	now := time.Now()
	conversations := []convdomain.Conversation{
		{
			ID: "CONV-001", ChannelID: "whatsapp", UserID: "CUST-001", CreatedAt: now.Add(-5 * time.Hour), UpdatedAt: now.Add(-4 * time.Hour),
			Turns: []convdomain.Turn{
				{
					ID: "TURN-001", ConversationID: "CONV-001", CreatedAt: now.Add(-5 * time.Hour),
					UserMessage:       convdomain.Message{Role: "user", Content: "Hi, I want to check my order status"},
					AssistantResponse: convdomain.Message{Role: "assistant", Content: "Sure, could you provide your order ID?"},
					TokenUsage:        convdomain.TokenUsage{PromptTokens: 12, CompletionTokens: 8, TotalTokens: 20},
				},
				{
					ID: "TURN-002", ConversationID: "CONV-001", CreatedAt: now.Add(-4 * time.Hour),
					UserMessage:       convdomain.Message{Role: "user", Content: "My order ID is ORD-1001"},
					AssistantResponse: convdomain.Message{Role: "assistant", Content: "Order ORD-1001 is currently shipped and on its way."},
					TokenUsage:        convdomain.TokenUsage{PromptTokens: 18, CompletionTokens: 12, TotalTokens: 30},
				},
			},
		},
		{
			ID: "CONV-002", ChannelID: "devchat", UserID: "CUST-002", CreatedAt: now.Add(-3 * time.Hour), UpdatedAt: now.Add(-1 * time.Hour),
			Turns: []convdomain.Turn{
				{
					ID: "TURN-003", ConversationID: "CONV-002", CreatedAt: now.Add(-3 * time.Hour),
					UserMessage:       convdomain.Message{Role: "user", Content: "Can I return my chinos?"},
					AssistantResponse: convdomain.Message{Role: "assistant", Content: "Yes, unworn chinos can be returned within 30 days."},
					TokenUsage:        convdomain.TokenUsage{PromptTokens: 14, CompletionTokens: 10, TotalTokens: 24},
				},
				{
					ID: "TURN-004", ConversationID: "CONV-002", CreatedAt: now.Add(-2 * time.Hour),
					UserMessage:       convdomain.Message{Role: "user", Content: "What about the t-shirt?"},
					AssistantResponse: convdomain.Message{Role: "assistant", Content: "The linen t-shirt is also returnable within 30 days if unworn."},
					TokenUsage:        convdomain.TokenUsage{PromptTokens: 20, CompletionTokens: 14, TotalTokens: 34},
				},
				{
					ID: "TURN-005", ConversationID: "CONV-002", CreatedAt: now.Add(-1 * time.Hour),
					UserMessage:       convdomain.Message{Role: "user", Content: "Great, how do I start a return?"},
					AssistantResponse: convdomain.Message{Role: "assistant", Content: "You can start a return from the orders page or I can help you with it."},
					TokenUsage:        convdomain.TokenUsage{PromptTokens: 22, CompletionTokens: 16, TotalTokens: 38},
					ToolCalls:         []convdomain.ToolCall{{ID: "TC-001", Name: "initiate_return", Arguments: map[string]interface{}{"order_id": "ORD-1003"}}},
				},
			},
		},
		{
			ID: "CONV-003", ChannelID: "whatsapp", UserID: "CUST-003", CreatedAt: now.Add(-48 * time.Hour), UpdatedAt: now.Add(-47 * time.Hour),
			Turns: []convdomain.Turn{
				{
					ID: "TURN-006", ConversationID: "CONV-003", CreatedAt: now.Add(-48 * time.Hour),
					UserMessage:       convdomain.Message{Role: "user", Content: "Do you have the wool scarf in red?"},
					AssistantResponse: convdomain.Message{Role: "assistant", Content: "Yes, the wool scarf is available in red with SKU SCARF-RED."},
					TokenUsage:        convdomain.TokenUsage{PromptTokens: 15, CompletionTokens: 11, TotalTokens: 26},
				},
			},
		},
	}

	for _, conv := range conversations {
		r.conversations[conv.ID] = conv
		r.byUser[conv.UserID] = append(r.byUser[conv.UserID], conv.ID)
	}
}
