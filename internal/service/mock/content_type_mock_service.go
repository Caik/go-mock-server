package mock

import (
	"strings"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type contentTypeMockService struct {
	next mockService
}

func (c contentTypeMockService) getMockResponse(mockRequest MockRequest) *MockResponse {
	mockResponse := c.nextOrNil(mockRequest)

	if mockResponse == nil {
		return nil
	}

	if len(mockResponse.ContentType) > 0 {
		return mockResponse
	}

	contentType := c.setAppropriateContentType(mockRequest.Accept)

	log.WithField("uuid", mockRequest.Uuid).
		WithField("content_type", contentType).
		Info("setting content type")

	mockResponse.ContentType = contentType

	return mockResponse
}

func (c *contentTypeMockService) setNext(next mockService) {
	c.next = next
}

func (c contentTypeMockService) nextOrNil(mockRequest MockRequest) *MockResponse {
	if c.next == nil {
		return nil
	}

	return c.next.getMockResponse(mockRequest)
}

func (c contentTypeMockService) setAppropriateContentType(acceptHeader string) string {
	if len(acceptHeader) == 0 || strings.EqualFold(strings.TrimSpace(acceptHeader), "*/*") {
		return gin.MIMEPlain
	}

	parts := strings.Split(acceptHeader, ",")

	for _, accept := range parts {
		acceptParts := strings.Split(accept, ";")

		return acceptParts[0]
	}

	return gin.MIMEPlain
}
