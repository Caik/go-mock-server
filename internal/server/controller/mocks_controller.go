package controller

import (
	"strings"

	"github.com/Caik/go-mock-server/internal/service/mock"
	"github.com/Caik/go-mock-server/internal/util"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

var (
	badConfigurationResponseData = []byte("bad mock server configuration")
)

func handleMockRequest(c *gin.Context) {
	mockRequest := newMockRequest(c)
	mockResponse := mock.GetMockResponse(mockRequest)

	// bad mock server configuration
	if mockResponse == nil {
		log.WithField("uuid", mockRequest.Uuid).
			Warn("bad configuration found, mock response is nil!")

		mockResponse = &mock.MockResponse{
			StatusCode: 500,
			Data:       &badConfigurationResponseData,
		}
	}

	c.Data(mockResponse.StatusCode, mockResponse.ContentType, *mockResponse.Data)
}

func newMockRequest(c *gin.Context) mock.MockRequest {
	return mock.MockRequest{
		Host:   sanitizeHost(c.Request.Host),
		URI:    c.Request.RequestURI,
		Method: c.Request.Method,
		Accept: c.GetHeader("accept"),
		Uuid:   c.GetString(util.UuidKey),
	}
}

func sanitizeHost(host string) string {
	index := strings.Index(host, ":")

	if index == -1 {
		return host
	}

	return strings.ToLower(host[0:index])
}
