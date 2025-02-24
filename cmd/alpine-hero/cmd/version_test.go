// cmd/alpine-hero/cmd/version_test.go
package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestVersionCommand(t *testing.T) {
	// Store original version values
	origVersion := Version
	origBuildTime := BuildTime
	origCommitHash := CommitHash

	// Restore original values after test
	defer func() {
		Version = origVersion
		BuildTime = origBuildTime
		CommitHash = origCommitHash
	}()

	// Set test values
	Version = "test-version"
	BuildTime = "test-time"
	CommitHash = "test-hash"

	// Create a buffer to capture output
	var buf bytes.Buffer

	// Get the version command
	cmd := newVersionCmd()

	// Redirect command output to our buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	// Execute the command
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error executing version command: %v", err)
	}

	// Get the output
	output := buf.String()

	// Define our expected outputs
	expectedOutputs := []struct {
		content string
		errMsg  string
	}{
		{
			content: "test-version",
			errMsg:  "version information not found in output",
		},
		{
			content: "test-time",
			errMsg:  "build time not found in output",
		},
		{
			content: "test-hash",
			errMsg:  "commit hash not found in output",
		},
	}

	// Check each expected output
	for _, expected := range expectedOutputs {
		if !strings.Contains(output, expected.content) {
			t.Errorf("%s\nexpected output to contain %q\ngot: %q",
				expected.errMsg,
				expected.content,
				output)
		}
	}
}
