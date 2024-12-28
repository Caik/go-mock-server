package mock

import (
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"sync"

	"github.com/Caik/go-mock-server/internal/config"
	log "github.com/sirupsen/logrus"
)

type errorMockService struct {
	next        mockService
	once        sync.Once
	hostsConfig *config.HostsConfig
}

type errorPercentageWrapper struct {
	statusCode          int
	percentage          int
	originalErrorConfig config.ErrorConfig
}

func (e *errorMockService) getMockResponse(mockRequest MockRequest) *MockResponse {
	if err := e.ensureInit(); err != nil {
		return e.nextOrNil(mockRequest)
	}

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

func (e *errorMockService) ensureInit() error {
	if e.hostsConfig != nil {
		return nil
	}

	e.once.Do(func() {
		newHostsConfig, err := config.GetHostsConfig()

		if err != nil {
			log.Error(fmt.Sprintf("error while getting hosts config: %v", err))
			return
		}

		e.hostsConfig = newHostsConfig
	})

	if e.hostsConfig == nil {
		return errors.New("error while getting hosts config")
	}

	return nil
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

func newErrorMockService() *errorMockService {
	service := errorMockService{}
	service.ensureInit()

	return &service
}
