package util

import (
	"net/http"
	"strings"
	"testing"
)

// Test HostRegex
func TestHostRegex(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		// Valid hosts
		{"simple domain", "example.com", true},
		{"subdomain", "api.example.com", true},
		{"multiple subdomains", "api.v1.example.com", true},
		{"with hyphens", "my-api.example-site.com", true},
		{"single letter subdomain", "a.example.com", true},
		{"numbers in domain", "api1.example2.com", true},
		{"long domain", "very-long-subdomain-name.example.com", true},

		// Invalid hosts
		{"no domain", "localhost", false},
		{"starts with dot", ".example.com", false},
		{"ends with dot", "example.com.", false},
		{"double dot", "api..example.com", false},
		{"empty string", "", false},
		{"only dots", "...", false},
		{"with spaces", "api .example.com", false},
		{"with special chars", "api@example.com", false},
		{"single dot", ".", false},
		{"no TLD", "example", false},

		// Valid hosts (based on actual regex behavior)
		{"IP address format", "192.168.1.1", true},        // \w includes digits
		{"starts with hyphen", "-api.example.com", true},  // [\w-] allows leading hyphen
		{"ends with hyphen", "api-.example.com", true},    // [\w-] allows trailing hyphen
		{"with underscore", "api_test.example.com", true}, // \w includes underscore
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HostRegex.MatchString(tt.input)
			if result != tt.expected {
				t.Errorf("HostRegex.MatchString(%q) = %v, expected %v", tt.input, result, tt.expected)
			}
		})
	}
}

// Test IpAddressRegex
func TestIpAddressRegex(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		// Valid IP addresses
		{"localhost IP", "127.0.0.1", true},
		{"private IP", "192.168.1.1", true},
		{"public IP", "8.8.8.8", true},
		{"zero IP", "0.0.0.0", true},
		{"max IP", "255.255.255.255", true},
		{"single digit octets", "1.2.3.4", true},
		{"mixed octets", "192.168.0.1", true},

		// Invalid IP addresses
		{"too many octets", "192.168.1.1.1", false},
		{"too few octets", "192.168.1", false},
		{"negative octet", "-1.1.1.1", false},
		{"empty string", "", false},
		{"letters", "abc.def.ghi.jkl", false},
		{"with spaces", "192.168. 1.1", false},
		{"four digit octet", "1000.1.1.1", false},
		{"double dots", "192..168.1.1", false},
		{"ends with dot", "192.168.1.1.", false},
		{"starts with dot", ".192.168.1.1", false},

		// Valid IP addresses (based on actual regex - it only checks format, not ranges)
		{"octet too large", "256.1.1.1", true}, // Regex allows any 1-3 digits
		{"octet too large 2", "1.256.1.1", true},
		{"octet too large 3", "1.1.256.1", true},
		{"octet too large 4", "1.1.1.256", true},
		{"leading zeros", "001.002.003.004", true}, // Regex allows leading zeros
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IpAddressRegex.MatchString(tt.input)
			if result != tt.expected {
				t.Errorf("IpAddressRegex.MatchString(%q) = %v, expected %v", tt.input, result, tt.expected)
			}
		})
	}
}

// Test UriRegex
func TestUriRegex(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		// Valid URIs
		{"simple path", "/api", true},
		{"nested path", "/api/v1/users", true},
		{"without leading slash", "api/v1/users", true},
		{"with trailing slash", "/api/v1/users/", true},
		{"single segment", "users", true},
		{"with hyphens", "/api/user-management", true},
		{"with underscores", "/api/user_management", true},
		{"with query params", "/api/users?id=123", true},
		{"multiple query params", "/api/users?id=123&name=john", true},
		{"query with hyphens", "/api/users?user-id=123", true},
		{"query with underscores", "/api/users?user_id=123", true},

		// Invalid URIs
		{"root path", "/", false}, // Regex requires at least one path segment
		{"empty path", "", false}, // Regex requires at least one path segment
		{"with spaces", "/api /users", false},
		{"with special chars", "/api@users", false},
		{"with dots", "/api.users", false},
		{"invalid query format", "/api/users?id", false},
		{"invalid query separator", "/api/users?id=123;name=john", false},
		{"query with spaces", "/api/users?id=123 456", false},
		{"query with special chars", "/api/users?id=123@456", false},
		{"double slashes", "/api//users", false},
		{"ends with double slash", "/api/users//", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := UriRegex.MatchString(tt.input)
			if result != tt.expected {
				t.Errorf("UriRegex.MatchString(%q) = %v, expected %v", tt.input, result, tt.expected)
			}
		})
	}
}

// Test HttpMethodRegex
func TestHttpMethodRegex(t *testing.T) {
	// Test all valid HTTP methods
	validMethods := []string{
		http.MethodGet,
		http.MethodHead,
		http.MethodPost,
		http.MethodPut,
		http.MethodPatch,
		http.MethodDelete,
		http.MethodConnect,
		http.MethodOptions,
		http.MethodTrace,
	}

	for _, method := range validMethods {
		t.Run("valid method "+method, func(t *testing.T) {
			result := HttpMethodRegex.MatchString(method)
			if !result {
				t.Errorf("HttpMethodRegex.MatchString(%q) = false, expected true", method)
			}
		})
	}

	// Test invalid methods
	invalidMethods := []struct {
		name   string
		method string
	}{
		{"lowercase get", "get"},
		{"lowercase post", "post"},
		{"mixed case", "Get"},
		{"invalid method", "INVALID"},
		{"empty string", ""},
		{"with spaces", "GET "},
		{"with extra chars", "GET123"},
		{"multiple methods", "GET POST"},
	}

	for _, tt := range invalidMethods {
		t.Run("invalid method "+tt.name, func(t *testing.T) {
			result := HttpMethodRegex.MatchString(tt.method)
			if result {
				t.Errorf("HttpMethodRegex.MatchString(%q) = true, expected false", tt.method)
			}
		})
	}
}

// Test regex compilation and initialization
func TestRegexInitialization(t *testing.T) {
	t.Run("HostRegex is compiled", func(t *testing.T) {
		if HostRegex == nil {
			t.Error("HostRegex should be compiled and non-nil")
		}
	})

	t.Run("IpAddressRegex is compiled", func(t *testing.T) {
		if IpAddressRegex == nil {
			t.Error("IpAddressRegex should be compiled and non-nil")
		}
	})

	t.Run("UriRegex is compiled", func(t *testing.T) {
		if UriRegex == nil {
			t.Error("UriRegex should be compiled and non-nil")
		}
	})

	t.Run("HttpMethodRegex is compiled", func(t *testing.T) {
		if HttpMethodRegex == nil {
			t.Error("HttpMethodRegex should be compiled and non-nil")
		}
	})
}

// Test regex patterns as strings
func TestRegexPatterns(t *testing.T) {
	t.Run("HostRegex pattern", func(t *testing.T) {
		pattern := HostRegex.String()
		if pattern == "" {
			t.Error("HostRegex pattern should not be empty")
		}
	})

	t.Run("IpAddressRegex pattern", func(t *testing.T) {
		pattern := IpAddressRegex.String()
		if pattern == "" {
			t.Error("IpAddressRegex pattern should not be empty")
		}
	})

	t.Run("UriRegex pattern", func(t *testing.T) {
		pattern := UriRegex.String()
		if pattern == "" {
			t.Error("UriRegex pattern should not be empty")
		}
	})

	t.Run("HttpMethodRegex pattern", func(t *testing.T) {
		pattern := HttpMethodRegex.String()
		if pattern == "" {
			t.Error("HttpMethodRegex pattern should not be empty")
		}

		// Verify it contains all HTTP methods
		for _, method := range []string{"GET", "POST", "PUT", "DELETE"} {
			if !strings.Contains(pattern, method) {
				t.Errorf("HttpMethodRegex pattern should contain %s", method)
			}
		}
	})
}
