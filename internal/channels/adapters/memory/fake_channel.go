package memory

import (
	"context"
	"sync"

	channelsdomain "ecommerce-ai-assistant/internal/channels/domain"
	"ecommerce-ai-assistant/internal/channels/ports"
)

type FakeChannel struct {
	sent     []channelsdomain.OutboundMessage
	handlers []ports.InboundMessageHandler
	started  bool
	stopCh   chan struct{}
	mu       sync.RWMutex
}

var _ ports.MessagingChannel = (*FakeChannel)(nil)

func NewFakeChannel() *FakeChannel {
	return &FakeChannel{
		sent:     make([]channelsdomain.OutboundMessage, 0),
		handlers: make([]ports.InboundMessageHandler, 0),
	}
}

func (f *FakeChannel) Send(ctx context.Context, msg channelsdomain.OutboundMessage) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.sent = append(f.sent, msg)
	return nil
}

func (f *FakeChannel) SendBatch(ctx context.Context, msgs []channelsdomain.OutboundMessage) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.sent = append(f.sent, msgs...)
	return nil
}

func (f *FakeChannel) Subscribe(handler ports.InboundMessageHandler) error {
	if handler == nil {
		return nil
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.handlers = append(f.handlers, handler)
	return nil
}

func (f *FakeChannel) Start(ctx context.Context) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.started {
		return nil
	}
	f.started = true
	f.stopCh = make(chan struct{})
	return nil
}

func (f *FakeChannel) Stop(ctx context.Context) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if !f.started {
		return nil
	}
	f.started = false
	if f.stopCh != nil {
		close(f.stopCh)
		f.stopCh = nil
	}
	return nil
}

func (f *FakeChannel) SentMessages() []channelsdomain.OutboundMessage {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return append([]channelsdomain.OutboundMessage(nil), f.sent...)
}

func (f *FakeChannel) Reset() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.sent = make([]channelsdomain.OutboundMessage, 0)
	f.handlers = make([]ports.InboundMessageHandler, 0)
	f.started = false
	if f.stopCh != nil {
		close(f.stopCh)
		f.stopCh = nil
	}
}

func (f *FakeChannel) EmitInbound(ctx context.Context, msg channelsdomain.InboundMessage) error {
	f.mu.RLock()
	handlers := append([]ports.InboundMessageHandler(nil), f.handlers...)
	f.mu.RUnlock()
	for _, h := range handlers {
		if err := h(ctx, msg); err != nil {
			return err
		}
	}
	return nil
}
