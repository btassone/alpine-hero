package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	_ "github.com/spf13/cobra"
)

func TestRootCommand(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantErr  bool
		contains string
	}{
		{
			name:     "show help",
			args:     []string{"--help"},
			wantErr:  false,
			contains: "Alpine Linux answer file generator",
		},
		{
			name:    "invalid command",
			args:    []string{"invalid"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rootCmd.SetArgs(tt.args)
			err := rootCmd.Execute()
			if (err != nil) != tt.wantErr {
				t.Errorf("rootCmd.Execute() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGenerateCommand(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "alpine-hero-test")
	if err != nil {
		t.Fatal(err)
	}
	defer func(path string) {
		err := os.RemoveAll(path)
		if err != nil {
			t.Fatal(err)
		}
	}(tmpDir)

	// Create templates directory and template file
	templatesDir := filepath.Join(tmpDir, "templates")
	if err := os.MkdirAll(templatesDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Copy the template file from the project
	templateContent := `KEYMAPOPTS="{{ .Keymap }} {{ .Keymap }}"
HOSTNAMEOPTS="-n {{ .Hostname }}"
...` // Add the rest of your template content

	if err := os.WriteFile(filepath.Join(templatesDir, "answers.tmpl"), []byte(templateContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Set the TEMPLATE_DIR environment variable
	err = os.Setenv("TEMPLATE_DIR", templatesDir)
	if err != nil {
		return
	}
	defer func() {
		err := os.Unsetenv("TEMPLATE_DIR")
		if err != nil {
			t.Fatal(err)
		}
	}()

	// Save original stderr
	oldStderr := os.Stderr
	defer func() { os.Stderr = oldStderr }()

	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "custom hostname",
			args:    []string{"generate", "--hostname", "custom-host"},
			wantErr: false,
		},
		{
			name:    "custom username and password",
			args:    []string{"generate", "--username", "custom-user", "--password", "custom-pass"},
			wantErr: false,
		},
		{
			name:    "custom groups",
			args:    []string{"generate", "--groups", "docker,wheel,users"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary output file
			tmpFile, err := os.CreateTemp("", "test-answers-*.txt")
			if err != nil {
				t.Fatal(err)
			}
			defer func(name string) {
				err := os.Remove(name)
				if err != nil {
					t.Fatal(err)
				}
			}(tmpFile.Name())

			// Create a pipe for stderr
			r, w, err := os.Pipe()
			if err != nil {
				t.Fatal(err)
			}

			// Set stderr to our pipe
			os.Stderr = w

			// Create a channel to capture async read completion
			done := make(chan bool)
			var stderrBuf bytes.Buffer

			// Start a goroutine to read from the pipe
			go func() {
				_, err := io.Copy(&stderrBuf, r)
				if err != nil {
					t.Error("Failed to read stderr:", err)
				}
				done <- true
			}()

			// Set output file for the test
			outputFile = tmpFile.Name()

			// Execute command
			rootCmd.SetArgs(tt.args)
			err = rootCmd.Execute()

			// Close the write end of the pipe
			if err := w.Close(); err != nil {
				t.Error("Failed to close pipe:", err)
			}

			// Wait for the read to complete
			<-done

			// Close the read end of the pipe
			if err := r.Close(); err != nil {
				t.Error("Failed to close pipe:", err)
			}

			// Check error status
			if (err != nil) != tt.wantErr {
				t.Errorf("generateCmd.Execute() error = %v, wantErr %v", err, tt.wantErr)
				t.Errorf("stderr output: %s", stderrBuf.String())
			}
		})
	}
}
