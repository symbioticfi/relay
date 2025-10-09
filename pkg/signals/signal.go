package signals

import (
	"context"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-errors/errors"
)

const (
	// EmitTimeout is the default timeout for emitting events to the signal queue.
	EmitTimeout = 10 * time.Second
)

// Signal provides a type-safe, concurrent event processing system with configurable worker pools.
// It uses a buffered channel queue to handle events and multiple worker goroutines for parallel processing.
type Signal[T any] struct {
	queue      chan Event[T]
	handlers   []SignalListener[T]
	maxWorkers int
	id         string

	// Internal state
	started bool
	stopped atomic.Bool
	mutex   sync.RWMutex
}

// Config defines the configuration for a Signal instance.
type Config struct {
	// BufferSize sets the size of the internal event queue buffer (minimum 5)
	BufferSize int `mapstructure:"buffer-size" validate:"gte=5"`
	// WorkerCount sets the number of worker goroutines to process events (5-100)
	WorkerCount int `mapstructure:"worker-count" validate:"gte=5,lte=100"`
}

func DefaultConfig() Config {
	return Config{
		BufferSize:  10,
		WorkerCount: 10,
	}
}

// Event represents a signal event containing both payload data and execution context.
type Event[T any] struct {
	// Payload contains the actual event data
	Payload T
	// Ctx is the context associated with this event for cancellation and timeout handling
	Ctx context.Context
}

// SignalListener defines the function signature for handling signal events.
// It receives the event context and payload, returning an error if processing fails.
type SignalListener[T any] func(context.Context, T) error

// New creates a new Signal instance with the specified configuration.
// Handlers can be provided during construction or set later using SetHandlers.
// The id is used for logging and error identification.
// Nil handlers are automatically filtered out.
func New[T any](cfg Config, id string, handlers ...SignalListener[T]) *Signal[T] {
	// Filter out nil handlers
	var validHandlers []SignalListener[T]
	for _, h := range handlers {
		if h != nil {
			validHandlers = append(validHandlers, h)
		}
	}

	return &Signal[T]{
		queue:      make(chan Event[T], cfg.BufferSize),
		handlers:   validHandlers,
		maxWorkers: cfg.WorkerCount,
		id:         id,
	}
}

// SetHandlers sets the event handlers for this signal.
// Returns an error if workers are started or handlers are already set, as handlers cannot be replaced.
// This method is thread-safe and should be called before starting workers.
// Accepts one or more handlers which will be executed sequentially in the order provided.
// Nil handlers are automatically filtered out.
func (s *Signal[T]) SetHandlers(handlers ...SignalListener[T]) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if s.started {
		return errors.Errorf("signal workers are already started for %v signal, cannot set handler", s.id)
	}
	if len(s.handlers) > 0 {
		return errors.Errorf("signal handlers are already set for %v signal, cannot reset handlers", s.id)
	}

	// Filter out nil handlers
	var validHandlers []SignalListener[T]
	for _, h := range handlers {
		if h != nil {
			validHandlers = append(validHandlers, h)
		}
	}

	if len(validHandlers) == 0 {
		return errors.Errorf("at least one non-nil handler must be provided for %v signal", s.id)
	}
	s.handlers = validHandlers
	return nil
}

// Emit sends an event to the signal queue using the default EmitTimeout.
// It will block if the queue is full until the timeout is reached.
// Returns an error if workers have stopped.
// Use EmitNonBlocking if you want non-blocking behavior or EmitWithTimeout for custom timeouts.
func (s *Signal[T]) Emit(payload T) error {
	return s.EmitWithTimeout(payload, EmitTimeout)
}

// EmitWithTimeout sends an event to the signal queue with a custom timeout.
// Returns an error if the event cannot be queued within the timeout period or if workers have stopped.
func (s *Signal[T]) EmitWithTimeout(payload T, timeout time.Duration) error {
	if s.stopped.Load() {
		return errors.Errorf("cannot emit to stopped signal %v", s.id)
	}

	s.mutex.RLock()
	// We might not have handlers set for specific signals like the signals dealing
	// with aggregation on non aggregator nodes in such case we ignore the event completely
	// if we don't ignore the queue will get full and requests will timeout
	if len(s.handlers) == 0 {
		s.mutex.RUnlock()
		return nil
	}
	s.mutex.RUnlock()

	event := Event[T]{Payload: payload}

	emitCtx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	select {
	case s.queue <- event:
		return nil
	case <-emitCtx.Done():
		return errors.Errorf("failed to emit signal: %w", emitCtx.Err())
	}
}

// EmitNonBlocking attempts to send an event without blocking.
// Returns true if the event was sent, false if the queue is full or workers have stopped.
func (s *Signal[T]) EmitNonBlocking(ctx context.Context, payload T) bool {
	if s.stopped.Load() {
		return false
	}

	select {
	case s.queue <- Event[T]{Payload: payload, Ctx: ctx}:
		return true
	default:
		return false
	}
}

// StartWorkers starts the configured number of worker goroutines to process events from the queue.
// It should be called once during application startup and cannot be called again.
// All workers will gracefully shut down when the provided context is cancelled.
// The signal automatically marks itself as stopped and closes the queue when all workers exit.
// Returns an error if workers are already started or if no signal handler is set.
func (s *Signal[T]) StartWorkers(ctx context.Context) error {
	s.mutex.Lock()
	if s.started {
		s.mutex.Unlock()
		return errors.Errorf("signal workers are already started for %v signal", s.id)
	}
	if len(s.handlers) == 0 {
		s.mutex.Unlock()
		return errors.Errorf("signal handlers are not set for %v signal, cannot start workers", s.id)
	}
	handlers := s.handlers
	s.started = true
	s.mutex.Unlock()

	var shutdownWG sync.WaitGroup
	shutdownWG.Add(s.maxWorkers)

	for i := 0; i < s.maxWorkers; i++ {
		go func(j int) {
			defer shutdownWG.Done()
			for {
				select {
				case <-ctx.Done():
					slog.Info("shutting down signal worker", slog.Int("worker", j), slog.String("signal", s.id))
					return
				case event, ok := <-s.queue:
					if !ok {
						slog.Info("signal queue closed, shutting down worker", slog.Int("worker", j), slog.String("signal", s.id))
						return
					}
					// Execute all handlers regardless of errors
					for i, handler := range handlers {
						if err := handler(ctx, event.Payload); err != nil {
							slog.Error("handler failed",
								slog.Any("error", err),
								slog.Int("handler_index", i),
								slog.Int("worker", j),
								slog.Any("event", event.Payload),
								slog.String("signal", s.id))
							// Continue executing remaining handlers
						}
					}
				}
			}
		}(i)
	}

	go func() {
		// wait for all workers to finish
		shutdownWG.Wait()
		s.stopped.Store(true)
		slog.Info("all signal workers stopped", slog.String("signal", s.id))
	}()

	return nil
}
