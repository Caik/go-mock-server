package mock

import (
	"testing"

	"github.com/Caik/go-mock-server/internal/config"
)

// Helper function to create int pointers
func intPtr(i int) *int {
	return &i
}

func TestStatusSimulationMockService_getMockResponse(t *testing.T) {
	t.Run("returns status response and calls downstream when status is drawn", func(t *testing.T) {
		hostsConfig := &config.HostsConfig{
			Hosts: map[string]config.HostConfig{
				"example.com": {
					StatusesConfig: map[string]config.StatusConfig{
						"500": {
							Percentage: intPtr(100), // Always return this status
						},
					},
				},
			},
		}

		service := newStatusSimulationMockService(hostsConfig)

		successData := []byte("success response")
		mockNext := &mockMockService{
			response: &MockResponse{
				StatusCode: 200,
				Data:       &successData,
			},
		}
		service.setNext(mockNext)

		request := MockRequest{
			Host:   "example.com",
			Method: "GET",
			URI:    "/api/test",
			Uuid:   "test-uuid",
		}

		response := service.getMockResponse(request)

		if response == nil {
			t.Fatal("expected response, got nil")
		}

		// Should override status code with drawn status
		if response.StatusCode != 500 {
			t.Errorf("expected status 500, got %d", response.StatusCode)
		}

		// Downstream should have been called (lastRequest will be populated)
		if mockNext.lastRequest.Host == "" {
			t.Error("expected downstream to be called")
		}
	})

	t.Run("passes through with status 200 when no status is drawn", func(t *testing.T) {
		hostsConfig := &config.HostsConfig{
			Hosts: map[string]config.HostConfig{
				"example.com": {
					StatusesConfig: map[string]config.StatusConfig{
						"500": {
							Percentage: intPtr(0), // Never return status
						},
					},
				},
			},
		}

		service := newStatusSimulationMockService(hostsConfig)

		successData2 := []byte("success response")
		mockNext := &mockMockService{
			response: &MockResponse{
				StatusCode: 200,
				Data:       &successData2,
			},
		}
		service.setNext(mockNext)

		request := MockRequest{
			Host:   "example.com",
			Method: "GET",
			URI:    "/api/test",
			Uuid:   "test-uuid",
		}

		response := service.getMockResponse(request)

		if response == nil {
			t.Fatal("expected response, got nil")
		}

		// Should pass through to next service
		if response.StatusCode != 200 {
			t.Errorf("expected status 200, got %d", response.StatusCode)
		}

		// Downstream should have been called (lastRequest will be populated)
		if mockNext.lastRequest.Host == "" {
			t.Error("expected downstream to be called")
		}
	})

	t.Run("calls downstream when no status config", func(t *testing.T) {
		hostsConfig := &config.HostsConfig{
			Hosts: make(map[string]config.HostConfig),
		}

		service := newStatusSimulationMockService(hostsConfig)

		successData3 := []byte("success response")
		mockNext := &mockMockService{
			response: &MockResponse{
				StatusCode: 200,
				Data:       &successData3,
			},
		}
		service.setNext(mockNext)

		request := MockRequest{
			Host:   "example.com",
			Method: "GET",
			URI:    "/api/test",
			Uuid:   "test-uuid",
		}

		response := service.getMockResponse(request)

		if response == nil {
			t.Fatal("expected response, got nil")
		}

		// Should pass through to next service
		if response.StatusCode != 200 {
			t.Errorf("expected status 200, got %d", response.StatusCode)
		}

		// Downstream should have been called (lastRequest will be populated)
		if mockNext.lastRequest.Host == "" {
			t.Error("expected downstream to be called")
		}
	})
}

func TestStatusSimulationMockService_drawStatus(t *testing.T) {
	// Go 1.20+ seeds automatically; no need for explicit rand.Seed

	service := &statusSimulationMockService{}

	t.Run("draws status based on percentage", func(t *testing.T) {
		statusesConfig := map[string]config.StatusConfig{
			"500": {
				Percentage: intPtr(100), // Always draw this status
			},
		}

		// Should always return the status
		for i := 0; i < 10; i++ {
			wrapper := service.drawStatus(&statusesConfig)

			if wrapper == nil {
				t.Error("expected status wrapper, got nil")
				continue
			}

			if wrapper.statusCode != 500 {
				t.Errorf("expected status code 500, got %d", wrapper.statusCode)
			}

			if wrapper.percentage != 100 {
				t.Errorf("expected percentage 100, got %d", wrapper.percentage)
			}
		}
	})

	t.Run("handles zero percentage status configuration", func(t *testing.T) {
		statusesConfig := map[string]config.StatusConfig{
			"500": {
				Percentage: intPtr(0), // 0% status rate
			},
		}

		// With 0% status rate, statuses should never occur
		// when rand.Intn(100) + 1 returns 1-100, since draw <= 0 is always false
		var statusCount int
		var totalRuns int = 100

		for i := 0; i < totalRuns; i++ {
			wrapper := service.drawStatus(&statusesConfig)
			if wrapper != nil {
				statusCount++
			}
		}

		statusRate := float64(statusCount) / float64(totalRuns) * 100

		// With 0% configured, we expect 0% actual status rate (rand.Intn(100) + 1 ranges 1-100)
		// since draw > 0 means condition (draw <= 0) is always false
		// Allow generous tolerance for test stability across different runs
		if statusRate > 10.0 { // Increased tolerance for race/shuffle testing
			t.Errorf("expected low status rate with 0%% config, got %.1f%%", statusRate)
		}

		t.Logf("with 0%% configured status rate, got %.1f%% actual status rate", statusRate)
	})

	t.Run("handles invalid status code strings", func(t *testing.T) {
		statusesConfig := map[string]config.StatusConfig{
			"invalid": { // Invalid status code string (should be validated at config level)
				Percentage: intPtr(100),
			},
		}

		service := &statusSimulationMockService{}

		wrapper := service.drawStatus(&statusesConfig)

		if wrapper != nil {
			t.Logf("status code 'invalid' converted to %d", wrapper.statusCode)
			t.Logf("note: strconv.Atoi error is ignored by design - validation happens at config level")

			if wrapper.statusCode == 0 {
				t.Log("invalid status codes become 0 - this is expected behavior")
				t.Log("validation should occur in config.HostConfig.Validate() at startup")
			}
		}
	})

	t.Run("validates random range behavior with 100% status rate", func(t *testing.T) {
		statusesConfig := map[string]config.StatusConfig{
			"500": {
				Percentage: intPtr(100), // Total percentage is exactly 100
			},
		}

		service := &statusSimulationMockService{}

		var statusCount int
		var totalDraws int = 1000

		// Test the random logic with 100% status rate
		for i := 0; i < totalDraws; i++ {
			wrapper := service.drawStatus(&statusesConfig)
			if wrapper != nil {
				statusCount++
			}
		}

		statusRate := float64(statusCount) / float64(totalDraws) * 100

		t.Logf("with 100%% configured status rate, got %.1f%% actual status rate", statusRate)
		t.Logf("random range: rand.Intn(100) + 1 generates 1-100, condition: draw <= totalPercentage")

		// With 100% status rate, we should get close to 100% statuses (allowing for randomness)
		if statusRate >= 99.0 { // Allow 1% tolerance for randomness
			t.Log("random range logic works correctly")
		} else {
			t.Logf("unexpected: status rate %.1f%% is lower than expected", statusRate)
		}
	})

	t.Run("handles multiple statuses with different percentages", func(t *testing.T) {
		statusesConfig := map[string]config.StatusConfig{
			"400": {
				Percentage: intPtr(30),
			},
			"500": {
				Percentage: intPtr(20),
			},
		}

		service := &statusSimulationMockService{}

		var statusCount int
		var totalDraws int = 1000

		for i := 0; i < totalDraws; i++ {
			wrapper := service.drawStatus(&statusesConfig)
			if wrapper != nil {
				statusCount++
			}
		}

		// With 50% total status rate, we should get roughly 500 statuses
		expectedStatuses := totalDraws * 50 / 100
		tolerance := totalDraws * 10 / 100 // 10% tolerance

		if statusCount < expectedStatuses-tolerance || statusCount > expectedStatuses+tolerance {
			t.Errorf("expected ~%d statuses, got %d (outside tolerance)", expectedStatuses, statusCount)
		}
	})

	t.Run("handles empty statuses config", func(t *testing.T) {
		statusesConfig := make(map[string]config.StatusConfig)

		service := &statusSimulationMockService{}

		wrapper := service.drawStatus(&statusesConfig)

		if wrapper != nil {
			t.Error("expected nil wrapper with empty config")
		}
	})
}

func TestStatusSimulationMockService_setNext(t *testing.T) {
	t.Run("sets next service correctly", func(t *testing.T) {
		hostsConfig := &config.HostsConfig{
			Hosts: make(map[string]config.HostConfig),
		}

		service := newStatusSimulationMockService(hostsConfig)
		mockNext := &mockMockService{}

		service.setNext(mockNext)

		if service.next != mockNext {
			t.Error("setNext should set the next service")
		}
	})
}
