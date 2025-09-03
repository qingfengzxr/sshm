package config

import (
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestGetDefaultSSHConfigPath(t *testing.T) {
	tests := []struct {
		name     string
		goos     string
		expected string
	}{
		{"Linux", "linux", ".ssh/config"},
		{"macOS", "darwin", ".ssh/config"},
		{"Windows", "windows", ".ssh/config"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original GOOS
			originalGOOS := runtime.GOOS
			defer func() {
				// Note: We can't actually change runtime.GOOS at runtime
				// This test verifies the function logic with the current OS
				_ = originalGOOS
			}()

			configPath, err := GetDefaultSSHConfigPath()
			if err != nil {
				t.Fatalf("GetDefaultSSHConfigPath() error = %v", err)
			}

			if !strings.HasSuffix(configPath, tt.expected) {
				t.Errorf("Expected path to end with %q, got %q", tt.expected, configPath)
			}

			// Verify the path uses the correct separator for current OS
			expectedSeparator := string(filepath.Separator)
			if !strings.Contains(configPath, expectedSeparator) && len(configPath) > len(tt.expected) {
				t.Errorf("Path should use OS-specific separator %q, got %q", expectedSeparator, configPath)
			}
		})
	}
}

func TestGetSSHDirectory(t *testing.T) {
	sshDir, err := GetSSHDirectory()
	if err != nil {
		t.Fatalf("GetSSHDirectory() error = %v", err)
	}

	if !strings.HasSuffix(sshDir, ".ssh") {
		t.Errorf("Expected directory to end with .ssh, got %q", sshDir)
	}

	// Verify the path uses the correct separator for current OS
	expectedSeparator := string(filepath.Separator)
	if !strings.Contains(sshDir, expectedSeparator) && len(sshDir) > 4 {
		t.Errorf("Path should use OS-specific separator %q, got %q", expectedSeparator, sshDir)
	}
}

func TestEnsureSSHDirectory(t *testing.T) {
	// This test just ensures the function doesn't panic
	// and returns without error when .ssh directory already exists
	err := ensureSSHDirectory()
	if err != nil {
		t.Fatalf("ensureSSHDirectory() error = %v", err)
	}
}
