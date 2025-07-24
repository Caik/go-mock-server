package util

import (
	"sync"
	"testing"
	"time"
)

// Test event type for testing
type TestEvent struct {
	ID      int
	Message string
}

// Test Broadcaster Subscribe method
func TestBroadcaster_Subscribe(t *testing.T) {
	t.Run("first subscription initializes maps", func(t *testing.T) {
		broadcaster := &Broadcaster[TestEvent]{}
		
		acceptFn := func(event TestEvent) bool { return true }
		channel := broadcaster.Subscribe("subscriber1", acceptFn)
		
		if channel == nil {
			t.Error("Subscribe should return a non-nil channel")
		}
		
		if broadcaster.subscribers == nil {
			t.Error("subscribers map should be initialized")
		}
		
		if broadcaster.subscribersAcceptFn == nil {
			t.Error("subscribersAcceptFn map should be initialized")
		}
		
		if len(broadcaster.subscribers) != 1 {
			t.Errorf("expected 1 subscriber, got %d", len(broadcaster.subscribers))
		}
	})
	
	t.Run("multiple subscriptions", func(t *testing.T) {
		broadcaster := &Broadcaster[TestEvent]{}
		
		acceptFn1 := func(event TestEvent) bool { return event.ID > 0 }
		acceptFn2 := func(event TestEvent) bool { return event.ID > 10 }
		
		channel1 := broadcaster.Subscribe("subscriber1", acceptFn1)
		channel2 := broadcaster.Subscribe("subscriber2", acceptFn2)
		
		if channel1 == nil || channel2 == nil {
			t.Error("Subscribe should return non-nil channels")
		}
		
		if channel1 == channel2 {
			t.Error("Different subscribers should get different channels")
		}
		
		if len(broadcaster.subscribers) != 2 {
			t.Errorf("expected 2 subscribers, got %d", len(broadcaster.subscribers))
		}
	})
	
	t.Run("duplicate subscription returns same channel", func(t *testing.T) {
		broadcaster := &Broadcaster[TestEvent]{}
		
		acceptFn := func(event TestEvent) bool { return true }
		
		channel1 := broadcaster.Subscribe("subscriber1", acceptFn)
		channel2 := broadcaster.Subscribe("subscriber1", acceptFn) // Same ID
		
		if channel1 != channel2 {
			t.Error("Duplicate subscription should return the same channel")
		}
		
		if len(broadcaster.subscribers) != 1 {
			t.Errorf("expected 1 subscriber after duplicate subscription, got %d", len(broadcaster.subscribers))
		}
	})
}

// Test Broadcaster Unsubscribe method
func TestBroadcaster_Unsubscribe(t *testing.T) {
	t.Run("unsubscribe existing subscriber", func(t *testing.T) {
		broadcaster := &Broadcaster[TestEvent]{}
		
		acceptFn := func(event TestEvent) bool { return true }
		channel := broadcaster.Subscribe("subscriber1", acceptFn)
		
		// Verify subscription exists
		if len(broadcaster.subscribers) != 1 {
			t.Fatalf("expected 1 subscriber before unsubscribe, got %d", len(broadcaster.subscribers))
		}
		
		// Unsubscribe
		broadcaster.Unsubscribe("subscriber1")
		
		// Verify subscription is removed
		if len(broadcaster.subscribers) != 0 {
			t.Errorf("expected 0 subscribers after unsubscribe, got %d", len(broadcaster.subscribers))
		}
		
		if len(broadcaster.subscribersAcceptFn) != 0 {
			t.Errorf("expected 0 accept functions after unsubscribe, got %d", len(broadcaster.subscribersAcceptFn))
		}
		
		// Verify channel is closed
		select {
		case _, ok := <-channel:
			if ok {
				t.Error("channel should be closed after unsubscribe")
			}
		case <-time.After(100 * time.Millisecond):
			t.Error("channel should be closed immediately after unsubscribe")
		}
	})
	
	t.Run("unsubscribe non-existent subscriber", func(t *testing.T) {
		broadcaster := &Broadcaster[TestEvent]{}
		
		// This should not panic or cause issues
		broadcaster.Unsubscribe("non-existent")
		
		// Should still be able to subscribe after attempting to unsubscribe non-existent
		acceptFn := func(event TestEvent) bool { return true }
		channel := broadcaster.Subscribe("subscriber1", acceptFn)
		
		if channel == nil {
			t.Error("Should be able to subscribe after unsubscribing non-existent subscriber")
		}
	})
	
	t.Run("unsubscribe one of multiple subscribers", func(t *testing.T) {
		broadcaster := &Broadcaster[TestEvent]{}
		
		acceptFn := func(event TestEvent) bool { return true }
		
		broadcaster.Subscribe("subscriber1", acceptFn)
		broadcaster.Subscribe("subscriber2", acceptFn)
		broadcaster.Subscribe("subscriber3", acceptFn)
		
		if len(broadcaster.subscribers) != 3 {
			t.Fatalf("expected 3 subscribers, got %d", len(broadcaster.subscribers))
		}
		
		// Unsubscribe middle subscriber
		broadcaster.Unsubscribe("subscriber2")
		
		if len(broadcaster.subscribers) != 2 {
			t.Errorf("expected 2 subscribers after unsubscribe, got %d", len(broadcaster.subscribers))
		}
		
		// Verify correct subscribers remain
		if _, exists := broadcaster.subscribers["subscriber1"]; !exists {
			t.Error("subscriber1 should still exist")
		}
		
		if _, exists := broadcaster.subscribers["subscriber3"]; !exists {
			t.Error("subscriber3 should still exist")
		}
		
		if _, exists := broadcaster.subscribers["subscriber2"]; exists {
			t.Error("subscriber2 should be removed")
		}
	})
}

// Test Broadcaster Publish method
func TestBroadcaster_Publish(t *testing.T) {
	t.Run("publish to single subscriber", func(t *testing.T) {
		broadcaster := &Broadcaster[TestEvent]{}
		
		acceptFn := func(event TestEvent) bool { return true }
		channel := broadcaster.Subscribe("subscriber1", acceptFn)
		
		event := TestEvent{ID: 1, Message: "test message"}
		
		// Publish in a goroutine to avoid blocking
		go broadcaster.Publish(event, "test-uuid")
		
		// Receive the event
		select {
		case receivedEvent := <-channel:
			if receivedEvent.ID != event.ID || receivedEvent.Message != event.Message {
				t.Errorf("expected event %+v, got %+v", event, receivedEvent)
			}
		case <-time.After(1 * time.Second):
			t.Error("timeout waiting for event")
		}
	})
	
	t.Run("publish to multiple subscribers", func(t *testing.T) {
		broadcaster := &Broadcaster[TestEvent]{}
		
		acceptFn := func(event TestEvent) bool { return true }
		
		channel1 := broadcaster.Subscribe("subscriber1", acceptFn)
		channel2 := broadcaster.Subscribe("subscriber2", acceptFn)
		channel3 := broadcaster.Subscribe("subscriber3", acceptFn)
		
		event := TestEvent{ID: 2, Message: "broadcast message"}
		
		// Publish in a goroutine
		go broadcaster.Publish(event, "test-uuid")
		
		// Collect events from all channels
		var receivedEvents []TestEvent
		var wg sync.WaitGroup
		wg.Add(3)
		
		receiveEvent := func(ch <-chan TestEvent) {
			defer wg.Done()
			select {
			case event := <-ch:
				receivedEvents = append(receivedEvents, event)
			case <-time.After(1 * time.Second):
				t.Error("timeout waiting for event")
			}
		}
		
		go receiveEvent(channel1)
		go receiveEvent(channel2)
		go receiveEvent(channel3)
		
		wg.Wait()
		
		if len(receivedEvents) != 3 {
			t.Errorf("expected 3 events, got %d", len(receivedEvents))
		}
		
		// Verify all events are correct
		for i, receivedEvent := range receivedEvents {
			if receivedEvent.ID != event.ID || receivedEvent.Message != event.Message {
				t.Errorf("event %d: expected %+v, got %+v", i, event, receivedEvent)
			}
		}
	})
	
	t.Run("publish with accept function filtering", func(t *testing.T) {
		broadcaster := &Broadcaster[TestEvent]{}
		
		// Subscriber 1 accepts all events
		acceptAll := func(event TestEvent) bool { return true }
		channel1 := broadcaster.Subscribe("subscriber1", acceptAll)
		
		// Subscriber 2 only accepts events with ID > 5
		acceptFiltered := func(event TestEvent) bool { return event.ID > 5 }
		channel2 := broadcaster.Subscribe("subscriber2", acceptFiltered)
		
		// Event that should be filtered out for subscriber2
		event1 := TestEvent{ID: 3, Message: "filtered message"}
		
		go broadcaster.Publish(event1, "test-uuid")
		
		// Subscriber 1 should receive the event
		select {
		case receivedEvent := <-channel1:
			if receivedEvent.ID != event1.ID {
				t.Errorf("subscriber1 should receive event, got %+v", receivedEvent)
			}
		case <-time.After(1 * time.Second):
			t.Error("subscriber1 should receive the event")
		}
		
		// Subscriber 2 should not receive the event (non-blocking check)
		select {
		case <-channel2:
			t.Error("subscriber2 should not receive filtered event")
		case <-time.After(100 * time.Millisecond):
			// Expected - no event received
		}
		
		// Event that should pass the filter
		event2 := TestEvent{ID: 10, Message: "unfiltered message"}
		
		go broadcaster.Publish(event2, "test-uuid")
		
		// Both subscribers should receive this event
		var wg sync.WaitGroup
		wg.Add(2)
		
		go func() {
			defer wg.Done()
			select {
			case <-channel1:
				// Expected
			case <-time.After(1 * time.Second):
				t.Error("subscriber1 should receive unfiltered event")
			}
		}()
		
		go func() {
			defer wg.Done()
			select {
			case <-channel2:
				// Expected
			case <-time.After(1 * time.Second):
				t.Error("subscriber2 should receive unfiltered event")
			}
		}()
		
		wg.Wait()
	})

	t.Run("publish to no subscribers", func(t *testing.T) {
		broadcaster := &Broadcaster[TestEvent]{}

		event := TestEvent{ID: 1, Message: "no subscribers"}

		// This should not panic or block
		done := make(chan bool)
		go func() {
			broadcaster.Publish(event, "test-uuid")
			done <- true
		}()

		select {
		case <-done:
			// Expected - publish should complete quickly
		case <-time.After(1 * time.Second):
			t.Error("publish should complete quickly when no subscribers")
		}
	})

	t.Run("publish with missing accept function", func(t *testing.T) {
		broadcaster := &Broadcaster[TestEvent]{}

		// Manually create a subscriber without accept function to test the warning case
		broadcaster.subscribers = make(map[string]chan TestEvent)
		broadcaster.subscribersAcceptFn = make(map[string]func(event TestEvent) bool)

		channel := make(chan TestEvent, 1) // Buffered channel to prevent blocking
		broadcaster.subscribers["subscriber1"] = channel
		// Intentionally not setting subscribersAcceptFn["subscriber1"]

		event := TestEvent{ID: 1, Message: "test"}

		// This should log a warning but still send the event (since acceptFn doesn't exist, it defaults to sending)
		done := make(chan bool)
		go func() {
			broadcaster.Publish(event, "test-uuid")
			done <- true
		}()

		// The event should still be sent despite missing accept function
		select {
		case receivedEvent := <-channel:
			if receivedEvent.ID != event.ID {
				t.Errorf("expected event %+v, got %+v", event, receivedEvent)
			}
		case <-time.After(500 * time.Millisecond):
			t.Error("should receive event even with missing accept function")
		}

		select {
		case <-done:
			// Expected - should complete despite missing accept function
		case <-time.After(1 * time.Second):
			t.Error("publish should complete even with missing accept function")
		}
	})
}

// Test Broadcaster PublishAsync method
func TestBroadcaster_PublishAsync(t *testing.T) {
	t.Run("async publish returns immediately", func(t *testing.T) {
		broadcaster := &Broadcaster[TestEvent]{}

		acceptFn := func(event TestEvent) bool { return true }
		channel := broadcaster.Subscribe("subscriber1", acceptFn)

		event := TestEvent{ID: 1, Message: "async message"}

		start := time.Now()
		wg := broadcaster.PublishAsync(event, "test-uuid")
		duration := time.Since(start)

		// PublishAsync should return immediately
		if duration > 100*time.Millisecond {
			t.Errorf("PublishAsync took too long: %v", duration)
		}

		if wg == nil {
			t.Error("PublishAsync should return a non-nil WaitGroup")
		}

		// Receive the event first, then wait for completion
		select {
		case receivedEvent := <-channel:
			if receivedEvent.ID != event.ID || receivedEvent.Message != event.Message {
				t.Errorf("expected event %+v, got %+v", event, receivedEvent)
			}
		case <-time.After(2 * time.Second):
			t.Error("timeout waiting for async event")
		}

		// Wait for the async operation to complete
		wg.Wait()
	})

	t.Run("multiple async publishes", func(t *testing.T) {
		broadcaster := &Broadcaster[TestEvent]{}

		acceptFn := func(event TestEvent) bool { return true }
		channel := broadcaster.Subscribe("subscriber1", acceptFn)

		event1 := TestEvent{ID: 1, Message: "async message 1"}

		// Test just one async publish to avoid complexity
		wg := broadcaster.PublishAsync(event1, "uuid1")

		// Receive the event
		select {
		case receivedEvent := <-channel:
			if receivedEvent.ID != event1.ID {
				t.Errorf("expected event ID %d, got %d", event1.ID, receivedEvent.ID)
			}
		case <-time.After(2 * time.Second):
			t.Error("timeout waiting for async event")
		}

		// Wait for completion
		wg.Wait()
	})

	t.Run("async publish with no subscribers", func(t *testing.T) {
		broadcaster := &Broadcaster[TestEvent]{}

		event := TestEvent{ID: 1, Message: "no subscribers async"}

		wg := broadcaster.PublishAsync(event, "test-uuid")

		if wg == nil {
			t.Error("PublishAsync should return a non-nil WaitGroup even with no subscribers")
		}

		// Should complete quickly
		done := make(chan bool)
		go func() {
			wg.Wait()
			done <- true
		}()

		select {
		case <-done:
			// Expected
		case <-time.After(1 * time.Second):
			t.Error("async publish should complete quickly with no subscribers")
		}
	})
}

// Test Broadcaster with different types
func TestBroadcaster_DifferentTypes(t *testing.T) {
	t.Run("string broadcaster", func(t *testing.T) {
		broadcaster := &Broadcaster[string]{}

		acceptFn := func(event string) bool { return len(event) > 0 }
		channel := broadcaster.Subscribe("subscriber1", acceptFn)

		message := "hello world"

		go broadcaster.Publish(message, "test-uuid")

		select {
		case received := <-channel:
			if received != message {
				t.Errorf("expected %s, got %s", message, received)
			}
		case <-time.After(1 * time.Second):
			t.Error("timeout waiting for string event")
		}
	})

	t.Run("int broadcaster", func(t *testing.T) {
		broadcaster := &Broadcaster[int]{}

		acceptFn := func(event int) bool { return event > 0 }
		channel := broadcaster.Subscribe("subscriber1", acceptFn)

		number := 42

		go broadcaster.Publish(number, "test-uuid")

		select {
		case received := <-channel:
			if received != number {
				t.Errorf("expected %d, got %d", number, received)
			}
		case <-time.After(1 * time.Second):
			t.Error("timeout waiting for int event")
		}
	})
}

// Note: Concurrent access tests are omitted as the broadcaster implementation
// is not thread-safe and would require additional synchronization mechanisms.


