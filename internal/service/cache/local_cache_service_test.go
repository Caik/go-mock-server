package cache

import (
	"testing"
)

func TestNewInMemoryCacheService(t *testing.T) {
	t.Run("creates new cache service with initialized map", func(t *testing.T) {
		service := NewInMemoryCacheService()

		if service == nil {
			t.Fatal("NewInMemoryCacheService should return non-nil service")
		}

		if service.cache == nil {
			t.Error("cache map should be initialized")
		}

		if len(service.cache) != 0 {
			t.Error("cache should be empty initially")
		}
	})
}

func TestInMemoryCacheService_Set(t *testing.T) {
	t.Run("stores data in cache", func(t *testing.T) {
		service := NewInMemoryCacheService()
		testData := []byte("test data")
		cacheKey := "test-key"
		uuid := "test-uuid"

		service.Set(cacheKey, &testData, uuid)

		// Verify data was stored
		if len(service.cache) != 1 {
			t.Errorf("expected cache to have 1 item, got %d", len(service.cache))
		}

		storedData, exists := service.cache[cacheKey]
		if !exists {
			t.Error("data should be stored in cache")
		}

		if storedData == nil {
			t.Error("stored data should not be nil")
		}

		if string(*storedData) != string(testData) {
			t.Errorf("expected stored data '%s', got '%s'", string(testData), string(*storedData))
		}
	})

	t.Run("initializes cache map if nil", func(t *testing.T) {
		service := &InMemoryCacheService{
			cache: nil, // Simulate uninitialized cache
		}

		testData := []byte("test data")
		cacheKey := "test-key"
		uuid := "test-uuid"

		service.Set(cacheKey, &testData, uuid)

		// Verify cache was initialized
		if service.cache == nil {
			t.Error("cache should be initialized after Set call")
		}

		// Verify data was stored
		storedData, exists := service.cache[cacheKey]
		if !exists {
			t.Error("data should be stored in cache")
		}

		if string(*storedData) != string(testData) {
			t.Errorf("expected stored data '%s', got '%s'", string(testData), string(*storedData))
		}
	})

	t.Run("overwrites existing data", func(t *testing.T) {
		service := NewInMemoryCacheService()
		cacheKey := "test-key"
		uuid := "test-uuid"

		// Store initial data
		initialData := []byte("initial data")
		service.Set(cacheKey, &initialData, uuid)

		// Overwrite with new data
		newData := []byte("new data")
		service.Set(cacheKey, &newData, uuid)

		// Verify new data overwrote old data
		storedData, exists := service.cache[cacheKey]
		if !exists {
			t.Error("data should exist in cache")
		}

		if string(*storedData) != string(newData) {
			t.Errorf("expected stored data '%s', got '%s'", string(newData), string(*storedData))
		}

		// Verify cache still has only one item
		if len(service.cache) != 1 {
			t.Errorf("expected cache to have 1 item, got %d", len(service.cache))
		}
	})

	t.Run("handles multiple cache entries", func(t *testing.T) {
		service := NewInMemoryCacheService()
		uuid := "test-uuid"

		// Store multiple entries
		entries := map[string][]byte{
			"key1": []byte("data1"),
			"key2": []byte("data2"),
			"key3": []byte("data3"),
		}

		for key, data := range entries {
			dataCopy := data // Create a copy to avoid pointer issues
			service.Set(key, &dataCopy, uuid)
		}

		// Verify all entries are stored
		if len(service.cache) != len(entries) {
			t.Errorf("expected cache to have %d items, got %d", len(entries), len(service.cache))
		}

		for key, expectedData := range entries {
			storedData, exists := service.cache[key]
			if !exists {
				t.Errorf("key '%s' should exist in cache", key)
				continue
			}

			if string(*storedData) != string(expectedData) {
				t.Errorf("for key '%s', expected data '%s', got '%s'",
					key, string(expectedData), string(*storedData))
			}
		}
	})

	t.Run("handles nil data pointer", func(t *testing.T) {
		service := NewInMemoryCacheService()
		cacheKey := "test-key"
		uuid := "test-uuid"

		service.Set(cacheKey, nil, uuid)

		// Verify nil was stored
		storedData, exists := service.cache[cacheKey]
		if !exists {
			t.Error("nil data should be stored in cache")
		}

		if storedData != nil {
			t.Error("stored data should be nil")
		}
	})

	t.Run("handles empty data", func(t *testing.T) {
		service := NewInMemoryCacheService()
		cacheKey := "test-key"
		uuid := "test-uuid"
		emptyData := []byte{}

		service.Set(cacheKey, &emptyData, uuid)

		// Verify empty data was stored
		storedData, exists := service.cache[cacheKey]
		if !exists {
			t.Error("empty data should be stored in cache")
		}

		if storedData == nil {
			t.Error("stored data pointer should not be nil")
		}

		if len(*storedData) != 0 {
			t.Errorf("expected empty data, got %d bytes", len(*storedData))
		}
	})
}

func TestInMemoryCacheService_Get(t *testing.T) {
	t.Run("retrieves existing data", func(t *testing.T) {
		service := NewInMemoryCacheService()
		testData := []byte("test data")
		cacheKey := "test-key"
		uuid := "test-uuid"

		// Store data first
		service.Set(cacheKey, &testData, uuid)

		// Retrieve data
		retrievedData, exists := service.Get(cacheKey, uuid)

		if !exists {
			t.Error("data should exist in cache")
		}

		if retrievedData == nil {
			t.Error("retrieved data should not be nil")
		}

		if string(*retrievedData) != string(testData) {
			t.Errorf("expected retrieved data '%s', got '%s'", string(testData), string(*retrievedData))
		}
	})

	t.Run("returns false for non-existent key", func(t *testing.T) {
		service := NewInMemoryCacheService()
		uuid := "test-uuid"

		retrievedData, exists := service.Get("non-existent-key", uuid)

		if exists {
			t.Error("non-existent key should return false")
		}

		if retrievedData != nil {
			t.Error("retrieved data should be nil for non-existent key")
		}
	})

	t.Run("handles empty cache", func(t *testing.T) {
		service := NewInMemoryCacheService()
		uuid := "test-uuid"

		retrievedData, exists := service.Get("any-key", uuid)

		if exists {
			t.Error("empty cache should return false for any key")
		}

		if retrievedData != nil {
			t.Error("retrieved data should be nil for empty cache")
		}
	})

	t.Run("retrieves nil data correctly", func(t *testing.T) {
		service := NewInMemoryCacheService()
		cacheKey := "test-key"
		uuid := "test-uuid"

		// Store nil data
		service.Set(cacheKey, nil, uuid)

		// Retrieve nil data
		retrievedData, exists := service.Get(cacheKey, uuid)

		if !exists {
			t.Error("nil data should exist in cache")
		}

		if retrievedData != nil {
			t.Error("retrieved data should be nil")
		}
	})

	t.Run("retrieves empty data correctly", func(t *testing.T) {
		service := NewInMemoryCacheService()
		cacheKey := "test-key"
		uuid := "test-uuid"
		emptyData := []byte{}

		// Store empty data
		service.Set(cacheKey, &emptyData, uuid)

		// Retrieve empty data
		retrievedData, exists := service.Get(cacheKey, uuid)

		if !exists {
			t.Error("empty data should exist in cache")
		}

		if retrievedData == nil {
			t.Error("retrieved data pointer should not be nil")
		}

		if len(*retrievedData) != 0 {
			t.Errorf("expected empty data, got %d bytes", len(*retrievedData))
		}
	})

	t.Run("handles multiple concurrent gets", func(t *testing.T) {
		service := NewInMemoryCacheService()
		uuid := "test-uuid"

		// Store multiple entries
		entries := map[string][]byte{
			"key1": []byte("data1"),
			"key2": []byte("data2"),
			"key3": []byte("data3"),
		}

		for key, data := range entries {
			dataCopy := data // Create a copy to avoid pointer issues
			service.Set(key, &dataCopy, uuid)
		}

		// Retrieve all entries
		for key, expectedData := range entries {
			retrievedData, exists := service.Get(key, uuid)

			if !exists {
				t.Errorf("key '%s' should exist in cache", key)
				continue
			}

			if retrievedData == nil {
				t.Errorf("retrieved data for key '%s' should not be nil", key)
				continue
			}

			if string(*retrievedData) != string(expectedData) {
				t.Errorf("for key '%s', expected data '%s', got '%s'",
					key, string(expectedData), string(*retrievedData))
			}
		}
	})
}
