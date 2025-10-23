package broadcaster

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHub_SubscribeUnsubscribe(t *testing.T) {
	hub := NewHub[string]()

	// Subscribe
	ch := hub.Subscribe("sub1")
	assert.NotNil(t, ch)
	assert.Equal(t, 1, hub.Count())

	// Subscribe another
	ch2 := hub.Subscribe("sub2")
	assert.NotNil(t, ch2)
	assert.Equal(t, 2, hub.Count())

	// Unsubscribe
	hub.Unsubscribe("sub1")
	assert.Equal(t, 1, hub.Count())

	hub.Unsubscribe("sub2")
	assert.Equal(t, 0, hub.Count())
}

func TestHub_SubscribeIdempotent(t *testing.T) {
	hub := NewHub[string]()

	// Subscribe twice with same ID
	ch1 := hub.Subscribe("sub1")
	ch2 := hub.Subscribe("sub1")

	// Should return same channel
	assert.Equal(t, ch1, ch2)
	assert.Equal(t, 1, hub.Count())
}

func TestHub_BroadcastToAll(t *testing.T) {
	hub := NewHub[string](WithBufferSize[string](10))

	// Subscribe multiple subscribers
	ch1 := hub.Subscribe("sub1")
	ch2 := hub.Subscribe("sub2")
	ch3 := hub.Subscribe("sub3")

	// Broadcast message
	msg := "test message"
	hub.Broadcast(msg)

	// All should receive the message
	assert.Equal(t, msg, <-ch1)
	assert.Equal(t, msg, <-ch2)
	assert.Equal(t, msg, <-ch3)
}

func TestHub_BroadcastWithNoSubscribers(t *testing.T) {
	hub := NewHub[string]()

	// Should not panic
	require.NotPanics(t, func() {
		hub.Broadcast("test")
	})
}

func TestHub_BroadcastDropsWhenBufferFull(t *testing.T) {
	hub := NewHub[string](WithBufferSize[string](1))

	ch := hub.Subscribe("sub1")

	// Fill the buffer
	hub.Broadcast("msg1")
	hub.Broadcast("msg2") // This should be dropped

	// Read one message
	msg := <-ch
	assert.Equal(t, "msg1", msg)

	// Channel should be empty now
	select {
	case <-ch:
		t.Fatal("unexpected message in channel")
	case <-time.After(10 * time.Millisecond):
		// Expected - no more messages
	}
}

func TestHub_BroadcastWaitBlocks(t *testing.T) {
	hub := NewHub[string](WithBufferSize[string](1))

	ch := hub.Subscribe("sub1")

	// Fill the buffer
	hub.Broadcast("msg1")

	// BroadcastWait should block until we read
	done := make(chan bool)
	go func() {
		hub.BroadcastWait("msg2")
		close(done)
	}()

	// Give it time to block
	time.Sleep(50 * time.Millisecond)

	// Should still be blocking
	select {
	case <-done:
		t.Fatal("BroadcastWait should have blocked")
	default:
		// Expected
	}

	// Read from channel to unblock
	<-ch
	<-ch

	// Now should complete
	select {
	case <-done:
		// Expected
	case <-time.After(100 * time.Millisecond):
		t.Fatal("BroadcastWait did not complete")
	}
}

func TestHub_ConcurrentAccess(t *testing.T) {
	hub := NewHub[int](WithBufferSize[int](100))

	var wg sync.WaitGroup
	numSubscribers := 10
	numMessages := 100

	// Start subscribers
	receivers := make([]chan int, numSubscribers)
	for i := 0; i < numSubscribers; i++ {
		receivers[i] = hub.Subscribe(string(rune('A' + i)))
	}

	// Start broadcasting
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < numMessages; i++ {
			hub.Broadcast(i)
		}
	}()

	// Start receiving
	for i := 0; i < numSubscribers; i++ {
		wg.Add(1)
		go func(ch chan int) {
			defer wg.Done()
			count := 0
			timeout := time.After(5 * time.Second)
			for count < numMessages {
				select {
				case <-ch:
					count++
				case <-timeout:
					return
				}
			}
		}(receivers[i])
	}

	wg.Wait()
	assert.Equal(t, numSubscribers, hub.Count())
}

func TestHub_UnsubscribeClosesChannel(t *testing.T) {
	hub := NewHub[string]()

	ch := hub.Subscribe("sub1")
	hub.Unsubscribe("sub1")

	// Channel should be closed
	_, ok := <-ch
	assert.False(t, ok, "channel should be closed")
}

func TestHub_Close(t *testing.T) {
	hub := NewHub[string]()

	ch1 := hub.Subscribe("sub1")
	ch2 := hub.Subscribe("sub2")

	hub.Close()

	// All channels should be closed
	_, ok1 := <-ch1
	_, ok2 := <-ch2
	assert.False(t, ok1)
	assert.False(t, ok2)

	// Count should be 0
	assert.Equal(t, 0, hub.Count())
}

func TestHub_WithOptions(t *testing.T) {
	hub := NewHub[string](
		WithBufferSize[string](50),
	)

	assert.Equal(t, 50, hub.bufferSize)
}

func TestHub_BroadcastAfterUnsubscribe(t *testing.T) {
	hub := NewHub[string](WithBufferSize[string](10))

	ch1 := hub.Subscribe("sub1")
	ch2 := hub.Subscribe("sub2")

	// Unsubscribe one
	hub.Unsubscribe("sub1")

	// Broadcast should only go to remaining subscriber
	hub.Broadcast("test")

	// sub2 should receive
	assert.Equal(t, "test", <-ch2)

	// sub1 should be closed
	_, ok := <-ch1
	assert.False(t, ok)
}
