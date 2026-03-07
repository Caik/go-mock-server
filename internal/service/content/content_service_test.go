package content

import "testing"

func TestContentEventType_String(t *testing.T) {
	tests := []struct {
		eventType ContentEventType
		expected  string
	}{
		{Created, "CREATED"},
		{Updated, "UPDATED"},
		{Removed, "REMOVED"},
		{ContentEventType(99), ""},
	}

	for _, tt := range tests {
		got := tt.eventType.String()

		if got != tt.expected {
			t.Errorf("ContentEventType(%d).String() = %q, want %q", tt.eventType, got, tt.expected)
		}
	}
}
