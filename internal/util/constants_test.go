package util

import (
	"testing"
)

// Test constants
func TestConstants(t *testing.T) {
	t.Run("UuidKey constant", func(t *testing.T) {
		expected := "uuid"
		if UuidKey != expected {
			t.Errorf("expected UuidKey to be '%s', got '%s'", expected, UuidKey)
		}
	})

	t.Run("UuidKey is string type", func(t *testing.T) {
		// Verify it can be used as a string
		var key string = UuidKey
		if key != "uuid" {
			t.Errorf("UuidKey should be usable as string, got '%s'", key)
		}
	})

	t.Run("UuidKey consistency", func(t *testing.T) {
		// Test that the constant is consistent across multiple accesses
		key1 := UuidKey
		key2 := UuidKey
		if key1 != key2 {
			t.Error("UuidKey should be consistent across multiple accesses")
		}
	})
}
