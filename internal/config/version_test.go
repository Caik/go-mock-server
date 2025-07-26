package config

import (
	"testing"
)

func TestGetVersion(t *testing.T) {
	version := GetVersion()

	if version == "" {
		t.Error("expected version to be non-empty")
	}

	// The default version should be "dev"
	if version != "dev" {
		t.Errorf("expected version to be 'dev', got '%s'", version)
	}
}

func TestGetVersion_Consistency(t *testing.T) {
	// Test that GetVersion returns the same value on multiple calls
	version1 := GetVersion()
	version2 := GetVersion()

	if version1 != version2 {
		t.Errorf("expected consistent version, got '%s' and '%s'", version1, version2)
	}
}
