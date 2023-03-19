package mock

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/Caik/go-mock-server/internal/rest"
	"github.com/Caik/go-mock-server/internal/service/file"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type fileMockService struct {
}

func (f fileMockService) getMockResponse(mockRequest MockRequest) *MockResponse {
	data, err := f.readMockFile(mockRequest)

	if err != nil {
		return f.new404Response(err)
	}

	return &MockResponse{
		StatusCode: 200,
		Data:       data,
	}
}

func (f fileMockService) setNext(next mockService) {}

func (f fileMockService) readMockFile(mockRequest MockRequest) (*[]byte, error) {
	finalFilePath, err := file.GetFinalFilePath(mockRequest.Host, mockRequest.URI, mockRequest.Method)

	if err != nil {
		log.WithField("uuid", mockRequest.Uuid).
			WithField("path", finalFilePath).
			Info("error while generating final file path")

		return nil, err
	}

	data, err := os.ReadFile(finalFilePath)

	if err != nil {
		log.WithField("uuid", mockRequest.Uuid).
			WithField("path", finalFilePath).
			Info("mock not found")

		return nil, errors.New("mock not found")
	}

	return &data, err
}

func (f fileMockService) new404Response(err error) *MockResponse {
	msg := err.Error()

	restResponse := rest.Response{
		Status:  rest.Fail,
		Message: msg,
	}

	data, err := json.Marshal(restResponse)

	if err != nil {
		data = []byte(fmt.Sprintf("{%q:%q,%q:%q}", "status", rest.Fail, "message", msg))
	}

	return &MockResponse{
		StatusCode:  http.StatusNotFound,
		Data:        &data,
		ContentType: gin.MIMEJSON,
	}
}
