package cmd

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestGenerateCommand(t *testing.T) {
	// Create a temporary directory for test files
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
	if err := os.Setenv("TEMPLATE_DIR", templatesDir); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Unsetenv("TEMPLATE_DIR"); err != nil {
			t.Fatal(err)
		}
	}()

	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "custom hostname",
			args:    []string{"--hostname", "custom-host"},
			wantErr: false,
		},
		{
			name:    "custom username and password",
			args:    []string{"--username", "custom-user", "--password", "custom-pass"},
			wantErr: false,
		},
		{
			name:    "custom groups",
			args:    []string{"--groups", "docker,wheel,users"},
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
			defer func() {
				tmpFile.Close()
				os.Remove(tmpFile.Name())
			}()

			// Create a pipe for stderr
			r, w, err := os.Pipe()
			if err != nil {
				t.Fatal(err)
			}

			// Save original stderr
			oldStderr := os.Stderr
			os.Stderr = w
			defer func() { os.Stderr = oldStderr }()

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

			// Prepare command arguments
			args := append([]string{"generate"}, tt.args...)
			args = append(args, "--output", tmpFile.Name())
			rootCmd.SetArgs(args)

			// Execute command
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
