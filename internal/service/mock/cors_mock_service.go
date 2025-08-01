package mock

import (
	"github.com/rs/zerolog/log"
)

type corsMockService struct {
	next mockService
}

func (c *corsMockService) getMockResponse(mockRequest MockRequest) *MockResponse {
	mockResponse := c.nextOrNil(mockRequest)

	if mockResponse == nil {
		return nil
	}

	// Add CORS headers to the response
	corsHeaders := c.getCorsHeaders()
	mockResponse.AddHeaders(corsHeaders)

	log.Info().
		Str("uuid", mockRequest.Uuid).
		Msg("adding CORS headers")

	return mockResponse
}

func (c *corsMockService) setNext(next mockService) {
	c.next = next
}

func (c *corsMockService) nextOrNil(mockRequest MockRequest) *MockResponse {
	if c.next == nil {
		return nil
	}

	return c.next.getMockResponse(mockRequest)
}

func (c *corsMockService) getCorsHeaders() map[string]string {
	return map[string]string{
		"Access-Control-Allow-Origin":  "*",
		"Access-Control-Allow-Methods": "GET, POST, PUT, DELETE, OPTIONS, PATCH, HEAD",
		"Access-Control-Allow-Headers": "Content-Type, Authorization, X-Requested-With, Accept, Origin",
		"Access-Control-Max-Age":       "86400",
	}
}

func newCorsMockService() *corsMockService {
	return &corsMockService{}
}
