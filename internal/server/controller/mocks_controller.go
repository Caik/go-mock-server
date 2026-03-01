package controller

import (
	"strings"
	"time"

	"github.com/Caik/go-mock-server/internal/service/mock"
	"github.com/Caik/go-mock-server/internal/service/traffic"
	"github.com/Caik/go-mock-server/internal/util"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

var (
	badConfigurationResponseData = []byte("bad mock server configuration")
)

// MockResponseProvider is an interface for getting mock responses
// This allows for easier testing by enabling mock implementations
type MockResponseProvider interface {
	GetMockResponse(mockRequest mock.MockRequest) *mock.MockResponse
}

type MocksController struct {
	factory           MockResponseProvider
	trafficLogService *traffic.TrafficLogService
}

func (m *MocksController) handleMockRequest(c *gin.Context) {
	startTime := time.Now()
	mockRequest := m.newMockRequest(c)
	mockResponse := m.factory.GetMockResponse(mockRequest)

	// bad mock server configuration
	if mockResponse == nil {
		log.Warn().
			Str("uuid", mockRequest.Uuid).
			Msg("Mock response is nil")

		mockResponse = &mock.MockResponse{
			StatusCode: 500,
			Data:       &badConfigurationResponseData,
		}
	}

	// Set additional headers if present
	if mockResponse.Headers != nil {
		for key, value := range *mockResponse.Headers {
			c.Header(key, value)
		}
	}

	c.Data(mockResponse.StatusCode, mockResponse.ContentType, *mockResponse.Data)

	// Capture traffic after response is sent
	m.captureTraffic(c, mockRequest, mockResponse, startTime)
}

func (m *MocksController) newMockRequest(c *gin.Context) mock.MockRequest {
	return mock.MockRequest{
		Host:   m.sanitizeHost(c.Request.Host),
		URI:    c.Request.RequestURI,
		Method: c.Request.Method,
		Accept: c.GetHeader("accept"),
		Uuid:   c.GetString(util.UuidKey),
	}
}

func (m *MocksController) sanitizeHost(host string) string {
	index := strings.Index(host, ":")

	if index == -1 {
		return host
	}

	return strings.ToLower(host[0:index])
}

func NewMocksController(factory MockResponseProvider, trafficLogService *traffic.TrafficLogService) *MocksController {
	controller := MocksController{
		factory:           factory,
		trafficLogService: trafficLogService,
	}

	return &controller
}

// captureTraffic logs the request/response traffic for debugging
func (m *MocksController) captureTraffic(c *gin.Context, mockRequest mock.MockRequest, mockResponse *mock.MockResponse, startTime time.Time) {
	if m.trafficLogService == nil {
		return
	}

	// Build TrafficEntry from request and response
	entry := traffic.TrafficEntry{
		UUID:      mockRequest.Uuid,
		Timestamp: startTime,
		Request: traffic.TrafficRequest{
			Method: mockRequest.Method,
			Host:   mockRequest.Host,
			Path:   c.Request.URL.Path,
			Query:  c.Request.URL.RawQuery,
		},
		Response: traffic.TrafficResponse{
			StatusCode:  mockResponse.StatusCode,
			ContentType: mockResponse.ContentType,
			BodySize:    len(*mockResponse.Data),
			LatencyMs:   time.Since(startTime).Milliseconds(),
		},
	}

	// Set mock metadata
	if mockResponse.Metadata != nil {
		entry.Mock = traffic.TrafficMock{
			Matched: mockResponse.Metadata.Matched,
			Source:  mockResponse.Metadata.Source,
			Path:    mockResponse.Metadata.Path,
		}
	}

	m.trafficLogService.Capture(entry)
}
