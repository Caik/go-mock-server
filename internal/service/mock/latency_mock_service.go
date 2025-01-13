package mock

import (
	"github.com/Caik/go-mock-server/internal/config"
	"math/rand"
	"time"

	"github.com/rs/zerolog/log"
)

type latencyMockService struct {
	next        mockService
	hostsConfig *config.HostsConfig
}

func (l *latencyMockService) getMockResponse(mockRequest MockRequest) *MockResponse {
	startTime := time.Now()

	// calling next in the chain
	mockResponse := l.nextOrNil(mockRequest)

	// getting default appropriate latency config
	latencyConfig := l.hostsConfig.GetAppropriateLatencyConfig(mockRequest.Host, mockRequest.URI)

	// overriding the default latency config, if an error has been drawn
	if mockResponse.activeErrorConfig != nil && mockResponse.activeErrorConfig.LatencyConfig != nil {
		latencyConfig = mockResponse.activeErrorConfig.LatencyConfig
	}

	if latencyConfig == nil {
		return mockResponse
	}

	mockResponse.activeLatencyConfig = latencyConfig
	drawnLatency := l.drawLatency(latencyConfig)

	// simulating the latency
	log.Info().
		Str("uuid", mockRequest.Uuid).
		Dur("target_latency", time.Duration(drawnLatency*int(time.Millisecond))).
		Dur("processing_latency", time.Since(startTime)).
		Msg("simulating latency")

	targetLatencyTime := startTime.Add(time.Duration(drawnLatency * int(time.Millisecond)))
	<-time.NewTimer(time.Until(targetLatencyTime)).C

	return mockResponse
}

func (l *latencyMockService) setNext(next mockService) {
	l.next = next
}

func (l *latencyMockService) nextOrNil(mockRequest MockRequest) *MockResponse {
	if l.next == nil {
		return nil
	}

	return l.next.getMockResponse(mockRequest)
}

func (l *latencyMockService) drawLatency(latencyConfig *config.LatencyConfig) int {
	hasP95 := latencyConfig.P95 != nil
	hasP99 := latencyConfig.P99 != nil

	drawn := rand.Intn(100)

	drawLatencyWithUpperAndLowerBounds := func(lowerBound, upperBound *int) int {
		return rand.Intn(*upperBound-*lowerBound+1) + *lowerBound
	}

	if drawn <= 1 && hasP99 {
		return drawLatencyWithUpperAndLowerBounds(latencyConfig.P99, latencyConfig.Max)
	}

	if drawn <= 5 && hasP95 {
		return drawLatencyWithUpperAndLowerBounds(latencyConfig.P95, latencyConfig.P99)
	}

	if !hasP99 && !hasP95 {
		return drawLatencyWithUpperAndLowerBounds(latencyConfig.Min, latencyConfig.Max)
	}

	if hasP95 {
		return drawLatencyWithUpperAndLowerBounds(latencyConfig.Min, latencyConfig.P95)
	}

	// hasP99
	return drawLatencyWithUpperAndLowerBounds(latencyConfig.Min, latencyConfig.P99)
}

func newLatencyMockService(hostsConfig *config.HostsConfig) *latencyMockService {
	return &latencyMockService{
		hostsConfig: hostsConfig,
	}
}
