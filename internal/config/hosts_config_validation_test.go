package config

import (
	"strings"
	"testing"
)

// Test LatencyConfig validation

func TestLatencyConfig_Validate_Valid(t *testing.T) {
	tests := []struct {
		name   string
		config LatencyConfig
	}{
		{
			name: "min and max only",
			config: LatencyConfig{
				Min: intPtr(100),
				Max: intPtr(200),
			},
		},
		{
			name: "with P95",
			config: LatencyConfig{
				Min: intPtr(100),
				P95: intPtr(150),
				Max: intPtr(200),
			},
		},
		{
			name: "with P95 and P99",
			config: LatencyConfig{
				Min: intPtr(100),
				P95: intPtr(150),
				P99: intPtr(180),
				Max: intPtr(200),
			},
		},
		{
			name: "P95 equals min",
			config: LatencyConfig{
				Min: intPtr(100),
				P95: intPtr(100),
				Max: intPtr(200),
			},
		},
		{
			name: "P99 equals P95",
			config: LatencyConfig{
				Min: intPtr(100),
				P95: intPtr(150),
				P99: intPtr(150),
				Max: intPtr(200),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.validate()
			if err != nil {
				t.Errorf("expected no error for valid config, got %v", err)
			}
		})
	}
}

func TestLatencyConfig_Validate_Invalid(t *testing.T) {
	tests := []struct {
		name        string
		config      LatencyConfig
		expectedErr string
	}{
		{
			name: "missing min",
			config: LatencyConfig{
				Max: intPtr(200),
			},
			expectedErr: "you should define at least 'min' and 'max'",
		},
		{
			name: "missing max",
			config: LatencyConfig{
				Min: intPtr(100),
			},
			expectedErr: "you should define at least 'min' and 'max'",
		},
		{
			name: "min greater than max",
			config: LatencyConfig{
				Min: intPtr(200),
				Max: intPtr(100),
			},
			expectedErr: "min can not be greater than max",
		},
		{
			name: "P95 less than min",
			config: LatencyConfig{
				Min: intPtr(100),
				P95: intPtr(50),
				Max: intPtr(200),
			},
			expectedErr: "p95 can not be lesser than min or greater than max",
		},
		{
			name: "P95 greater than max",
			config: LatencyConfig{
				Min: intPtr(100),
				P95: intPtr(250),
				Max: intPtr(200),
			},
			expectedErr: "p95 can not be lesser than min or greater than max",
		},
		{
			name: "P99 less than min",
			config: LatencyConfig{
				Min: intPtr(100),
				P95: intPtr(150),
				P99: intPtr(50),
				Max: intPtr(200),
			},
			expectedErr: "p99 can not be lesser than min/p95 or greater than max",
		},
		{
			name: "P99 less than P95",
			config: LatencyConfig{
				Min: intPtr(100),
				P95: intPtr(150),
				P99: intPtr(140),
				Max: intPtr(200),
			},
			expectedErr: "p99 can not be lesser than min/p95 or greater than max",
		},
		{
			name: "P99 greater than max",
			config: LatencyConfig{
				Min: intPtr(100),
				P95: intPtr(150),
				P99: intPtr(250),
				Max: intPtr(200),
			},
			expectedErr: "p99 can not be lesser than min/p95 or greater than max",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.validate()
			if err == nil {
				t.Error("expected error for invalid config")
			} else if err.Error() != "invalid latency config found: "+tt.expectedErr {
				t.Errorf("expected error message to contain '%s', got '%s'", tt.expectedErr, err.Error())
			}
		})
	}
}

// Test ErrorConfig validation

func TestErrorConfig_Validate_Valid(t *testing.T) {
	tests := []struct {
		name   string
		config ErrorConfig
	}{
		{
			name: "percentage only",
			config: ErrorConfig{
				Percentage: intPtr(50),
			},
		},
		{
			name: "with latency config",
			config: ErrorConfig{
				Percentage: intPtr(25),
				LatencyConfig: &LatencyConfig{
					Min: intPtr(100),
					Max: intPtr(200),
				},
			},
		},
		{
			name: "minimum percentage",
			config: ErrorConfig{
				Percentage: intPtr(1),
			},
		},
		{
			name: "maximum percentage",
			config: ErrorConfig{
				Percentage: intPtr(100),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.validate()
			if err != nil {
				t.Errorf("expected no error for valid config, got %v", err)
			}
		})
	}
}

func TestErrorConfig_Validate_Invalid(t *testing.T) {
	tests := []struct {
		name        string
		config      ErrorConfig
		expectedErr string
	}{
		{
			name:        "nil percentage",
			config:      ErrorConfig{},
			expectedErr: "percentage should be greater than 0 and lesser than 100",
		},
		{
			name: "zero percentage",
			config: ErrorConfig{
				Percentage: intPtr(0),
			},
			expectedErr: "percentage should be greater than 0 and lesser than 100",
		},
		{
			name: "negative percentage",
			config: ErrorConfig{
				Percentage: intPtr(-10),
			},
			expectedErr: "percentage should be greater than 0 and lesser than 100",
		},
		{
			name: "percentage over 100",
			config: ErrorConfig{
				Percentage: intPtr(150),
			},
			expectedErr: "percentage should be greater than 0 and lesser than 100",
		},
		{
			name: "invalid latency config",
			config: ErrorConfig{
				Percentage: intPtr(50),
				LatencyConfig: &LatencyConfig{
					Min: intPtr(200),
					Max: intPtr(100), // min > max
				},
			},
			expectedErr: "invalid latency config found: min can not be greater than max",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.validate()
			if err == nil {
				t.Error("expected error for invalid config")
			} else if !strings.Contains(err.Error(), tt.expectedErr) {
				t.Errorf("expected error message to contain '%s', got '%s'", tt.expectedErr, err.Error())
			}
		})
	}
}

// Test UriConfig validation

func TestUriConfig_Validate_Valid(t *testing.T) {
	tests := []struct {
		name   string
		config UriConfig
	}{
		{
			name: "latency config only",
			config: UriConfig{
				LatencyConfig: &LatencyConfig{
					Min: intPtr(100),
					Max: intPtr(200),
				},
			},
		},
		{
			name: "errors config only",
			config: UriConfig{
				ErrorsConfig: map[string]ErrorConfig{
					"404": {
						Percentage: intPtr(25),
					},
				},
			},
		},
		{
			name: "both latency and errors config",
			config: UriConfig{
				LatencyConfig: &LatencyConfig{
					Min: intPtr(100),
					Max: intPtr(200),
				},
				ErrorsConfig: map[string]ErrorConfig{
					"404": {
						Percentage: intPtr(25),
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.validate()
			if err != nil {
				t.Errorf("expected no error for valid config, got %v", err)
			}
		})
	}
}

func TestUriConfig_Validate_Invalid(t *testing.T) {
	tests := []struct {
		name        string
		config      UriConfig
		expectedErr string
	}{
		{
			name:        "both configs nil",
			config:      UriConfig{},
			expectedErr: "latency or errors should not be both null",
		},
		{
			name: "invalid latency config",
			config: UriConfig{
				LatencyConfig: &LatencyConfig{
					Min: intPtr(200),
					Max: intPtr(100), // min > max
				},
			},
			expectedErr: "invalid latency config found: min can not be greater than max",
		},
		{
			name: "invalid status code",
			config: UriConfig{
				ErrorsConfig: map[string]ErrorConfig{
					"200": { // 2xx codes not allowed
						Percentage: intPtr(25),
					},
				},
			},
			expectedErr: "error status code should be between 400 and 599",
		},
		{
			name: "percentage sum over 100",
			config: UriConfig{
				ErrorsConfig: map[string]ErrorConfig{
					"404": {
						Percentage: intPtr(60),
					},
					"500": {
						Percentage: intPtr(50),
					},
				},
			},
			expectedErr: "the sum of all percentages should not exceed 100",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.validate()
			if err == nil {
				t.Error("expected error for invalid config")
			} else if !strings.Contains(err.Error(), tt.expectedErr) {
				t.Errorf("expected error message to contain '%s', got '%s'", tt.expectedErr, err.Error())
			}
		})
	}
}

// Test HostConfig validation

func TestHostConfig_Validate_Valid(t *testing.T) {
	tests := []struct {
		name   string
		config HostConfig
	}{
		{
			name:   "empty config",
			config: HostConfig{},
		},
		{
			name: "latency config only",
			config: HostConfig{
				LatencyConfig: &LatencyConfig{
					Min: intPtr(100),
					Max: intPtr(200),
				},
			},
		},
		{
			name: "errors config only",
			config: HostConfig{
				ErrorsConfig: map[string]ErrorConfig{
					"500": {
						Percentage: intPtr(25),
					},
				},
			},
		},
		{
			name: "uris config only",
			config: HostConfig{
				UrisConfig: map[string]UriConfig{
					"/api/v1/users": {
						LatencyConfig: &LatencyConfig{
							Min: intPtr(50),
							Max: intPtr(100),
						},
					},
				},
			},
		},
		{
			name: "all configs",
			config: HostConfig{
				LatencyConfig: &LatencyConfig{
					Min: intPtr(100),
					Max: intPtr(200),
				},
				ErrorsConfig: map[string]ErrorConfig{
					"500": {
						Percentage: intPtr(25),
					},
				},
				UrisConfig: map[string]UriConfig{
					"/api/v1/users": {
						LatencyConfig: &LatencyConfig{
							Min: intPtr(50),
							Max: intPtr(100),
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if err != nil {
				t.Errorf("expected no error for valid config, got %v", err)
			}
		})
	}
}

func TestHostConfig_Validate_Invalid(t *testing.T) {
	tests := []struct {
		name        string
		config      HostConfig
		expectedErr string
	}{
		{
			name: "invalid latency config",
			config: HostConfig{
				LatencyConfig: &LatencyConfig{
					Min: intPtr(200),
					Max: intPtr(100), // min > max
				},
			},
			expectedErr: "invalid latency config found: min can not be greater than max",
		},
		{
			name: "invalid error code - not numeric",
			config: HostConfig{
				ErrorsConfig: map[string]ErrorConfig{
					"abc": {
						Percentage: intPtr(25),
					},
				},
			},
			expectedErr: "invalid error code",
		},
		{
			name: "invalid error code - 2xx",
			config: HostConfig{
				ErrorsConfig: map[string]ErrorConfig{
					"200": {
						Percentage: intPtr(25),
					},
				},
			},
			expectedErr: "error should belong to either 4xx or 5xx classes",
		},
		{
			name: "invalid error code - 3xx",
			config: HostConfig{
				ErrorsConfig: map[string]ErrorConfig{
					"301": {
						Percentage: intPtr(25),
					},
				},
			},
			expectedErr: "error should belong to either 4xx or 5xx classes",
		},
		{
			name: "invalid error code - 6xx",
			config: HostConfig{
				ErrorsConfig: map[string]ErrorConfig{
					"600": {
						Percentage: intPtr(25),
					},
				},
			},
			expectedErr: "error should belong to either 4xx or 5xx classes",
		},
		{
			name: "invalid error config",
			config: HostConfig{
				ErrorsConfig: map[string]ErrorConfig{
					"500": {
						Percentage: intPtr(150), // > 100
					},
				},
			},
			expectedErr: "percentage should be greater than 0 and lesser than 100",
		},
		{
			name: "percentage sum over 100",
			config: HostConfig{
				ErrorsConfig: map[string]ErrorConfig{
					"500": {
						Percentage: intPtr(60),
					},
					"503": {
						Percentage: intPtr(50),
					},
				},
			},
			expectedErr: "the sum of all percentages should not exceed 100",
		},
		{
			name: "invalid URI pattern",
			config: HostConfig{
				UrisConfig: map[string]UriConfig{
					"invalid uri": { // doesn't match URI regex
						LatencyConfig: &LatencyConfig{
							Min: intPtr(50),
							Max: intPtr(100),
						},
					},
				},
			},
			expectedErr: "it doesn't match a uri pattern",
		},
		{
			name: "invalid URI config",
			config: HostConfig{
				UrisConfig: map[string]UriConfig{
					"/api/v1/users": {
						// Both configs are nil
					},
				},
			},
			expectedErr: "latency or errors should not be both null",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if err == nil {
				t.Error("expected error for invalid config")
			} else if !strings.Contains(err.Error(), tt.expectedErr) {
				t.Errorf("expected error message to contain '%s', got '%s'", tt.expectedErr, err.Error())
			}
		})
	}
}


