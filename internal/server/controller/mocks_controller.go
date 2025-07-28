package controller

import (
	"strings"

	"github.com/Caik/go-mock-server/internal/service/mock"
	"github.com/Caik/go-mock-server/internal/util"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

var (
	badConfigurationResponseData = []byte("bad mock server configuration")
)

type MocksController struct {
	factory *mock.MockServiceFactory
}

func (m *MocksController) handleMockRequest(c *gin.Context) {
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

func NewMocksController(factory *mock.MockServiceFactory) *MocksController {
	controller := MocksController{
		factory: factory,
	}

	return &controller
}
