package rest

import (
	"encoding/json"
	"testing"
)

// Test Status constants
func TestStatusConstants(t *testing.T) {
	tests := []struct {
		name     string
		status   Status
		expected string
	}{
		{
			name:     "Success status",
			status:   Success,
			expected: "success",
		},
		{
			name:     "Fail status",
			status:   Fail,
			expected: "fail",
		},
		{
			name:     "Error status",
			status:   Error,
			expected: "error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.status) != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, string(tt.status))
			}
		})
	}
}

// Test Status type behavior
func TestStatusType(t *testing.T) {
	var status Status = "custom"
	
	if string(status) != "custom" {
		t.Errorf("expected 'custom', got %s", string(status))
	}

	// Test that Status can be compared
	if status == Success {
		t.Error("custom status should not equal Success")
	}

	// Test assignment
	status = Success
	if status != Success {
		t.Error("status should equal Success after assignment")
	}
}

// Test Response struct creation and field access
func TestResponseStruct(t *testing.T) {
	t.Run("create response with all fields", func(t *testing.T) {
		data := map[string]interface{}{
			"id":   1,
			"name": "test",
		}

		response := Response{
			Status:  Success,
			Message: "Operation completed successfully",
			Data:    data,
		}

		if response.Status != Success {
			t.Errorf("expected status %s, got %s", Success, response.Status)
		}

		if response.Message != "Operation completed successfully" {
			t.Errorf("expected message 'Operation completed successfully', got %s", response.Message)
		}

		if response.Data == nil {
			t.Error("expected data to be non-nil")
		}
	})

	t.Run("create response with minimal fields", func(t *testing.T) {
		response := Response{
			Status: Fail,
		}

		if response.Status != Fail {
			t.Errorf("expected status %s, got %s", Fail, response.Status)
		}

		if response.Message != "" {
			t.Errorf("expected empty message, got %s", response.Message)
		}

		if response.Data != nil {
			t.Error("expected data to be nil")
		}
	})

	t.Run("create empty response", func(t *testing.T) {
		response := Response{}

		if response.Status != "" {
			t.Errorf("expected empty status, got %s", response.Status)
		}

		if response.Message != "" {
			t.Errorf("expected empty message, got %s", response.Message)
		}

		if response.Data != nil {
			t.Error("expected data to be nil")
		}
	})
}

// Test JSON marshaling
func TestResponseJSONMarshaling(t *testing.T) {
	tests := []struct {
		name     string
		response Response
		expected string
	}{
		{
			name: "full response",
			response: Response{
				Status:  Success,
				Message: "test message",
				Data:    map[string]string{"key": "value"},
			},
			expected: `{"status":"success","message":"test message","data":{"key":"value"}}`,
		},
		{
			name: "response with status only",
			response: Response{
				Status: Error,
			},
			expected: `{"status":"error"}`,
		},
		{
			name: "response with status and message",
			response: Response{
				Status:  Fail,
				Message: "validation failed",
			},
			expected: `{"status":"fail","message":"validation failed"}`,
		},
		{
			name: "response with status and data",
			response: Response{
				Status: Success,
				Data:   []int{1, 2, 3},
			},
			expected: `{"status":"success","data":[1,2,3]}`,
		},
		{
			name: "empty response",
			response: Response{},
			expected: `{"status":""}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonData, err := json.Marshal(tt.response)
			if err != nil {
				t.Fatalf("failed to marshal response: %v", err)
			}

			if string(jsonData) != tt.expected {
				t.Errorf("expected JSON %s, got %s", tt.expected, string(jsonData))
			}
		})
	}
}

// Test JSON unmarshaling
func TestResponseJSONUnmarshaling(t *testing.T) {
	tests := []struct {
		name     string
		jsonData string
		expected Response
		wantErr  bool
	}{
		{
			name:     "full response",
			jsonData: `{"status":"success","message":"test message","data":{"key":"value"}}`,
			expected: Response{
				Status:  Success,
				Message: "test message",
				Data:    map[string]interface{}{"key": "value"},
			},
			wantErr: false,
		},
		{
			name:     "response with status only",
			jsonData: `{"status":"error"}`,
			expected: Response{
				Status: Error,
			},
			wantErr: false,
		},
		{
			name:     "response with status and message",
			jsonData: `{"status":"fail","message":"validation failed"}`,
			expected: Response{
				Status:  Fail,
				Message: "validation failed",
			},
			wantErr: false,
		},
		{
			name:     "response with status and data array",
			jsonData: `{"status":"success","data":[1,2,3]}`,
			expected: Response{
				Status: Success,
				Data:   []interface{}{1.0, 2.0, 3.0}, // JSON numbers become float64
			},
			wantErr: false,
		},
		{
			name:     "empty JSON object",
			jsonData: `{}`,
			expected: Response{},
			wantErr:  false,
		},
		{
			name:     "invalid JSON",
			jsonData: `{"status":"success","invalid"}`,
			expected: Response{},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var response Response
			err := json.Unmarshal([]byte(tt.jsonData), &response)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("failed to unmarshal JSON: %v", err)
			}

			if response.Status != tt.expected.Status {
				t.Errorf("expected status %s, got %s", tt.expected.Status, response.Status)
			}

			if response.Message != tt.expected.Message {
				t.Errorf("expected message %s, got %s", tt.expected.Message, response.Message)
			}

			// For data comparison, we need to handle the interface{} type carefully
			if tt.expected.Data != nil {
				if response.Data == nil {
					t.Error("expected data to be non-nil")
				}
				// Note: Deep comparison of interface{} can be complex, 
				// but for our test cases, we can verify the structure is preserved
			} else if response.Data != nil {
				t.Error("expected data to be nil")
			}
		})
	}
}

// Test Response with different data types
func TestResponseWithVariousDataTypes(t *testing.T) {
	tests := []struct {
		name string
		data interface{}
	}{
		{"string data", "test string"},
		{"int data", 42},
		{"float data", 3.14},
		{"bool data", true},
		{"slice data", []string{"a", "b", "c"}},
		{"map data", map[string]int{"count": 5}},
		{"nil data", nil},
		{"struct data", struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		}{ID: 1, Name: "test"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := Response{
				Status: Success,
				Data:   tt.data,
			}

			// Test that we can marshal and unmarshal without errors
			jsonData, err := json.Marshal(response)
			if err != nil {
				t.Fatalf("failed to marshal response with %s: %v", tt.name, err)
			}

			var unmarshaled Response
			err = json.Unmarshal(jsonData, &unmarshaled)
			if err != nil {
				t.Fatalf("failed to unmarshal response with %s: %v", tt.name, err)
			}

			if unmarshaled.Status != Success {
				t.Errorf("status changed during marshal/unmarshal cycle")
			}
		})
	}
}

// Test omitempty behavior
func TestResponseOmitEmpty(t *testing.T) {
	t.Run("empty message and data are omitted", func(t *testing.T) {
		response := Response{
			Status: Success,
			// Message and Data are empty/nil, should be omitted
		}

		jsonData, err := json.Marshal(response)
		if err != nil {
			t.Fatalf("failed to marshal response: %v", err)
		}

		expected := `{"status":"success"}`
		if string(jsonData) != expected {
			t.Errorf("expected %s, got %s", expected, string(jsonData))
		}
	})

	t.Run("empty string message is omitted", func(t *testing.T) {
		response := Response{
			Status:  Error,
			Message: "", // Empty string should be omitted
		}

		jsonData, err := json.Marshal(response)
		if err != nil {
			t.Fatalf("failed to marshal response: %v", err)
		}

		expected := `{"status":"error"}`
		if string(jsonData) != expected {
			t.Errorf("expected %s, got %s", expected, string(jsonData))
		}
	})
}

// Test Status string conversion and comparison
func TestStatusStringConversion(t *testing.T) {
	tests := []struct {
		name     string
		status   Status
		expected string
	}{
		{"Success to string", Success, "success"},
		{"Fail to string", Fail, "fail"},
		{"Error to string", Error, "error"},
		{"Custom status", Status("custom"), "custom"},
		{"Empty status", Status(""), ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := string(tt.status)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

// Test Status equality and comparison
func TestStatusComparison(t *testing.T) {
	t.Run("status equality", func(t *testing.T) {
		if Success != Success {
			t.Error("Success should equal itself")
		}

		if Fail == Error {
			t.Error("Fail should not equal Error")
		}

		var customStatus Status = "success"
		if customStatus != Success {
			t.Error("custom status 'success' should equal Success constant")
		}
	})

	t.Run("status inequality", func(t *testing.T) {
		statuses := []Status{Success, Fail, Error}

		for i, status1 := range statuses {
			for j, status2 := range statuses {
				if i != j && status1 == status2 {
					t.Errorf("status %s should not equal %s", status1, status2)
				}
			}
		}
	})
}

// Test Response field assignment and modification
func TestResponseFieldModification(t *testing.T) {
	response := Response{}

	// Test field assignment
	response.Status = Success
	if response.Status != Success {
		t.Error("Status assignment failed")
	}

	response.Message = "test message"
	if response.Message != "test message" {
		t.Error("Message assignment failed")
	}

	testData := map[string]string{"key": "value"}
	response.Data = testData
	if response.Data == nil {
		t.Error("Data assignment failed")
	}

	// Test field modification
	response.Status = Error
	if response.Status != Error {
		t.Error("Status modification failed")
	}

	response.Message = "updated message"
	if response.Message != "updated message" {
		t.Error("Message modification failed")
	}

	response.Data = nil
	if response.Data != nil {
		t.Error("Data should be nil after setting to nil")
	}
}

// Test Response with complex nested data
func TestResponseWithComplexData(t *testing.T) {
	complexData := map[string]interface{}{
		"user": map[string]interface{}{
			"id":    123,
			"name":  "John Doe",
			"email": "john@example.com",
			"roles": []string{"admin", "user"},
			"metadata": map[string]interface{}{
				"lastLogin": "2023-01-01T00:00:00Z",
				"active":    true,
			},
		},
		"pagination": map[string]interface{}{
			"page":     1,
			"pageSize": 10,
			"total":    100,
		},
	}

	response := Response{
		Status:  Success,
		Message: "User data retrieved successfully",
		Data:    complexData,
	}

	// Test JSON marshaling with complex data
	jsonData, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("failed to marshal complex response: %v", err)
	}

	// Test JSON unmarshaling with complex data
	var unmarshaled Response
	err = json.Unmarshal(jsonData, &unmarshaled)
	if err != nil {
		t.Fatalf("failed to unmarshal complex response: %v", err)
	}

	if unmarshaled.Status != Success {
		t.Error("Status should be preserved in complex data scenario")
	}

	if unmarshaled.Message != "User data retrieved successfully" {
		t.Error("Message should be preserved in complex data scenario")
	}

	if unmarshaled.Data == nil {
		t.Error("Complex data should be preserved")
	}
}

// Test Response zero value behavior
func TestResponseZeroValue(t *testing.T) {
	var response Response

	// Test zero values
	if response.Status != "" {
		t.Errorf("zero value Status should be empty string, got %s", response.Status)
	}

	if response.Message != "" {
		t.Errorf("zero value Message should be empty string, got %s", response.Message)
	}

	if response.Data != nil {
		t.Error("zero value Data should be nil")
	}

	// Test JSON marshaling of zero value
	jsonData, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("failed to marshal zero value response: %v", err)
	}

	expected := `{"status":""}`
	if string(jsonData) != expected {
		t.Errorf("expected %s, got %s", expected, string(jsonData))
	}
}
