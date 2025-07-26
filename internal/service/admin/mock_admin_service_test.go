package admin

import (
	"errors"
	"testing"

	"github.com/Caik/go-mock-server/internal/service/content"
)

// Mock content service for testing
type mockContentService struct {
	contents    map[string][]byte
	events      chan content.ContentEvent
	shouldError bool
	errorMsg    string
}

func (m *mockContentService) GetContent(host, uri, method, uuid string) (*[]byte, error) {
	if m.shouldError {
		return nil, errors.New(m.errorMsg)
	}

	key := host + ":" + uri + ":" + method
	if data, exists := m.contents[key]; exists {
		return &data, nil
	}
	return nil, errors.New("not found")
}

func (m *mockContentService) SetContent(host, uri, method, uuid string, data *[]byte) error {
	if m.shouldError {
		return errors.New(m.errorMsg)
	}

	key := host + ":" + uri + ":" + method
	if data != nil {
		m.contents[key] = *data
	} else {
		m.contents[key] = nil
	}
	return nil
}

func (m *mockContentService) DeleteContent(host, uri, method, uuid string) error {
	if m.shouldError {
		return errors.New(m.errorMsg)
	}

	key := host + ":" + uri + ":" + method
	delete(m.contents, key)
	return nil
}

func (m *mockContentService) ListContents(uuid string) (*[]content.ContentData, error) {
	if m.shouldError {
		return nil, errors.New(m.errorMsg)
	}

	var contents []content.ContentData
	for range m.contents {
		contents = append(contents, content.ContentData{
			Host:   "example.com",
			Uri:    "/api/test",
			Method: "GET",
		})
	}
	return &contents, nil
}

func (m *mockContentService) Subscribe(subscriberId string, eventTypes ...content.ContentEventType) <-chan content.ContentEvent {
	return m.events
}

func (m *mockContentService) Unsubscribe(subscriberId string) {
	// Mock implementation
}

func TestNewMockAdminService(t *testing.T) {
	t.Run("creates service with content service", func(t *testing.T) {
		contentService := &mockContentService{
			contents: make(map[string][]byte),
			events:   make(chan content.ContentEvent),
		}

		service := NewMockAdminService(contentService)

		if service == nil {
			t.Fatal("NewMockAdminService should return non-nil service")
		}

		if service.contentService != contentService {
			t.Error("service should store the provided content service")
		}
	})

	t.Run("handles nil content service", func(t *testing.T) {
		service := NewMockAdminService(nil)

		if service == nil {
			t.Fatal("NewMockAdminService should return non-nil service even with nil content service")
		}

		if service.contentService != nil {
			t.Error("service should store nil content service")
		}
	})
}

func TestMockAdminService_AddUpdateMock(t *testing.T) {
	t.Run("adds mock successfully", func(t *testing.T) {
		contentService := &mockContentService{
			contents: make(map[string][]byte),
			events:   make(chan content.ContentEvent),
		}

		service := NewMockAdminService(contentService)

		testData := []byte(`{"message": "test response"}`)
		request := MockAddDeleteRequest{
			Host:   "example.com",
			URI:    "/api/users",
			Method: "GET",
			Data:   &testData,
		}

		err := service.AddUpdateMock(request, "test-uuid")

		if err != nil {
			t.Fatalf("AddUpdateMock should not return error: %v", err)
		}

		// Verify mock was stored in content service
		key := "example.com:/api/users:GET"
		storedData, exists := contentService.contents[key]
		if !exists {
			t.Error("mock should be stored in content service")
		}

		if string(storedData) != string(testData) {
			t.Errorf("expected stored data '%s', got '%s'", string(testData), string(storedData))
		}
	})

	t.Run("updates existing mock successfully", func(t *testing.T) {
		contentService := &mockContentService{
			contents: map[string][]byte{
				"example.com:/api/users:GET": []byte(`{"message": "old response"}`),
			},
			events: make(chan content.ContentEvent),
		}

		service := NewMockAdminService(contentService)

		newTestData := []byte(`{"message": "new response"}`)
		request := MockAddDeleteRequest{
			Host:   "example.com",
			URI:    "/api/users",
			Method: "GET",
			Data:   &newTestData,
		}

		err := service.AddUpdateMock(request, "test-uuid")

		if err != nil {
			t.Fatalf("AddUpdateMock should not return error: %v", err)
		}

		// Verify mock was updated in content service
		key := "example.com:/api/users:GET"
		storedData, exists := contentService.contents[key]
		if !exists {
			t.Error("mock should exist in content service")
		}

		if string(storedData) != string(newTestData) {
			t.Errorf("expected updated data '%s', got '%s'", string(newTestData), string(storedData))
		}
	})

	t.Run("handles nil data", func(t *testing.T) {
		contentService := &mockContentService{
			contents: make(map[string][]byte),
			events:   make(chan content.ContentEvent),
		}

		service := NewMockAdminService(contentService)

		request := MockAddDeleteRequest{
			Host:   "example.com",
			URI:    "/api/users",
			Method: "GET",
			Data:   nil, // Nil data
		}

		err := service.AddUpdateMock(request, "test-uuid")

		if err != nil {
			t.Fatalf("AddUpdateMock should not return error for nil data: %v", err)
		}

		// Verify nil data was stored
		key := "example.com:/api/users:GET"
		_, exists := contentService.contents[key]
		if !exists {
			t.Error("mock with nil data should be stored in content service")
		}
	})

	t.Run("handles empty data", func(t *testing.T) {
		contentService := &mockContentService{
			contents: make(map[string][]byte),
			events:   make(chan content.ContentEvent),
		}

		service := NewMockAdminService(contentService)

		emptyData := []byte{}
		request := MockAddDeleteRequest{
			Host:   "example.com",
			URI:    "/api/users",
			Method: "GET",
			Data:   &emptyData,
		}

		err := service.AddUpdateMock(request, "test-uuid")

		if err != nil {
			t.Fatalf("AddUpdateMock should not return error for empty data: %v", err)
		}

		// Verify empty data was stored
		key := "example.com:/api/users:GET"
		storedData, exists := contentService.contents[key]
		if !exists {
			t.Error("mock with empty data should be stored in content service")
		}

		if len(storedData) != 0 {
			t.Errorf("expected empty data, got %d bytes", len(storedData))
		}
	})

	t.Run("returns error when content service fails", func(t *testing.T) {
		contentService := &mockContentService{
			contents:    make(map[string][]byte),
			events:      make(chan content.ContentEvent),
			shouldError: true,
			errorMsg:    "content service error",
		}

		service := NewMockAdminService(contentService)

		testData := []byte(`{"message": "test response"}`)
		request := MockAddDeleteRequest{
			Host:   "example.com",
			URI:    "/api/users",
			Method: "GET",
			Data:   &testData,
		}

		err := service.AddUpdateMock(request, "test-uuid")

		if err == nil {
			t.Error("AddUpdateMock should return error when content service fails")
		}

		if err.Error() != "content service error" {
			t.Errorf("expected error message 'content service error', got '%s'", err.Error())
		}
	})

	t.Run("handles different HTTP methods", func(t *testing.T) {
		contentService := &mockContentService{
			contents: make(map[string][]byte),
			events:   make(chan content.ContentEvent),
		}

		service := NewMockAdminService(contentService)

		methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH"}

		for _, method := range methods {
			testData := []byte(`{"method": "` + method + `"}`)
			request := MockAddDeleteRequest{
				Host:   "example.com",
				URI:    "/api/test",
				Method: method,
				Data:   &testData,
			}

			err := service.AddUpdateMock(request, "test-uuid")

			if err != nil {
				t.Fatalf("AddUpdateMock should not return error for method %s: %v", method, err)
			}

			// Verify mock was stored for each method
			key := "example.com:/api/test:" + method
			storedData, exists := contentService.contents[key]
			if !exists {
				t.Errorf("mock should be stored for method %s", method)
			}

			expectedData := `{"method": "` + method + `"}`
			if string(storedData) != expectedData {
				t.Errorf("for method %s, expected data '%s', got '%s'", method, expectedData, string(storedData))
			}
		}
	})
}

func TestMockAdminService_DeleteMock(t *testing.T) {
	t.Run("deletes existing mock successfully", func(t *testing.T) {
		contentService := &mockContentService{
			contents: map[string][]byte{
				"example.com:/api/users:GET": []byte(`{"message": "test response"}`),
			},
			events: make(chan content.ContentEvent),
		}

		service := NewMockAdminService(contentService)

		request := MockAddDeleteRequest{
			Host:   "example.com",
			URI:    "/api/users",
			Method: "GET",
		}

		err := service.DeleteMock(request, "test-uuid")

		if err != nil {
			t.Fatalf("DeleteMock should not return error: %v", err)
		}

		// Verify mock was deleted from content service
		key := "example.com:/api/users:GET"
		_, exists := contentService.contents[key]
		if exists {
			t.Error("mock should be deleted from content service")
		}
	})

	t.Run("handles deletion of non-existent mock", func(t *testing.T) {
		contentService := &mockContentService{
			contents: make(map[string][]byte),
			events:   make(chan content.ContentEvent),
		}

		service := NewMockAdminService(contentService)

		request := MockAddDeleteRequest{
			Host:   "example.com",
			URI:    "/api/nonexistent",
			Method: "GET",
		}

		err := service.DeleteMock(request, "test-uuid")

		if err != nil {
			t.Fatalf("DeleteMock should not return error for non-existent mock: %v", err)
		}

		// Verify content service is still empty
		if len(contentService.contents) != 0 {
			t.Error("content service should remain empty")
		}
	})

	t.Run("returns error when content service fails", func(t *testing.T) {
		contentService := &mockContentService{
			contents:    make(map[string][]byte),
			events:      make(chan content.ContentEvent),
			shouldError: true,
			errorMsg:    "delete error",
		}

		service := NewMockAdminService(contentService)

		request := MockAddDeleteRequest{
			Host:   "example.com",
			URI:    "/api/users",
			Method: "GET",
		}

		err := service.DeleteMock(request, "test-uuid")

		if err == nil {
			t.Error("DeleteMock should return error when content service fails")
		}

		if err.Error() != "delete error" {
			t.Errorf("expected error message 'delete error', got '%s'", err.Error())
		}
	})

	t.Run("deletes multiple mocks independently", func(t *testing.T) {
		contentService := &mockContentService{
			contents: map[string][]byte{
				"example.com:/api/users:GET":    []byte(`{"users": []}`),
				"example.com:/api/users:POST":   []byte(`{"created": true}`),
				"example.com:/api/products:GET": []byte(`{"products": []}`),
			},
			events: make(chan content.ContentEvent),
		}

		service := NewMockAdminService(contentService)

		// Delete one mock
		request := MockAddDeleteRequest{
			Host:   "example.com",
			URI:    "/api/users",
			Method: "GET",
		}

		err := service.DeleteMock(request, "test-uuid")

		if err != nil {
			t.Fatalf("DeleteMock should not return error: %v", err)
		}

		// Verify only the specified mock was deleted
		if len(contentService.contents) != 2 {
			t.Errorf("expected 2 remaining mocks, got %d", len(contentService.contents))
		}

		// Verify specific mock was deleted
		_, exists := contentService.contents["example.com:/api/users:GET"]
		if exists {
			t.Error("GET /api/users mock should be deleted")
		}

		// Verify other mocks still exist
		_, exists = contentService.contents["example.com:/api/users:POST"]
		if !exists {
			t.Error("POST /api/users mock should still exist")
		}

		_, exists = contentService.contents["example.com:/api/products:GET"]
		if !exists {
			t.Error("GET /api/products mock should still exist")
		}
	})
}
