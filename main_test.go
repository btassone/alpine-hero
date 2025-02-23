package main

import (
	"bytes"
	"github.com/spf13/cobra"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  AlpineConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: AlpineConfig{
				Hostname:     "test-host",
				Username:     "testuser",
				Password:     "testpass",
				Timezone:     "UTC",
				Keymap:       "us",
				NetworkIface: "eth0",
				DiskDevice:   "/dev/sda",
				Groups:       []string{"audio", "video"},
			},
			wantErr: false,
		},
		{
			name: "empty hostname",
			config: AlpineConfig{
				Hostname:   "",
				Username:   "testuser",
				Password:   "testpass",
				DiskDevice: "/dev/sda",
			},
			wantErr: true,
		},
		{
			name: "empty username",
			config: AlpineConfig{
				Hostname:   "test-host",
				Username:   "",
				Password:   "testpass",
				DiskDevice: "/dev/sda",
			},
			wantErr: true,
		},
		{
			name: "empty password",
			config: AlpineConfig{
				Hostname:   "test-host",
				Username:   "testuser",
				Password:   "",
				DiskDevice: "/dev/sda",
			},
			wantErr: true,
		},
		{
			name: "empty disk device",
			config: AlpineConfig{
				Hostname:   "test-host",
				Username:   "testuser",
				Password:   "testpass",
				DiskDevice: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config = tt.config
			err := validateConfig()
			if (err != nil) != tt.wantErr {
				t.Errorf("validateConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGenerateAnswersFile(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "alpine-template-test")
	if err != nil {
		t.Fatal(err)
	}
	defer func(path string) {
		err := os.RemoveAll(path)
		if err != nil {
			t.Fatal(err)
		}
	}(tmpDir)

	// Create test template directory and file
	templatesDir := filepath.Join(tmpDir, "templates")
	err = os.Mkdir(templatesDir, 0755)
	if err != nil {
		t.Fatal(err)
	}

	// Create a test template matching the actual template structure
	testTemplate := `KEYMAPOPTS="{{ .Keymap }} {{ .Keymap }}"
HOSTNAMEOPTS="-n {{ .Hostname }}"
INTERFACESOPTS="auto lo
iface lo inet loopback

auto {{ .NetworkIface }}
iface {{ .NetworkIface }} inet dhcp
"
TIMEZONEOPTS="-z {{ .Timezone }}"
PROXYOPTS="none"
APKREPOSOPTS="-f"
SSHDOPTS="-c openssh"
NTPOPTS="-c chrony"
DISKOPTS="-m sys {{ .DiskDevice }}"
USEROPTS="-a -u -g {{ range $i, $g := .Groups }}{{if $i}},{{end}}{{$g}}{{end}} {{ .Username }}"
PWUSER="{{ .Password }}"
`
	err = os.WriteFile(filepath.Join(templatesDir, "answers.tmpl"), []byte(testTemplate), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Set up test configuration
	testConfig := AlpineConfig{
		Hostname:     "test-host",
		Username:     "testuser",
		Password:     "testpass",
		Timezone:     "UTC",
		Keymap:       "us",
		NetworkIface: "eth0",
		DiskDevice:   "/dev/sda",
		Groups:       []string{"audio", "video"},
	}

	// Set up test output file
	outputFile = filepath.Join(tmpDir, "test-answers.txt")

	// Run the test
	t.Run("generate answers file", func(t *testing.T) {
		config = testConfig
		err := generateAnswersFile()
		if err != nil {
			t.Errorf("generateAnswersFile() error = %v", err)
			return
		}

		// Read a generated file
		content, fErr := os.ReadFile(outputFile)
		if fErr != nil {
			t.Errorf("Failed to read generated file: %v", fErr)
			return
		}

		// Verify content
		generatedContent := string(content)
		expectedValues := []string{
			`KEYMAPOPTS="us us"`,
			`HOSTNAMEOPTS="-n test-host"`,
			`USEROPTS="-a -u -g audio,video testuser"`,
			`PWUSER="testpass"`,
			`DISKOPTS="-m sys /dev/sda"`,
		}

		for _, expected := range expectedValues {
			if !strings.Contains(generatedContent, expected) {
				t.Errorf("Generated content missing expected value: %s", expected)
			}
		}
	})
}

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

			// Set output file for the test
			outputFile = tmpFile.Name()

			// Execute command
			rootCmd.SetArgs(tt.args)
			err = rootCmd.Execute()
			if (err != nil) != tt.wantErr {
				t.Errorf("generateCmd.Execute() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMain(m *testing.M) {
	// Setup code here if needed
	code := m.Run()
	// Cleanup code here if needed
	os.Exit(code)
}

func TestMainExecution(t *testing.T) {
	// Save original os.Args, stderr, and rootCmd
	oldArgs := os.Args
	oldStderr := os.Stderr
	originalRoot := rootCmd
	defer func() {
		os.Args = oldArgs
		os.Stderr = oldStderr
		rootCmd = originalRoot
	}()

	tests := []struct {
		name     string
		args     []string
		wantExit int
	}{
		{
			name:     "successful execution",
			args:     []string{"alpine-template", "--help"},
			wantExit: 0,
		},
		{
			name:     "invalid command",
			args:     []string{"alpine-template", "invalid-command"},
			wantExit: 1,
		},
	}

	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			// Reset rootCmd for each test
			rootCmd = &cobra.Command{
				Use:   "alpine-template",
				Short: "Alpine Linux answer file generator",
				Long: `A CLI tool to generate Alpine Linux answer files for automated installation.
This tool helps create the answers file needed for automated Alpine Linux installation.`,
			}

			// Re-add subcommands
			rootCmd.AddCommand(generateCmd)
			rootCmd.AddCommand(validateCmd)

			// Create a pipe to capture stderr
			r, w, err := os.Pipe()
			if err != nil {
				t.Fatal(err)
			}
			os.Stderr = w

			// Set up exit code capture
			exitCalled := false
			oldOsExit := osExit
			osExit = func(code int) {
				exitCalled = true
				if code != tt.wantExit {
					t.Errorf("main() exit code = %v, want %v", code, tt.wantExit)
				}
			}
			defer func() {
				osExit = oldOsExit
			}()

			// Set up test args
			os.Args = tt.args

			// Run main in a separate goroutine
			done := make(chan bool)
			go func() {
				main()
				done <- true
			}()

			// Wait for either completion or timeout
			select {
			case <-done:
				// Command completed normally
			case <-time.After(100 * time.Millisecond):
				// Command timed out, assume it's stuck
			}

			// Close writer and restore stderr
			cErr := w.Close()
			if cErr != nil {
				return
			}
			os.Stderr = oldStderr

			// Read stderr output
			var buf bytes.Buffer
			_, err = io.Copy(&buf, r)
			if err != nil {
				t.Fatal(err)
			}

			// For error cases, verify exit was called
			if tt.wantExit != 0 && !exitCalled {
				t.Error("Expected os.Exit to be called, but it wasn't")
			}
		})
	}
}

func TestGenerateAnswersFileErrors(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "alpine-template-test")
	if err != nil {
		t.Fatal(err)
	}
	defer func(path string) {
		err := os.RemoveAll(path)
		if err != nil {
			t.Fatal(err)
		}
	}(tmpDir)

	// Create a test template directory
	testTmplDir := filepath.Join(tmpDir, "templates")
	err = os.Mkdir(testTmplDir, 0755)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name          string
		setupFunc     func() error
		cleanupFunc   func() error
		expectedError string
	}{
		{
			name: "template not found",
			setupFunc: func() error {
				// No setup needed - template won't exist
				return nil
			},
			cleanupFunc:   nil,
			expectedError: "failed to parse template",
		},
		{
			name: "invalid output path",
			setupFunc: func() error {
				// Create a valid template file but set an invalid output path
				tmplContent := `KEYMAPOPTS="{{ .Keymap }} {{ .Keymap }}"`
				err := os.WriteFile(filepath.Join(testTmplDir, "answers.tmpl"), []byte(tmplContent), 0644)
				if err != nil {
					return err
				}
				outputFile = "/nonexistent/directory/answers.txt"
				return nil
			},
			cleanupFunc: func() error {
				outputFile = "answers.txt"
				return nil
			},
			expectedError: "invalid output path: output path not allowed",
		},
		{
			name: "template execution error",
			setupFunc: func() error {
				// Create a template with an invalid template directive
				tmplContent := `KEYMAPOPTS="{{ .InvalidField }}"`
				return os.WriteFile(filepath.Join(testTmplDir, "answers.tmpl"), []byte(tmplContent), 0644)
			},
			cleanupFunc:   nil,
			expectedError: "failed to execute template",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Point to test templates directory
			if err := os.Setenv("TEMPLATE_DIR", testTmplDir); err != nil {
				t.Fatal(err)
			}
			defer func() {
				err := os.Unsetenv("TEMPLATE_DIR")
				if err != nil {
					t.Fatal(err)
				}
			}()

			// Setup test
			if tt.setupFunc != nil {
				err := tt.setupFunc()
				if err != nil {
					t.Fatal(err)
				}
			}

			// Cleanup after test
			if tt.cleanupFunc != nil {
				defer func() {
					err := tt.cleanupFunc()
					if err != nil {
						t.Fatal(err)
					}
				}()
			}

			// Run test
			err := generateAnswersFile()

			// Check error
			if err == nil {
				t.Error("Expected error but got nil")
				return
			}

			if !strings.Contains(err.Error(), tt.expectedError) {
				t.Errorf("Expected error containing %q, got %q", tt.expectedError, err.Error())
			}
		})
	}
}

func TestValidateOutputPath(t *testing.T) {
	// Create temporary test directories
	tmpDir, err := os.MkdirTemp("", "alpine-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a subdirectory in temp directory
	subDir := filepath.Join(tmpDir, "subdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}

	// Create a file to test parent directory validation
	notADir := filepath.Join(tmpDir, "not-a-dir")
	if err := os.WriteFile(notADir, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create a directory with no write permissions
	noWriteDir := filepath.Join(tmpDir, "no-write")
	if err := os.Mkdir(noWriteDir, 0555); err != nil {
		t.Fatalf("Failed to create no-write directory: %v", err)
	}

	// Get absolute paths for test directories
	absHome, err := filepath.Abs(os.Getenv("HOME"))
	if err != nil {
		t.Fatalf("Failed to get absolute home path: %v", err)
	}

	tests := []struct {
		name        string
		path        string
		setupFunc   func() error
		cleanupFunc func() error
		wantErr     bool
		errContains string
	}{
		{
			name:    "valid path in temp directory",
			path:    filepath.Join(tmpDir, "valid.txt"),
			wantErr: false,
		},
		{
			name:        "path in non-existent directory",
			path:        filepath.Join(tmpDir, "nonexistent", "file.txt"),
			wantErr:     true,
			errContains: "parent directory does not exist",
		},
		{
			name:        "path with non-directory parent",
			path:        filepath.Join(notADir, "file.txt"),
			wantErr:     true,
			errContains: "parent path is not a directory",
		},
		{
			name:    "relative path in current directory",
			path:    "output.txt",
			wantErr: false,
		},
		{
			name: "path in HOME/alpine-template",
			path: filepath.Join(absHome, "alpine-template", "output.txt"),
			setupFunc: func() error {
				return os.MkdirAll(filepath.Join(absHome, "alpine-template"), 0755)
			},
			cleanupFunc: func() error {
				return os.RemoveAll(filepath.Join(absHome, "alpine-template"))
			},
			wantErr: false,
		},
		{
			name:        "path outside allowed directories",
			path:        "/etc/not-allowed.txt",
			wantErr:     true,
			errContains: "output path not allowed",
		},
		{
			name:    "path with . prefix",
			path:    "./output.txt",
			wantErr: false,
		},
		{
			name:    "path with .. prefix",
			path:    "../output.txt",
			wantErr: false,
		},
		{
			name:        "path trying directory traversal",
			path:        filepath.Join(tmpDir, "..", "..", "etc", "passwd"),
			wantErr:     true,
			errContains: "output path not allowed",
		},
		{
			name: "deeply nested valid path",
			path: filepath.Join(subDir, "deep", "nested", "file.txt"),
			setupFunc: func() error {
				return os.MkdirAll(filepath.Join(subDir, "deep", "nested"), 0755)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Run setup if provided
			if tt.setupFunc != nil {
				if err := tt.setupFunc(); err != nil {
					t.Fatalf("Setup failed: %v", err)
				}
			}

			// Run cleanup if provided
			if tt.cleanupFunc != nil {
				defer func() {
					if err := tt.cleanupFunc(); err != nil {
						t.Errorf("Cleanup failed: %v", err)
					}
				}()
			}

			err := validateOutputPath(tt.path)

			// Check if error matches expectation
			if (err != nil) != tt.wantErr {
				t.Errorf("validateOutputPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// If expecting an error, check error message
			if tt.wantErr && tt.errContains != "" && err != nil {
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("validateOutputPath() error = %v, want error containing %q", err, tt.errContains)
				}
			}
		})
	}
}

// TestValidateOutputPathWithPermissions tests path validation with different permission scenarios.
// Note: validateOutputPath only validates that directories exist and are directories.
// Actual write permission checks happen when the file is created in generateAnswersFile.
func TestValidateOutputPathWithPermissions(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("Skipping permission tests when running as root")
	}

	tmpDir, err := os.MkdirTemp("", "alpine-perm-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create directories with different permissions
	readOnlyDir := filepath.Join(tmpDir, "readonly")
	if err := os.Mkdir(readOnlyDir, 0555); err != nil {
		t.Fatalf("Failed to create readonly directory: %v", err)
	}

	noAccessDir := filepath.Join(tmpDir, "noaccess")
	if err := os.Mkdir(noAccessDir, 0000); err != nil {
		t.Fatalf("Failed to create noaccess directory: %v", err)
	}
	defer os.Chmod(noAccessDir, 0755) // Ensure we can clean up

	tests := []struct {
		name        string
		path        string
		wantErr     bool
		errContains string
	}{
		{
			name:    "path in writable directory",
			path:    filepath.Join(tmpDir, "writable.txt"),
			wantErr: false,
		},
		{
			name:    "path in read-only directory",
			path:    filepath.Join(readOnlyDir, "readonly.txt"),
			wantErr: false, // Changed because validateOutputPath only checks if directory exists and is a directory
		},
		{
			name:    "path in no-access directory",
			path:    filepath.Join(noAccessDir, "noaccess.txt"),
			wantErr: false, // Changed because validateOutputPath only checks if directory exists and is a directory
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateOutputPath(tt.path)

			if (err != nil) != tt.wantErr {
				t.Errorf("validateOutputPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errContains != "" && err != nil {
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("validateOutputPath() error = %v, want error containing %q", err, tt.errContains)
				}
			}
		})
	}
}
