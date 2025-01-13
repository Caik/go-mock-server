package mock

import (
	"github.com/Caik/go-mock-server/internal/config"
	log "github.com/sirupsen/logrus"
	"math/rand"
	"strconv"
)

type errorMockService struct {
	next        mockService
	hostsConfig *config.HostsConfig
}

type errorPercentageWrapper struct {
	statusCode          int
	percentage          int
	originalErrorConfig config.ErrorConfig
}

func (e *errorMockService) getMockResponse(mockRequest MockRequest) *MockResponse {
	errorsConfig := e.hostsConfig.GetAppropriateErrorsConfig(mockRequest.Host, mockRequest.URI)

	if errorsConfig == nil {
		return e.nextOrNil(mockRequest)
	}

	drawnErrorWrapper := e.drawError(errorsConfig)

	// no error has been drawn
	if drawnErrorWrapper == nil {
		return e.nextOrNil(mockRequest)
	}

	log.WithField("uuid", mockRequest.Uuid).
		WithField("status_code", drawnErrorWrapper.statusCode).
		Info("simulating error")

	emptyResponse := []byte("")

	return &MockResponse{
		StatusCode:        drawnErrorWrapper.statusCode,
		Data:              &emptyResponse,
		activeErrorConfig: &drawnErrorWrapper.originalErrorConfig,
	}
}

func (e *errorMockService) setNext(next mockService) {
	e.next = next
}

func (e *errorMockService) nextOrNil(mockRequest MockRequest) *MockResponse {
	if e.next == nil {
		return nil
	}

	return e.next.getMockResponse(mockRequest)
}

func (e *errorMockService) drawError(errorsConfig *map[string]config.ErrorConfig) *errorPercentageWrapper {
	errorsWrapper := make([]errorPercentageWrapper, len(*errorsConfig))
	i := 0

	for statusCode, errorConfig := range *errorsConfig {
		intStatusCode, _ := strconv.Atoi(statusCode)

		errorsWrapper[i] = errorPercentageWrapper{
			statusCode:          intStatusCode,
			percentage:          *errorConfig.Percentage,
			originalErrorConfig: errorConfig,
		}
		i++
	}

	// randomly selecting a status code based on the percentage drawn
	draw := rand.Intn(101)
	sumErrorPercentage := 0

	for _, errorWrapper := range errorsWrapper {
		sumErrorPercentage += errorWrapper.percentage

		if draw <= sumErrorPercentage {
			return &errorWrapper
		}
	}

	return nil
}

func newErrorMockService(hostsConfig *config.HostsConfig) *errorMockService {
	return &errorMockService{
		hostsConfig: hostsConfig,
	}
}
