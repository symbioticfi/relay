// Package broadcaster provides a generic Hub implementation for dynamic fan-out messaging.
//
// The Hub allows publishers to broadcast messages to a dynamic set of subscribers.
// Subscribers can join and leave at any time (0 to N subscribers).
//
// Usage example:
//
//	// Create a hub for string messages with custom buffer size
//	hub := broadcaster.NewHub[string](
//	    broadcaster.WithBufferSize[string](100),
//	)
//
//	// Subscribe to receive messages
//	ch := hub.Subscribe("subscriber-1")
//	defer hub.Unsubscribe("subscriber-1")
//
//	// Receive messages in a goroutine
//	go func() {
//	    for msg := range ch {
//	        fmt.Println("Received:", msg)
//	    }
//	}()
//
//	// Broadcast to all subscribers (non-blocking)
//	hub.Broadcast("Hello, World!")
//
//	// Check subscriber count
//	count := hub.Count()
//	fmt.Printf("Active subscribers: %d\n", count)
//
// Thread Safety:
//
// All Hub methods are thread-safe and can be called concurrently from multiple goroutines.
//
// Backpressure Handling:
//
// The Hub uses buffered channels for each subscriber. If a subscriber's channel buffer is full,
// the Broadcast method will drop the message for that subscriber (non-blocking).
// Use BroadcastWait if you need to ensure all subscribers receive the message (blocking).
package broadcaster
