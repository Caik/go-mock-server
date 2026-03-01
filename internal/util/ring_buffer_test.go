package util

import (
	"sync"
	"testing"
)

func TestNewRingBuffer(t *testing.T) {
	t.Run("creates buffer with specified capacity", func(t *testing.T) {
		rb, err := NewRingBuffer[int](10)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if rb.Capacity() != 10 {
			t.Errorf("expected capacity 10, got %d", rb.Capacity())
		}

		if rb.Size() != 0 {
			t.Errorf("expected size 0, got %d", rb.Size())
		}
	})

	t.Run("returns error for zero capacity", func(t *testing.T) {
		rb, err := NewRingBuffer[int](0)

		if err == nil {
			t.Error("expected error for zero capacity")
		}

		if rb != nil {
			t.Error("expected nil buffer for zero capacity")
		}
	})

	t.Run("returns error for negative capacity", func(t *testing.T) {
		rb, err := NewRingBuffer[int](-5)

		if err == nil {
			t.Error("expected error for negative capacity")
		}

		if rb != nil {
			t.Error("expected nil buffer for negative capacity")
		}
	})
}

// mustNewRingBuffer creates a ring buffer or fails the test
func mustNewRingBuffer[T any](t *testing.T, capacity int) *RingBuffer[T] {
	t.Helper()
	rb, err := NewRingBuffer[T](capacity)

	if err != nil {
		t.Fatalf("failed to create ring buffer: %v", err)
	}

	return rb
}

func TestRingBuffer_Add(t *testing.T) {
	t.Run("adds items and increases size", func(t *testing.T) {
		rb := mustNewRingBuffer[string](t, 5)

		rb.Add("a")
		rb.Add("b")
		rb.Add("c")

		if rb.Size() != 3 {
			t.Errorf("expected size 3, got %d", rb.Size())
		}
	})

	t.Run("overwrites oldest when full", func(t *testing.T) {
		rb := mustNewRingBuffer[int](t, 3)

		rb.Add(1)
		rb.Add(2)
		rb.Add(3)
		rb.Add(4) // Should overwrite 1

		if rb.Size() != 3 {
			t.Errorf("expected size 3, got %d", rb.Size())
		}

		items := rb.GetAll()
		expected := []int{2, 3, 4}

		for i, v := range expected {
			if items[i] != v {
				t.Errorf("expected items[%d] = %d, got %d", i, v, items[i])
			}
		}
	})
}

func TestRingBuffer_GetAll(t *testing.T) {
	t.Run("returns empty slice for empty buffer", func(t *testing.T) {
		rb := mustNewRingBuffer[int](t, 5)

		items := rb.GetAll()

		if len(items) != 0 {
			t.Errorf("expected empty slice, got %d items", len(items))
		}
	})

	t.Run("returns items in order oldest to newest", func(t *testing.T) {
		rb := mustNewRingBuffer[string](t, 5)

		rb.Add("first")
		rb.Add("second")
		rb.Add("third")

		items := rb.GetAll()

		if len(items) != 3 {
			t.Fatalf("expected 3 items, got %d", len(items))
		}

		if items[0] != "first" || items[1] != "second" || items[2] != "third" {
			t.Errorf("expected [first, second, third], got %v", items)
		}
	})

	t.Run("returns correct order after wrap-around", func(t *testing.T) {
		rb := mustNewRingBuffer[int](t, 3)

		rb.Add(1)
		rb.Add(2)
		rb.Add(3)
		rb.Add(4)
		rb.Add(5)

		items := rb.GetAll()

		expected := []int{3, 4, 5}

		for i, v := range expected {
			if items[i] != v {
				t.Errorf("expected items[%d] = %d, got %d", i, v, items[i])
			}
		}
	})
}

func TestRingBuffer_GetRecent(t *testing.T) {
	t.Run("returns empty slice for empty buffer", func(t *testing.T) {
		rb := mustNewRingBuffer[int](t, 5)

		items := rb.GetRecent(3)

		if len(items) != 0 {
			t.Errorf("expected empty slice, got %d items", len(items))
		}
	})

	t.Run("returns n most recent items", func(t *testing.T) {
		rb := mustNewRingBuffer[int](t, 10)

		for i := 1; i <= 5; i++ {
			rb.Add(i)
		}

		items := rb.GetRecent(3)

		expected := []int{3, 4, 5}

		if len(items) != 3 {
			t.Fatalf("expected 3 items, got %d", len(items))
		}

		for i, v := range expected {
			if items[i] != v {
				t.Errorf("expected items[%d] = %d, got %d", i, v, items[i])
			}
		}
	})

	t.Run("returns all items when n > size", func(t *testing.T) {
		rb := mustNewRingBuffer[int](t, 10)

		rb.Add(1)
		rb.Add(2)

		items := rb.GetRecent(5)

		if len(items) != 2 {
			t.Errorf("expected 2 items, got %d", len(items))
		}
	})

	t.Run("handles zero n", func(t *testing.T) {
		rb := mustNewRingBuffer[int](t, 5)
		rb.Add(1)

		items := rb.GetRecent(0)

		if len(items) != 0 {
			t.Errorf("expected empty slice, got %d items", len(items))
		}
	})
}

func TestRingBuffer_Clear(t *testing.T) {
	t.Run("clears all items", func(t *testing.T) {
		rb := mustNewRingBuffer[int](t, 5)

		rb.Add(1)
		rb.Add(2)
		rb.Add(3)

		rb.Clear()

		if rb.Size() != 0 {
			t.Errorf("expected size 0 after clear, got %d", rb.Size())
		}

		items := rb.GetAll()

		if len(items) != 0 {
			t.Errorf("expected empty slice after clear, got %d items", len(items))
		}
	})

	t.Run("can add items after clear", func(t *testing.T) {
		rb := mustNewRingBuffer[string](t, 3)

		rb.Add("old1")
		rb.Add("old2")
		rb.Clear()
		rb.Add("new1")

		if rb.Size() != 1 {
			t.Errorf("expected size 1, got %d", rb.Size())
		}

		items := rb.GetAll()

		if items[0] != "new1" {
			t.Errorf("expected 'new1', got '%s'", items[0])
		}
	})
}

func TestRingBuffer_Concurrent(t *testing.T) {
	t.Run("handles concurrent access", func(t *testing.T) {
		rb := mustNewRingBuffer[int](t, 100)
		var wg sync.WaitGroup

		// Multiple writers
		for i := 0; i < 10; i++ {
			wg.Add(1)

			go func(start int) {
				defer wg.Done()

				for j := 0; j < 100; j++ {
					rb.Add(start*100 + j)
				}
			}(i)
		}

		// Multiple readers
		for i := 0; i < 5; i++ {
			wg.Add(1)

			go func() {
				defer wg.Done()

				for j := 0; j < 50; j++ {
					_ = rb.GetAll()
					_ = rb.Size()
				}
			}()
		}

		wg.Wait()

		// Should have exactly capacity items
		if rb.Size() != 100 {
			t.Errorf("expected size 100, got %d", rb.Size())
		}
	})
}
