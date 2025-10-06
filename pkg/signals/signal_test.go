package signals

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		cfg              Config
		id               string
		handlers         []SignalListener[string]
		expectedHandlers int
	}{
		{
			name: "creates signal with valid handlers",
			cfg:  DefaultConfig(),
			id:   "test-signal",
			handlers: []SignalListener[string]{
				func(ctx context.Context, s string) error { return nil },
				func(ctx context.Context, s string) error { return nil },
			},
			expectedHandlers: 2,
		},
		{
			name:             "creates signal with no handlers",
			cfg:              DefaultConfig(),
			id:               "test-signal",
			handlers:         nil,
			expectedHandlers: 0,
		},
		{
			name: "filters out nil handlers",
			cfg:  DefaultConfig(),
			id:   "test-signal",
			handlers: []SignalListener[string]{
				func(ctx context.Context, s string) error { return nil },
				nil,
				func(ctx context.Context, s string) error { return nil },
			},
			expectedHandlers: 2,
		},
		{
			name: "creates signal with custom config",
			cfg: Config{
				BufferSize:  50,
				WorkerCount: 20,
			},
			id:               "custom-signal",
			handlers:         nil,
			expectedHandlers: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			sig := New(tt.cfg, tt.id, tt.handlers...)

			require.NotNil(t, sig)
			require.Equal(t, tt.id, sig.id)
			require.Equal(t, tt.cfg.WorkerCount, sig.maxWorkers)
			require.Len(t, sig.handlers, tt.expectedHandlers)
			require.NotNil(t, sig.queue)
			require.Equal(t, tt.cfg.BufferSize, cap(sig.queue))
			require.False(t, sig.started)
			require.False(t, sig.stopped.Load())
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	t.Parallel()

	cfg := DefaultConfig()
	require.Equal(t, 10, cfg.BufferSize)
	require.Equal(t, 10, cfg.WorkerCount)
}

func TestSetHandler_Success(t *testing.T) {
	t.Parallel()

	sig := New[string](DefaultConfig(), "test-signal")

	err := sig.SetHandlers(func(ctx context.Context, s string) error { return nil })

	require.NoError(t, err)
	require.NotEmpty(t, sig.handlers)
}

func TestSetHandler_FailsWhenHandlersAlreadySet(t *testing.T) {
	t.Parallel()

	sig := New[string](DefaultConfig(), "test-signal")
	sig.handlers = []SignalListener[string]{
		func(ctx context.Context, s string) error { return nil },
	}

	err := sig.SetHandlers(func(ctx context.Context, s string) error { return nil })

	require.Error(t, err)
	require.Contains(t, err.Error(), "handlers are already set")
}

func TestSetHandler_FailsWhenWorkersAlreadyStarted(t *testing.T) {
	t.Parallel()

	sig := New[string](DefaultConfig(), "test-signal")
	sig.started = true

	err := sig.SetHandlers(func(ctx context.Context, s string) error { return nil })

	require.Error(t, err)
	require.Contains(t, err.Error(), "workers are already started")
}

func TestSetHandler_FailsWhenAllHandlersAreNil(t *testing.T) {
	t.Parallel()

	sig := New[string](DefaultConfig(), "test-signal")

	err := sig.SetHandlers(nil, nil)

	require.Error(t, err)
	require.Contains(t, err.Error(), "at least one non-nil handler must be provided")
}

func TestSetHandler_FiltersNilHandlers(t *testing.T) {
	t.Parallel()

	sig := New[string](DefaultConfig(), "test-signal")

	err := sig.SetHandlers(
		nil,
		func(ctx context.Context, s string) error { return nil },
		nil,
	)

	require.NoError(t, err)
	require.NotEmpty(t, sig.handlers)
	require.Len(t, sig.handlers, 1)
}

func TestEmit_SuccessWithHandlers(t *testing.T) {
	t.Parallel()

	sig := New[string](DefaultConfig(), "test-signal")
	sig.handlers = []SignalListener[string]{
		func(ctx context.Context, s string) error { return nil },
	}

	err := sig.Emit("test-payload")

	require.NoError(t, err)
}

func TestEmit_ReturnsNilWhenNoHandlersSet(t *testing.T) {
	t.Parallel()

	sig := New[string](DefaultConfig(), "test-signal")

	err := sig.Emit("test-payload")

	require.NoError(t, err)
}

func TestEmit_FailsWhenSignalIsStopped(t *testing.T) {
	t.Parallel()

	sig := New[string](DefaultConfig(), "test-signal")
	sig.stopped.Store(true)

	err := sig.Emit("test-payload")

	require.Error(t, err)
	require.Contains(t, err.Error(), "cannot emit to stopped signal")
}

func TestEmitWithTimeout_SuccessWithinTimeout(t *testing.T) {
	t.Parallel()

	sig := New[string](DefaultConfig(), "test-signal")
	sig.handlers = []SignalListener[string]{
		func(ctx context.Context, s string) error { return nil },
	}

	err := sig.EmitWithTimeout("test-payload", 1*time.Second)

	require.NoError(t, err)
}

func TestEmitWithTimeout_TimesOutWhenQueueIsFull(t *testing.T) {
	t.Parallel()

	sig := New[string](DefaultConfig(), "test-signal")
	sig.handlers = []SignalListener[string]{
		func(ctx context.Context, s string) error { return nil },
	}
	// Fill the queue
	for i := 0; i < cap(sig.queue); i++ {
		sig.queue <- Event[string]{Payload: "filler"}
	}

	err := sig.EmitWithTimeout("test-payload", 10*time.Millisecond)

	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to emit signal")
}

func TestEmitWithTimeout_ReturnsNilWhenNoHandlersSet(t *testing.T) {
	t.Parallel()

	sig := New[string](DefaultConfig(), "test-signal")

	err := sig.EmitWithTimeout("test-payload", 1*time.Second)

	require.NoError(t, err)
}

func TestEmitNonBlocking_SuccessWhenQueueHasSpace(t *testing.T) {
	t.Parallel()

	sig := New[string](DefaultConfig(), "test-signal")
	ctx := context.Background()

	result := sig.EmitNonBlocking(ctx, "test-payload")

	require.True(t, result)
}

func TestEmitNonBlocking_ReturnsFalseWhenQueueIsFull(t *testing.T) {
	t.Parallel()

	sig := New[string](DefaultConfig(), "test-signal")
	// Fill the queue
	for i := 0; i < cap(sig.queue); i++ {
		sig.queue <- Event[string]{Payload: "filler"}
	}
	ctx := context.Background()

	result := sig.EmitNonBlocking(ctx, "test-payload")

	require.False(t, result)
}

func TestEmitNonBlocking_ReturnsFalseWhenSignalIsStopped(t *testing.T) {
	t.Parallel()

	sig := New[string](DefaultConfig(), "test-signal")
	sig.stopped.Store(true)
	ctx := context.Background()

	result := sig.EmitNonBlocking(ctx, "test-payload")

	require.False(t, result)
}

func TestStartWorkers_Success(t *testing.T) {
	t.Parallel()

	sig := New[string](DefaultConfig(), "test-signal")
	sig.handlers = []SignalListener[string]{
		func(ctx context.Context, s string) error { return nil },
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := sig.StartWorkers(ctx)

	require.NoError(t, err)
	require.True(t, sig.started)
}

func TestStartWorkers_FailsWhenNoHandlersSet(t *testing.T) {
	t.Parallel()

	sig := New[string](DefaultConfig(), "test-signal")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := sig.StartWorkers(ctx)

	require.Error(t, err)
	require.Contains(t, err.Error(), "handlers are not set")
}

func TestStartWorkers_FailsWhenWorkersAlreadyStarted(t *testing.T) {
	t.Parallel()

	sig := New[string](DefaultConfig(), "test-signal")
	sig.handlers = []SignalListener[string]{
		func(ctx context.Context, s string) error { return nil },
	}
	sig.started = true
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := sig.StartWorkers(ctx)

	require.Error(t, err)
	require.Contains(t, err.Error(), "workers are already started")
}

func TestWorkerProcessing_ProcessesEventsSuccessfully(t *testing.T) {
	var callCount atomic.Int32
	handler := func(ctx context.Context, i int) error {
		callCount.Add(1)
		return nil
	}

	sig := New(DefaultConfig(), "test-signal", handler)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := sig.StartWorkers(ctx)
	require.NoError(t, err)

	// Emit events
	for _, event := range []int{1, 2, 3} {
		err := sig.Emit(event)
		require.NoError(t, err)
	}

	// Wait for processing
	time.Sleep(100 * time.Millisecond)

	require.Equal(t, int32(3), callCount.Load())
}

func TestWorkerProcessing_ExecutesMultipleHandlersSequentially(t *testing.T) {
	var callCount atomic.Int32
	handler1 := func(ctx context.Context, i int) error {
		callCount.Add(1)
		return nil
	}
	handler2 := func(ctx context.Context, i int) error {
		callCount.Add(1)
		return nil
	}

	sig := New(DefaultConfig(), "test-signal", handler1, handler2)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := sig.StartWorkers(ctx)
	require.NoError(t, err)

	// Emit events
	for _, event := range []int{1, 2} {
		err := sig.Emit(event)
		require.NoError(t, err)
	}

	// Wait for processing
	time.Sleep(100 * time.Millisecond)

	require.Equal(t, int32(4), callCount.Load()) // 2 events * 2 handlers
}

func TestWorkerProcessing_StopsOnFirstHandlerError(t *testing.T) {
	var callCount atomic.Int32
	handler1 := func(ctx context.Context, i int) error {
		callCount.Add(1)
		return errors.New("handler error")
	}
	handler2 := func(ctx context.Context, i int) error {
		callCount.Add(1)
		return nil
	}

	sig := New(DefaultConfig(), "test-signal", handler1, handler2)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := sig.StartWorkers(ctx)
	require.NoError(t, err)

	// Emit event
	err = sig.Emit(1)
	require.NoError(t, err)

	// Wait for processing
	time.Sleep(100 * time.Millisecond)

	require.Equal(t, int32(1), callCount.Load()) // Only first handler should be called
}

func TestWorkerShutdown_OnContextCancellation(t *testing.T) {
	processingStarted := make(chan struct{})
	blockHandler := make(chan struct{})

	sig := New(DefaultConfig(), "test-signal", func(ctx context.Context, s string) error {
		close(processingStarted)
		<-blockHandler
		return nil
	})

	ctx, cancel := context.WithCancel(context.Background())

	err := sig.StartWorkers(ctx)
	require.NoError(t, err)

	// Emit an event
	err = sig.Emit("test")
	require.NoError(t, err)

	// Wait for processing to start
	<-processingStarted

	// Cancel context
	cancel()

	// Unblock handler
	close(blockHandler)

	// Wait for shutdown
	time.Sleep(100 * time.Millisecond)

	require.True(t, sig.stopped.Load())
}

func TestWorkerShutdown_MultipleWorkersGracefully(t *testing.T) {
	var wg sync.WaitGroup
	processedCount := atomic.Int32{}

	sig := New(Config{BufferSize: 100, WorkerCount: 5}, "test-signal", func(ctx context.Context, i int) error {
		processedCount.Add(1)
		time.Sleep(10 * time.Millisecond)
		return nil
	})

	ctx, cancel := context.WithCancel(context.Background())

	err := sig.StartWorkers(ctx)
	require.NoError(t, err)

	// Emit multiple events
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 20; i++ {
			_ = sig.Emit(i)
		}
	}()

	wg.Wait()

	// Cancel and wait for shutdown
	cancel()
	time.Sleep(200 * time.Millisecond)

	require.True(t, sig.stopped.Load())
	require.Positive(t, processedCount.Load())
}

func TestConcurrency_ConcurrentEmitsAreSafe(t *testing.T) {
	processedCount := atomic.Int32{}

	sig := New(Config{BufferSize: 100, WorkerCount: 10}, "test-signal", func(ctx context.Context, i int) error {
		processedCount.Add(1)
		return nil
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := sig.StartWorkers(ctx)
	require.NoError(t, err)

	var wg sync.WaitGroup
	emitCount := 100
	goroutines := 10

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < emitCount/goroutines; j++ {
				_ = sig.Emit(j)
			}
		}()
	}

	wg.Wait()

	// Wait for processing
	time.Sleep(500 * time.Millisecond)

	require.Equal(t, int32(emitCount), processedCount.Load())
}
