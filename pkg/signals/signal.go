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
	handler    SignalListener[T]
	maxWorkers int
	id         string

	// Internal state
	started bool
	stopped atomic.Bool
	mutex   sync.Mutex
}

// Config defines the configuration for a Signal instance.
type Config struct {
	// BufferSize sets the size of the internal event queue buffer (minimum 5)
	BufferSize int `mapstructure:"buffer-size" validate:"gte=5"`
	// WorkerCount sets the number of worker goroutines to process events (5-100)
	WorkerCount int `mapstructure:"worker-count" validate:"gte=5,lte=100"`
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
// The handler can be nil and set later using SetHandler.
// The id is used for logging and error identification.
func New[T any](cfg Config, id string, handler SignalListener[T]) *Signal[T] {
	return &Signal[T]{
		queue:      make(chan Event[T], cfg.BufferSize),
		handler:    handler,
		maxWorkers: cfg.WorkerCount,
		id:         id,
	}
}

// SetHandler sets the event handler for this signal.
// Returns an error if workers are started or handler is already set, as handlers cannot be replaced.
// This method is thread-safe and should be called before starting workers.
func (s *Signal[T]) SetHandler(handler SignalListener[T]) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if s.started {
		return errors.Errorf("signal workers are already started for %v signal, cannot set handler", s.id)
	}
	if s.handler != nil {
		return errors.Errorf("signal handler is already set for %v signal, cannot reset handler", s.id)
	}
	s.handler = handler
	return nil
}

// Emit sends an event to the signal queue using the default EmitTimeout.
// It will block if the queue is full until the timeout is reached.
// Returns an error if workers have stopped.
// Use EmitNonBlocking if you want non-blocking behavior or EmitWithTimeout for custom timeouts.
func (s *Signal[T]) Emit(ctx context.Context, payload T) error {
	return s.EmitWithTimeout(ctx, payload, EmitTimeout)
}

// EmitWithTimeout sends an event to the signal queue with a custom timeout.
// Returns an error if the event cannot be queued within the timeout period or if workers have stopped.
// The provided context is preserved and passed to the event handler when processed.
func (s *Signal[T]) EmitWithTimeout(ctx context.Context, payload T, timeout time.Duration) error {
	if s.stopped.Load() {
		return errors.Errorf("cannot emit to stopped signal %v", s.id)
	}

	event := Event[T]{Payload: payload, Ctx: ctx}

	emitCtx, cancel := context.WithTimeout(ctx, timeout)
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
	if s.handler == nil {
		s.mutex.Unlock()
		return errors.Errorf("signal handler is not set for %v signal, cannot start workers", s.id)
	}
	handler := s.handler
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
					//nolint:contextcheck // we need to use the context passed as part of the emit event
					if err := handler(event.Ctx, event.Payload); err != nil {
						slog.Error("failed to handle signal", slog.Any("error", err), slog.Int("worker", j), slog.Any("event", event.Payload), slog.String("signal", s.id))
					}
				}
			}
		}(i)
	}

	go func() {
		// wait for all workers to finish
		shutdownWG.Wait()
		s.stopped.Store(true)
		slog.Warn("all signal workers stopped", slog.String("signal", s.id))
	}()

	return nil
}
