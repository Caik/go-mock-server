package admin

import (
	"errors"
	"strconv"
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

func (m *mockContentService) GetContent(host, uri, method, uuid string, statusCode int) (*content.ContentResult, error) {
	if m.shouldError {
		return nil, errors.New(m.errorMsg)
	}

	key := host + ":" + uri + ":" + method + ":" + strconv.Itoa(statusCode)
	if data, exists := m.contents[key]; exists {
		return &content.ContentResult{
			Data:   &data,
			Source: "mock",
			Path:   "/mock/" + key,
		}, nil
	}
	return nil, errors.New("not found")
}

func (m *mockContentService) SetContent(host, uri, method, uuid string, statusCode int, data *[]byte) error {
	if m.shouldError {
		return errors.New(m.errorMsg)
	}

	key := host + ":" + uri + ":" + method + ":" + strconv.Itoa(statusCode)
	if data != nil {
		m.contents[key] = *data
	} else {
		m.contents[key] = nil
	}
	return nil
}

func (m *mockContentService) DeleteContent(host, uri, method, uuid string, statusCode int) error {
	if m.shouldError {
		return errors.New(m.errorMsg)
	}

	key := host + ":" + uri + ":" + method + ":" + strconv.Itoa(statusCode)
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
			Host:       "example.com",
			Uri:        "/api/test",
			Method:     "GET",
			StatusCode: 200,
		})
	}
	return &contents, nil
}

func (m *mockContentService) ListDefaultContents(uuid string) (*[]content.ContentData, error) {
	if m.shouldError {
		return nil, errors.New(m.errorMsg)
	}
	return &[]content.ContentData{}, nil
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
			Host:       "example.com",
			URI:        "/api/users",
			Method:     "GET",
			StatusCode: 200,
			Data:       &testData,
		}

		err := service.AddUpdateMock(request, "test-uuid")

		if err != nil {
			t.Fatalf("AddUpdateMock should not return error: %v", err)
		}

		// Verify mock was stored in content service
		key := "example.com:/api/users:GET:200"
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
				"example.com:/api/users:GET:200": []byte(`{"message": "old response"}`),
			},
			events: make(chan content.ContentEvent),
		}

		service := NewMockAdminService(contentService)

		newTestData := []byte(`{"message": "new response"}`)
		request := MockAddDeleteRequest{
			Host:       "example.com",
			URI:        "/api/users",
			Method:     "GET",
			StatusCode: 200,
			Data:       &newTestData,
		}

		err := service.AddUpdateMock(request, "test-uuid")

		if err != nil {
			t.Fatalf("AddUpdateMock should not return error: %v", err)
		}

		// Verify mock was updated in content service
		key := "example.com:/api/users:GET:200"
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
			Host:       "example.com",
			URI:        "/api/users",
			Method:     "GET",
			StatusCode: 200,
			Data:       nil, // Nil data
		}

		err := service.AddUpdateMock(request, "test-uuid")

		if err != nil {
			t.Fatalf("AddUpdateMock should not return error for nil data: %v", err)
		}

		// Verify nil data was stored
		key := "example.com:/api/users:GET:200"
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
			Host:       "example.com",
			URI:        "/api/users",
			Method:     "GET",
			StatusCode: 200,
			Data:       &emptyData,
		}

		err := service.AddUpdateMock(request, "test-uuid")

		if err != nil {
			t.Fatalf("AddUpdateMock should not return error for empty data: %v", err)
		}

		// Verify empty data was stored
		key := "example.com:/api/users:GET:200"
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
			Host:       "example.com",
			URI:        "/api/users",
			Method:     "GET",
			StatusCode: 200,
			Data:       &testData,
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
				Host:       "example.com",
				URI:        "/api/test",
				Method:     method,
				StatusCode: 200,
				Data:       &testData,
			}

			err := service.AddUpdateMock(request, "test-uuid")

			if err != nil {
				t.Fatalf("AddUpdateMock should not return error for method %s: %v", method, err)
			}

			// Verify mock was stored for each method
			key := "example.com:/api/test:" + method + ":200"
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
				"example.com:/api/users:GET:200": []byte(`{"message": "test response"}`),
			},
			events: make(chan content.ContentEvent),
		}

		service := NewMockAdminService(contentService)

		request := MockAddDeleteRequest{
			Host:       "example.com",
			URI:        "/api/users",
			Method:     "GET",
			StatusCode: 200,
		}

		err := service.DeleteMock(request, "test-uuid")

		if err != nil {
			t.Fatalf("DeleteMock should not return error: %v", err)
		}

		// Verify mock was deleted from content service
		key := "example.com:/api/users:GET:200"
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
			Host:       "example.com",
			URI:        "/api/nonexistent",
			Method:     "GET",
			StatusCode: 200,
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
			Host:       "example.com",
			URI:        "/api/users",
			Method:     "GET",
			StatusCode: 200,
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
				"example.com:/api/users:GET:200":    []byte(`{"users": []}`),
				"example.com:/api/users:POST:200":   []byte(`{"created": true}`),
				"example.com:/api/products:GET:200": []byte(`{"products": []}`),
			},
			events: make(chan content.ContentEvent),
		}

		service := NewMockAdminService(contentService)

		// Delete one mock
		request := MockAddDeleteRequest{
			Host:       "example.com",
			URI:        "/api/users",
			Method:     "GET",
			StatusCode: 200,
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
		_, exists := contentService.contents["example.com:/api/users:GET:200"]
		if exists {
			t.Error("GET /api/users mock should be deleted")
		}

		// Verify other mocks still exist
		_, exists = contentService.contents["example.com:/api/users:POST:200"]
		if !exists {
			t.Error("POST /api/users mock should still exist")
		}

		_, exists = contentService.contents["example.com:/api/products:GET:200"]
		if !exists {
			t.Error("GET /api/products mock should still exist")
		}
	})
}

func TestMockAdminService_ListMocks(t *testing.T) {
	t.Run("lists mocks successfully", func(t *testing.T) {
		contentService := &mockContentService{
			contents: map[string][]byte{
				"example.com:/api/users:GET:200":  []byte(`{}`),
				"example.com:/api/users:POST:200": []byte(`{}`),
			},
			events: make(chan content.ContentEvent),
		}

		service := NewMockAdminService(contentService)
		mocks, err := service.ListMocks("test-uuid")

		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		if len(mocks) != 2 {
			t.Errorf("expected 2 mocks, got %d", len(mocks))
		}
	})

	t.Run("returns empty list when no mocks exist", func(t *testing.T) {
		contentService := &mockContentService{
			contents: make(map[string][]byte),
			events:   make(chan content.ContentEvent),
		}

		service := NewMockAdminService(contentService)
		mocks, err := service.ListMocks("test-uuid")

		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		if len(mocks) != 0 {
			t.Errorf("expected 0 mocks, got %d", len(mocks))
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
		mocks, err := service.ListMocks("test-uuid")

		if err == nil {
			t.Error("expected error, got nil")
		}

		if mocks != nil {
			t.Error("expected nil mocks on error")
		}
	})

	t.Run("returns correct mock data", func(t *testing.T) {
		contentService := &mockContentService{
			contents: map[string][]byte{
				"api.example.com:/users:GET:200": []byte(`{}`),
			},
			events: make(chan content.ContentEvent),
		}

		service := NewMockAdminService(contentService)
		mocks, err := service.ListMocks("test-uuid")

		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		if len(mocks) != 1 {
			t.Fatalf("expected 1 mock, got %d", len(mocks))
		}

		// The mockContentService returns fixed values
		mock := mocks[0]
		// ID should be a non-empty base64-encoded string
		if mock.ID == "" {
			t.Errorf("expected non-empty id")
		}

		// Verify it can be decoded back to host|uri|method|statusCode
		host, uri, method, statusCode, err := decodeMockID(mock.ID)

		if err != nil {
			t.Errorf("failed to decode mock ID: %v", err)
		}

		if host != "example.com" || uri != "/api/test" || method != "GET" || statusCode != 200 {
			t.Errorf("decoded values don't match: host=%s, uri=%s, method=%s, statusCode=%d", host, uri, method, statusCode)
		}

		if mock.Host != "example.com" {
			t.Errorf("expected host 'example.com', got '%s'", mock.Host)
		}

		if mock.URI != "/api/test" {
			t.Errorf("expected uri '/api/test', got '%s'", mock.URI)
		}

		if mock.Method != "GET" {
			t.Errorf("expected method 'GET', got '%s'", mock.Method)
		}

		if mock.StatusCode != 200 {
			t.Errorf("expected status code 200, got %d", mock.StatusCode)
		}
	})
}

func TestGenerateMockID(t *testing.T) {
	t.Run("generates valid base64 encoded ID", func(t *testing.T) {
		id := generateMockID("example.com", "/api/users", "GET", 200)
		if id == "" {
			t.Error("expected non-empty id")
		}

		// Verify it can be decoded
		host, uri, method, statusCode, err := decodeMockID(id)
		if err != nil {
			t.Errorf("failed to decode: %v", err)
		}

		if host != "example.com" || uri != "/api/users" || method != "GET" || statusCode != 200 {
			t.Errorf("decoded values don't match: host=%s, uri=%s, method=%s, statusCode=%d", host, uri, method, statusCode)
		}
	})

	t.Run("generates consistent IDs for same input", func(t *testing.T) {
		id1 := generateMockID("example.com", "/api/users", "GET", 200)
		id2 := generateMockID("example.com", "/api/users", "GET", 200)
		if id1 != id2 {
			t.Errorf("expected same ID for same input, got '%s' and '%s'", id1, id2)
		}
	})

	t.Run("generates different IDs for different inputs", func(t *testing.T) {
		id1 := generateMockID("example.com", "/api/users", "GET", 200)
		id2 := generateMockID("example.com", "/api/users", "POST", 200)
		id3 := generateMockID("other.com", "/api/users", "GET", 200)
		id4 := generateMockID("example.com", "/api/users", "GET", 404)
		if id1 == id2 {
			t.Error("expected different IDs for different methods")
		}
		if id1 == id3 {
			t.Error("expected different IDs for different hosts")
		}
		if id1 == id4 {
			t.Error("expected different IDs for different status codes")
		}
	})
}

func TestMockAdminService_DeleteMockByID(t *testing.T) {
	t.Run("deletes mock by valid ID", func(t *testing.T) {
		contentService := &mockContentService{
			contents: map[string][]byte{
				"example.com:/api/users:GET:200": []byte(`{}`),
			},
			events: make(chan content.ContentEvent),
		}
		service := NewMockAdminService(contentService)

		id := generateMockID("example.com", "/api/users", "GET", 200)
		err := service.DeleteMockByID(id, "test-uuid")

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("returns ErrInvalidMockID for bad base64", func(t *testing.T) {
		contentService := &mockContentService{
			contents: make(map[string][]byte),
			events:   make(chan content.ContentEvent),
		}
		service := NewMockAdminService(contentService)

		err := service.DeleteMockByID("not-valid-base64!!!", "test-uuid")

		if err == nil {
			t.Fatal("expected error")
		}

		if !errors.Is(err, ErrInvalidMockID) {
			t.Errorf("expected ErrInvalidMockID, got %v", err)
		}
	})

	t.Run("returns ErrInvalidMockID for invalid format", func(t *testing.T) {
		contentService := &mockContentService{
			contents: make(map[string][]byte),
			events:   make(chan content.ContentEvent),
		}
		service := NewMockAdminService(contentService)

		// Valid base64 but missing separator
		err := service.DeleteMockByID("aW52YWxpZA==", "test-uuid")

		if err == nil {
			t.Fatal("expected error")
		}

		if !errors.Is(err, ErrInvalidMockID) {
			t.Errorf("expected ErrInvalidMockID, got %v", err)
		}
	})
}

func TestMockAdminService_GetMockContent(t *testing.T) {
	t.Run("returns content for valid ID", func(t *testing.T) {
		data := []byte(`{"message":"hello"}`)
		contentService := &mockContentService{
			contents: map[string][]byte{
				"example.com:/api/users:GET:200": data,
			},
			events: make(chan content.ContentEvent),
		}
		service := NewMockAdminService(contentService)

		id := generateMockID("example.com", "/api/users", "GET", 200)
		result, err := service.GetMockContent(id, "test-uuid")

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if string(result) != string(data) {
			t.Errorf("expected %q, got %q", string(data), string(result))
		}
	})

	t.Run("returns ErrInvalidMockID for bad ID", func(t *testing.T) {
		contentService := &mockContentService{
			contents: make(map[string][]byte),
			events:   make(chan content.ContentEvent),
		}
		service := NewMockAdminService(contentService)

		_, err := service.GetMockContent("not-valid!!!", "test-uuid")

		if err == nil {
			t.Fatal("expected error")
		}

		if !errors.Is(err, ErrInvalidMockID) {
			t.Errorf("expected ErrInvalidMockID, got %v", err)
		}
	})

	t.Run("returns ErrMockNotFound when content not found", func(t *testing.T) {
		contentService := &mockContentService{
			contents:    make(map[string][]byte),
			events:      make(chan content.ContentEvent),
			shouldError: true,
			errorMsg:    "not found",
		}
		service := NewMockAdminService(contentService)

		id := generateMockID("example.com", "/api/users", "GET", 200)
		_, err := service.GetMockContent(id, "test-uuid")

		if err == nil {
			t.Fatal("expected error")
		}

		if !errors.Is(err, ErrMockNotFound) {
			t.Errorf("expected ErrMockNotFound, got %v", err)
		}
	})

	t.Run("returns ErrMockNotFound when result data is nil", func(t *testing.T) {
		contentService := &nilDataContentService{}
		service := NewMockAdminService(contentService)

		id := generateMockID("example.com", "/api/users", "GET", 200)
		_, err := service.GetMockContent(id, "test-uuid")

		if err == nil {
			t.Fatal("expected error")
		}

		if !errors.Is(err, ErrMockNotFound) {
			t.Errorf("expected ErrMockNotFound, got %v", err)
		}
	})
}

// nilDataContentService returns a ContentResult with nil Data to test the nil-data branch.
type nilDataContentService struct{}

func (n *nilDataContentService) GetContent(host, uri, method, uuid string, statusCode int) (*content.ContentResult, error) {
	return &content.ContentResult{Data: nil}, nil
}

func (n *nilDataContentService) SetContent(host, uri, method, uuid string, statusCode int, data *[]byte) error {
	return nil
}

func (n *nilDataContentService) DeleteContent(host, uri, method, uuid string, statusCode int) error {
	return nil
}

func (n *nilDataContentService) ListContents(uuid string) (*[]content.ContentData, error) {
	return nil, nil
}

func (n *nilDataContentService) ListDefaultContents(uuid string) (*[]content.ContentData, error) {
	return nil, nil
}

func (n *nilDataContentService) Subscribe(subscriberId string, eventTypes ...content.ContentEventType) <-chan content.ContentEvent {
	return make(chan content.ContentEvent)
}

func (n *nilDataContentService) Unsubscribe(subscriberId string) {}

func TestDecodeMockID(t *testing.T) {
	t.Run("decodes valid ID", func(t *testing.T) {
		id := generateMockID("example.com", "/api/users", "GET", 200)
		host, uri, method, statusCode, err := decodeMockID(id)

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if host != "example.com" {
			t.Errorf("expected host 'example.com', got '%s'", host)
		}

		if uri != "/api/users" {
			t.Errorf("expected uri '/api/users', got '%s'", uri)
		}

		if method != "GET" {
			t.Errorf("expected method 'GET', got '%s'", method)
		}

		if statusCode != 200 {
			t.Errorf("expected statusCode 200, got %d", statusCode)
		}
	})

	t.Run("handles URI with special characters", func(t *testing.T) {
		id := generateMockID("example.com", "/api/users/:id/posts", "POST", 201)
		host, uri, method, statusCode, err := decodeMockID(id)

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if uri != "/api/users/:id/posts" {
			t.Errorf("expected uri '/api/users/:id/posts', got '%s'", uri)
		}

		if host != "example.com" || method != "POST" || statusCode != 201 {
			t.Errorf("host, method, or statusCode mismatch: host=%s method=%s statusCode=%d", host, method, statusCode)
		}
	})

	t.Run("returns error for invalid base64", func(t *testing.T) {
		_, _, _, _, err := decodeMockID("not-valid-base64!!!")

		if err == nil {
			t.Error("expected error for invalid base64")
		}
	})

	t.Run("returns error for invalid format", func(t *testing.T) {
		// Valid base64 but missing separator
		_, _, _, _, err := decodeMockID("aW52YWxpZA==") // "invalid" in base64

		if err == nil {
			t.Error("expected error for invalid format")
		}
	})
}
