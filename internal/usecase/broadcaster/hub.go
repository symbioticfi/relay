package broadcaster

import (
	"log/slog"
	"sync"
)

// Hub is a generic broadcaster that manages dynamic fan-out of messages to subscribers
// T is the type of messages being broadcast
type Hub[T any] struct {
	mu          sync.RWMutex
	subscribers map[string]*subscriber[T]
	bufferSize  int
}

// subscriber represents a single subscriber with its channel
type subscriber[T any] struct {
	id      string
	ch      chan T
	stopped bool
	mu      sync.Mutex
}

// HubOption configures the Hub
type HubOption[T any] func(*Hub[T])

// WithBufferSize sets the buffer size for subscriber channels
func WithBufferSize[T any](size int) HubOption[T] {
	return func(h *Hub[T]) {
		h.bufferSize = size
	}
}

// NewHub creates a new Hub with optional configuration
func NewHub[T any](opts ...HubOption[T]) *Hub[T] {
	h := &Hub[T]{
		subscribers: make(map[string]*subscriber[T]),
		bufferSize:  100, // default buffer size
	}

	for _, opt := range opts {
		opt(h)
	}

	return h
}

// Subscribe adds a new subscriber with the given ID and returns the channel to receive messages
// If a subscriber with the same ID already exists, it returns the existing channel
func (h *Hub[T]) Subscribe(id string) chan T {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Check if subscriber already exists
	if sub, exists := h.subscribers[id]; exists {
		slog.Warn(
			"Subscriber already exists",
			"id", id,
		)

		return sub.ch
	}

	ch := make(chan T, h.bufferSize)

	h.subscribers[id] = &subscriber[T]{
		id:      id,
		ch:      ch,
		stopped: false,
	}

	slog.Debug(
		"Subscriber added",
		"id", id,
		"totalSubscribers", len(h.subscribers),
	)

	return ch
}

// Unsubscribe removes a subscriber and closes its channel
func (h *Hub[T]) Unsubscribe(id string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	sub, exists := h.subscribers[id]
	if !exists {
		return
	}

	sub.mu.Lock()
	if !sub.stopped {
		close(sub.ch)
		sub.stopped = true
	}
	sub.mu.Unlock()

	delete(h.subscribers, id)

	slog.Debug(
		"Subscriber removed",
		"id", id,
		"remainingSubscribers", len(h.subscribers),
	)
}

// Broadcast sends a message to all active subscribers
// Slow subscribers will have messages dropped (non-blocking send)
func (h *Hub[T]) Broadcast(msg T) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if len(h.subscribers) == 0 {
		return
	}

	droppedCount := 0
	for _, sub := range h.subscribers {
		sub.mu.Lock()
		stopped := sub.stopped
		sub.mu.Unlock()

		if stopped {
			continue
		}

		select {
		case sub.ch <- msg:
			// successfully sent
		default:
			// channel is full, drop message
			droppedCount++
			slog.Warn(
				"Dropped message for slow subscriber",
				"subscriberId", sub.id,
				"bufferSize", h.bufferSize,
			)
		}
	}

	if droppedCount > 0 {
		slog.Warn(
			"Broadcast completed with drops",
			"totalSubscribers", len(h.subscribers),
			"droppedCount", droppedCount,
		)
	}
}

// BroadcastWait sends a message to all active subscribers and waits for all sends to complete
// This is a blocking operation that ensures all subscribers receive the message
func (h *Hub[T]) BroadcastWait(msg T) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if len(h.subscribers) == 0 {
		return
	}

	for _, sub := range h.subscribers {
		sub.mu.Lock()
		stopped := sub.stopped
		sub.mu.Unlock()

		if stopped {
			continue
		}

		// Blocking send
		sub.ch <- msg
	}
}

// Count returns the number of active subscribers
func (h *Hub[T]) Count() int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return len(h.subscribers)
}

// Close unsubscribes all subscribers and cleans up resources
func (h *Hub[T]) Close() {
	h.mu.Lock()
	defer h.mu.Unlock()

	for id, sub := range h.subscribers {
		sub.mu.Lock()

		if !sub.stopped {
			close(sub.ch)
			sub.stopped = true
		}

		sub.mu.Unlock()

		delete(h.subscribers, id)
	}

	slog.Info("Hub closed, all subscribers removed")
}
