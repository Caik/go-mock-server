package mock

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/Caik/go-mock-server/internal/rest"
	"github.com/Caik/go-mock-server/internal/service/content"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

var (
	errContentServiceNotFound = errors.New("internal server error: nil content service")
)

type contentMockService struct {
}

func (f contentMockService) getMockResponse(mockRequest MockRequest) *MockResponse {
	data, err := f.readMockFile(mockRequest)

	if err != nil {
		if err == errContentServiceNotFound {
			return f.new500Response(err)
		}

		return f.new404Response(err)
	}

	return &MockResponse{
		StatusCode: 200,
		Data:       data,
	}
}

func (f contentMockService) setNext(next mockService) {}

func (f contentMockService) readMockFile(mockRequest MockRequest) (*[]byte, error) {
	contentService := content.GetContentService()

	if contentService == nil {
		log.WithField("uuid", mockRequest.Uuid).
			Warn("bad configuration found, content service is nil!")

		return nil, errContentServiceNotFound
	}

	return contentService.GetContent(mockRequest.Host, mockRequest.URI, mockRequest.Method, mockRequest.Uuid)
}

func (f contentMockService) new404Response(err error) *MockResponse {
	msg := err.Error()

	res := rest.Response{
		Status:  rest.Fail,
		Message: msg,
	}

	data, err := json.Marshal(res)

	if err != nil {
		data = []byte(fmt.Sprintf("{%q:%q,%q:%q}", "status", res.Status, "message", res.Message))
	}

	return &MockResponse{
		StatusCode:  http.StatusNotFound,
		Data:        &data,
		ContentType: gin.MIMEJSON,
	}
}

func (f contentMockService) new500Response(err error) *MockResponse {
	msg := err.Error()

	res := rest.Response{
		Status:  rest.Error,
		Message: msg,
	}

	data, err := json.Marshal(res)

	if err != nil {
		data = []byte(fmt.Sprintf("{%q:%q,%q:%q}", "status", res.Status, "message", res.Message))
	}

	return &MockResponse{
		StatusCode:  http.StatusInternalServerError,
		Data:        &data,
		ContentType: gin.MIMEJSON,
	}
}

func NewContentMockService() *contentMockService {
	return &contentMockService{}
}
