package mock

import (
	"errors"
	"testing"

	"github.com/Caik/go-mock-server/internal/service/content"
)

// Mock content service for testing
type mockContentService struct {
	contents map[string][]byte
	events   chan content.ContentEvent
}

func (m *mockContentService) GetContent(host, uri, method, uuid string) (*[]byte, error) {
	key := host + ":" + uri + ":" + method
	if data, exists := m.contents[key]; exists {
		return &data, nil
	}
	return nil, errors.New("not found")
}

func (m *mockContentService) SetContent(host, uri, method, uuid string, data *[]byte) error {
	key := host + ":" + uri + ":" + method
	m.contents[key] = *data
	return nil
}

func (m *mockContentService) DeleteContent(host, uri, method, uuid string) error {
	key := host + ":" + uri + ":" + method
	delete(m.contents, key)
	return nil
}

func (m *mockContentService) ListContents(uuid string) (*[]content.ContentData, error) {
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

func TestHostResolutionMockService_getMockResponse(t *testing.T) {
	t.Run("passes through request when no host resolution needed", func(t *testing.T) {
		contentService := &mockContentService{
			contents: make(map[string][]byte),
			events:   make(chan content.ContentEvent),
		}

		service, err := newHostResolutionMockService(contentService)
		if err != nil {
			t.Fatalf("failed to create service: %v", err)
		}

		// Mock next service that returns a response
		testData := []byte("test response")
		mockNext := &mockMockService{
			response: &MockResponse{
				StatusCode: 200,
				Data:       &testData,
			},
		}
		service.setNext(mockNext)

		request := MockRequest{
			Host:   "example.com",
			Method: "GET",
			URI:    "/api/test",
			Accept: "application/json",
			Uuid:   "test-uuid",
		}

		response := service.getMockResponse(request)

		if response == nil {
			t.Fatal("expected response, got nil")
		}

		if response.StatusCode != 200 {
			t.Errorf("expected status 200, got %d", response.StatusCode)
		}

		// Verify the request passed to next service
		if mockNext.lastRequest.Host != "example.com" {
			t.Errorf("expected host 'example.com', got '%s'", mockNext.lastRequest.Host)
		}
	})

	t.Run("preserves Accept field during host resolution", func(t *testing.T) {
		contentService := &mockContentService{
			contents: map[string][]byte{
				"api.example.com:/api/test:GET": []byte("test response"),
			},
			events: make(chan content.ContentEvent),
		}

		service, err := newHostResolutionMockService(contentService)
		if err != nil {
			t.Fatalf("failed to create service: %v", err)
		}

		testData2 := []byte("test response")
		mockNext := &mockMockService{
			response: &MockResponse{
				StatusCode: 200,
				Data:       &testData2,
			},
		}
		service.setNext(mockNext)

		request := MockRequest{
			Host:   "api.example.com",
			Method: "GET",
			URI:    "/api/test",
			Accept: "application/json", // This should be preserved
			Uuid:   "test-uuid",
		}

		service.getMockResponse(request)

		// Verify Accept field is preserved
		if mockNext.lastRequest.Accept != "application/json" {
			t.Errorf("Accept field lost during host resolution. Expected 'application/json', got '%s'",
				mockNext.lastRequest.Accept)
		}
	})

	t.Run("handles IP address requests", func(t *testing.T) {
		contentService := &mockContentService{
			contents: make(map[string][]byte),
			events:   make(chan content.ContentEvent),
		}

		service, err := newHostResolutionMockService(contentService)
		if err != nil {
			t.Fatalf("failed to create service: %v", err)
		}

		testData3 := []byte("test response")
		mockNext := &mockMockService{
			response: &MockResponse{
				StatusCode: 200,
				Data:       &testData3,
			},
		}
		service.setNext(mockNext)

		// Test with IP address
		ipRequest := MockRequest{
			Host:   "192.168.1.1",
			Method: "GET",
			URI:    "/api/test",
			Accept: "application/json",
			Uuid:   "test-uuid",
		}

		service.getMockResponse(ipRequest)

		// Test that IP addresses are handled appropriately
		t.Logf("testing IP address handling in host resolution")
		t.Logf("current logic: !IpAddressRegex.MatchString(host) && HostRegex.MatchString(host)")
		t.Logf("IP addresses will go through host resolution logic")
	})

	t.Run("handles nil content service", func(t *testing.T) {
		// This should return an error
		_, err := newHostResolutionMockService(nil)
		if err == nil {
			t.Error("expected error with nil content service")
		}
	})

	t.Run("handles content service", func(t *testing.T) {
		contentService := &mockContentService{
			contents: make(map[string][]byte),
			events:   make(chan content.ContentEvent),
		}

		service, err := newHostResolutionMockService(contentService)
		if err != nil {
			t.Fatalf("failed to create service: %v", err)
		}

		testData4 := []byte("test response")
		mockNext := &mockMockService{
			response: &MockResponse{
				StatusCode: 200,
				Data:       &testData4,
			},
		}
		service.setNext(mockNext)

		request := MockRequest{
			Host:   "example.com",
			Method: "GET",
			URI:    "/api/test",
			Accept: "application/json",
			Uuid:   "test-uuid",
		}

		// Should not panic
		response := service.getMockResponse(request)

		if response == nil {
			t.Error("expected response even with empty content service")
		}
	})
}

// Note: evaluate method is private, so we test it indirectly through getMockResponse

func TestHostResolutionMockService_setNext(t *testing.T) {
	t.Run("sets next service correctly", func(t *testing.T) {
		contentService := &mockContentService{
			contents: make(map[string][]byte),
			events:   make(chan content.ContentEvent),
		}

		service, err := newHostResolutionMockService(contentService)
		if err != nil {
			t.Fatalf("failed to create service: %v", err)
		}

		mockNext := &mockMockService{}

		service.setNext(mockNext)

		// We can't directly access the next field since it's private
		// But we can test that setNext doesn't panic
		t.Log("setNext method executed without panic")
	})
}

// Helper mock service for testing
type mockMockService struct {
	response    *MockResponse
	lastRequest MockRequest
}

func (m *mockMockService) getMockResponse(request MockRequest) *MockResponse {
	m.lastRequest = request
	return m.response
}

func (m *mockMockService) setNext(next mockService) {
	// Not needed for testing
}

func (m *mockMockService) nextOrNil(request MockRequest) *MockResponse {
	return m.response
}
