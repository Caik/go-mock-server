package mock

import (
	"math/rand"
	"strconv"

	"github.com/Caik/go-mock-server/internal/config"
	"github.com/rs/zerolog/log"
)

type statusSimulationMockService struct {
	next        mockService
	hostsConfig *config.HostsConfig
}

type statusPercentageWrapper struct {
	statusCode           int
	percentage           int
	originalStatusConfig config.StatusConfig
}

func (e *statusSimulationMockService) getMockResponse(mockRequest MockRequest) *MockResponse {
	statusesConfig, scope := e.hostsConfig.GetAppropriateStatusesConfig(mockRequest.Host, mockRequest.URI)

	statusCode := 200
	var drawnWrapper *statusPercentageWrapper

	if statusesConfig != nil {
		drawnWrapper = e.drawStatus(statusesConfig)

		if drawnWrapper != nil {
			statusCode = drawnWrapper.statusCode
		}
	}

	// Always set status code on request before passing downstream,
	// so ContentMockService can find the correct status-specific file.
	mockRequest.StatusCode = statusCode

	log.Info().
		Str("uuid", mockRequest.Uuid).
		Int("status_code", statusCode).
		Msg("simulating status")

	resp := e.nextOrNil(mockRequest)

	if drawnWrapper != nil && resp != nil {
		resp.StatusCode = statusCode
		resp.activeStatusConfig = &drawnWrapper.originalStatusConfig
		resp.AddMetadata(MetadataSimulatedStatus, "true")
		resp.AddMetadata(MetadataStatusRuleScope, scope)
	}

	return resp
}

func (e *statusSimulationMockService) setNext(next mockService) {
	e.next = next
}

func (e *statusSimulationMockService) nextOrNil(mockRequest MockRequest) *MockResponse {
	if e.next == nil {
		return nil
	}

	return e.next.getMockResponse(mockRequest)
}

func (e *statusSimulationMockService) drawStatus(statusesConfig *map[string]config.StatusConfig) *statusPercentageWrapper {
	statusesWrapper := make([]statusPercentageWrapper, len(*statusesConfig))
	i := 0

	for statusCode, statusConfig := range *statusesConfig {
		intStatusCode, _ := strconv.Atoi(statusCode)

		statusesWrapper[i] = statusPercentageWrapper{
			statusCode:           intStatusCode,
			percentage:           *statusConfig.Percentage,
			originalStatusConfig: statusConfig,
		}
		i++
	}

	// randomly selecting a status code based on the percentage drawn
	// Using 1-100 range so that a 0% status rate truly never fires (0 would satisfy draw <= 0)
	draw := rand.Intn(100) + 1
	sumStatusPercentage := 0

	for _, statusWrapper := range statusesWrapper {
		sumStatusPercentage += statusWrapper.percentage

		if draw <= sumStatusPercentage {
			return &statusWrapper
		}
	}

	return nil
}

func newStatusSimulationMockService(hostsConfig *config.HostsConfig) *statusSimulationMockService {
	return &statusSimulationMockService{
		hostsConfig: hostsConfig,
	}
}
