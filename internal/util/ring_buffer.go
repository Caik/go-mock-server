package util

import (
	"fmt"
	"sync"
)

// RingBuffer is a generic, thread-safe circular buffer with fixed capacity.
// When full, new items overwrite the oldest items.
type RingBuffer[T any] struct {
	data     []T
	capacity int
	size     int
	head     int // index of next write position
	mu       sync.RWMutex
}

// NewRingBuffer creates a new ring buffer with the specified capacity.
// Returns an error if capacity is less than or equal to 0.
func NewRingBuffer[T any](capacity int) (*RingBuffer[T], error) {
	if capacity <= 0 {
		return nil, fmt.Errorf("invalid capacity %d: must be greater than 0", capacity)
	}

	return &RingBuffer[T]{
		data:     make([]T, capacity),
		capacity: capacity,
		size:     0,
		head:     0,
	}, nil
}

// Add adds an item to the buffer. If the buffer is full, the oldest item is overwritten.
func (r *RingBuffer[T]) Add(item T) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.data[r.head] = item
	r.head = (r.head + 1) % r.capacity

	if r.size < r.capacity {
		r.size++
	}
}

// GetAll returns all items in the buffer, ordered from oldest to newest.
func (r *RingBuffer[T]) GetAll() []T {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.size == 0 {
		return []T{}
	}

	result := make([]T, r.size)

	// Calculate start position (oldest item)
	start := 0

	if r.size == r.capacity {
		start = r.head // When full, head points to oldest item
	}

	for i := 0; i < r.size; i++ {
		idx := (start + i) % r.capacity
		result[i] = r.data[idx]
	}

	return result
}

// GetRecent returns up to n most recent items, ordered from oldest to newest.
func (r *RingBuffer[T]) GetRecent(n int) []T {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.size == 0 || n <= 0 {
		return []T{}
	}

	count := n

	if count > r.size {
		count = r.size
	}

	result := make([]T, count)

	// Calculate start position for the last 'count' items
	// head points to next write position, so (head - 1) is the most recent
	// We want items from (head - count) to (head - 1)
	startOffset := r.size - count
	start := 0

	if r.size == r.capacity {
		start = r.head
	}

	for i := 0; i < count; i++ {
		idx := (start + startOffset + i) % r.capacity
		result[i] = r.data[idx]
	}

	return result
}

// Size returns the current number of items in the buffer.
func (r *RingBuffer[T]) Size() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.size
}

// Capacity returns the maximum capacity of the buffer.
func (r *RingBuffer[T]) Capacity() int {
	return r.capacity
}

// Clear removes all items from the buffer.
func (r *RingBuffer[T]) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.data = make([]T, r.capacity)
	r.size = 0
	r.head = 0
}
