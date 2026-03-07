package mock

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/Caik/go-mock-server/internal/rest"
	"github.com/Caik/go-mock-server/internal/service/content"
	"github.com/gin-gonic/gin"
)

var (
	errContentServiceNotFound = errors.New("internal server error: nil content service")
)

type contentMockService struct {
	contentService content.ContentService
}

func (c *contentMockService) getMockResponse(mockRequest MockRequest) *MockResponse {
	result, err := c.contentService.GetContent(mockRequest.Host, mockRequest.URI, mockRequest.Method, mockRequest.Uuid)

	if err != nil {
		if errors.Is(err, errContentServiceNotFound) {
			return c.new500Response(err)
		}

		return c.new404Response(err)
	}

	resp := &MockResponse{
		StatusCode: 200,
		Data:       result.Data,
	}

	resp.AddMetadata(MetadataMatched, "true")
	resp.AddMetadata(MetadataSource, result.Source)
	resp.AddMetadata(MetadataPath, result.Path)

	return resp
}

func (c *contentMockService) setNext(next mockService) {}

func (c *contentMockService) new404Response(err error) *MockResponse {
	msg := err.Error()

	res := rest.Response{
		Status:  rest.Fail,
		Message: msg,
	}

	data, err := json.Marshal(res)

	if err != nil {
		data = []byte(fmt.Sprintf("{%q:%q,%q:%q}", "status", res.Status, "message", res.Message))
	}

	resp := &MockResponse{
		StatusCode:  http.StatusNotFound,
		Data:        &data,
		ContentType: gin.MIMEJSON,
	}

	resp.AddMetadata(MetadataMatched, "false")

	return resp
}

func (c *contentMockService) new500Response(err error) *MockResponse {
	msg := err.Error()

	res := rest.Response{
		Status:  rest.Error,
		Message: msg,
	}

	data, err := json.Marshal(res)

	if err != nil {
		data = []byte(fmt.Sprintf("{%q:%q,%q:%q}", "status", res.Status, "message", res.Message))
	}

	resp := &MockResponse{
		StatusCode:  http.StatusInternalServerError,
		Data:        &data,
		ContentType: gin.MIMEJSON,
	}

	resp.AddMetadata(MetadataMatched, "false")

	return resp
}

func newContentMockService(contentService content.ContentService) *contentMockService {
	return &contentMockService{
		contentService: contentService,
	}
}
