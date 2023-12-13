package mock

import (
	"strings"

	"github.com/Caik/go-mock-server/internal/config"
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
	activeErrorConfig   *config.ErrorConfig
	activeLatencyConfig *config.LatencyConfig
}

func GenerateCacheKey(mockRequest MockRequest) string {
	return strings.Join([]string{mockRequest.Host, mockRequest.Method, mockRequest.URI}, ":")
}
