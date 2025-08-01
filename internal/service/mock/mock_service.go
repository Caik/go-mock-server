package mock

import (
	"github.com/Caik/go-mock-server/internal/config"
	"strings"
)

type mockService interface {
	getMockResponse(mockRequest MockRequest) *MockResponse
	setNext(next mockService)
}

type MockRequest struct {
	Host   string
	Method string
	URI    string
	Accept string
	Uuid   string
}

type MockResponse struct {
	StatusCode          int
	Data                *[]byte
	ContentType         string
	Headers             *map[string]string
	activeErrorConfig   *config.ErrorConfig
	activeLatencyConfig *config.LatencyConfig
}

// AddHeaders adds the provided headers to the MockResponse, merging with existing headers if present
func (m *MockResponse) AddHeaders(headers map[string]string) {
	if m.Headers == nil {
		m.Headers = &headers
	} else {
		// Merge headers with existing ones
		for key, value := range headers {
			(*m.Headers)[key] = value
		}
	}
}

func GenerateCacheKey(mockRequest MockRequest) string {
	return strings.Join([]string{mockRequest.Host, mockRequest.Method, mockRequest.URI}, ":")
}
