package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/btassone/alpine-hero/internal/config"
)

func TestGenerator_Generate(t *testing.T) {
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

	// Create a test template
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
	testConfig := &config.Config{
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
	outputFile := filepath.Join(tmpDir, "test-answers.txt")

	// Set up environment
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

	// Create generator and run test
	t.Run("generate answers file", func(t *testing.T) {
		gen := New(testConfig, outputFile)
		err := gen.Generate()
		if err != nil {
			t.Errorf("Generator.Generate() error = %v", err)
			return
		}

		// Read generated file
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

func TestValidateOutputPath(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "alpine-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer func(path string) {
		err := os.RemoveAll(path)
		if err != nil {
			t.Fatal(err)
		}
	}(tmpDir)

	// Create a subdirectory
	subDir := filepath.Join(tmpDir, "subdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}

	// Create a file to test parent directory validation
	notADir := filepath.Join(tmpDir, "not-a-dir")
	if err := os.WriteFile(notADir, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tests := []struct {
		name        string
		path        string
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

func TestGenerator_GenerateEdgeCases(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "alpine-generator-test")
	if err != nil {
		t.Fatal(err)
	}
	defer func(path string) {
		err := os.RemoveAll(path)
		if err != nil {
			t.Fatal(err)
		}
	}(tmpDir)

	// Create templates directory
	templatesDir := filepath.Join(tmpDir, "templates")
	if err := os.MkdirAll(templatesDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Test cases for different template content
	tests := []struct {
		name         string
		templateData string
		config       *config.Config
		outputPath   string
		expectedErr  bool
		validateFunc func(t *testing.T, content string) error
		errContains  string
		setupFunc    func() error
		cleanupFunc  func() error
	}{
		{
			name: "template with special characters",
			templateData: `KEYMAPOPTS="{{ .Keymap }}$special#chars"
HOSTNAMEOPTS="-n {{ .Hostname }}"
USEROPTS="-a -u -g {{ range $i, $g := .Groups }}{{if $i}},{{end}}{{$g}}{{end}} {{ .Username }}"`,
			config: &config.Config{
				Hostname: "test-host",
				Username: "test@user",
				Password: "pass!word123",
				Groups:   []string{"wheel", "docker"},
			},
			validateFunc: func(t *testing.T, content string) error {
				if !strings.Contains(content, "test@user") {
					t.Error("Generated content missing special character username")
				}
				return nil
			},
		},
		{
			name:         "template with empty groups",
			templateData: `USEROPTS="-a -u -g {{ range $i, $g := .Groups }}{{if $i}},{{end}}{{$g}}{{end}} {{ .Username }}"`,
			config: &config.Config{
				Hostname: "test-host",
				Username: "testuser",
				Groups:   []string{},
			},
			validateFunc: func(t *testing.T, content string) error {
				if !strings.Contains(content, `USEROPTS="-a -u -g  testuser"`) {
					t.Error("Generated content doesn't handle empty groups correctly")
				}
				return nil
			},
		},
		{
			name:         "template with read-only output directory",
			templateData: "HOSTNAMEOPTS=\"{{ .Hostname }}\"",
			config: &config.Config{
				Hostname: "test-host",
			},
			setupFunc: func() error {
				readOnlyDir := filepath.Join(tmpDir, "readonly")
				if err := os.MkdirAll(readOnlyDir, 0755); err != nil {
					return err
				}
				return os.Chmod(readOnlyDir, 0555)
			},
			outputPath:  filepath.Join(tmpDir, "readonly", "answers.txt"),
			expectedErr: true,
			errContains: "permission denied",
		},
		{
			name: "template with very long values",
			templateData: `HOSTNAMEOPTS="-n {{ .Hostname }}"
USEROPTS="-a -u -g {{ range $i, $g := .Groups }}{{if $i}},{{end}}{{$g}}{{end}} {{ .Username }}"`,
			config: &config.Config{
				Hostname: strings.Repeat("h", 255),
				Username: strings.Repeat("u", 255),
				Groups:   []string{strings.Repeat("g", 100)},
			},
			validateFunc: func(t *testing.T, content string) error {
				if !strings.Contains(content, strings.Repeat("h", 255)) {
					t.Error("Generated content missing long hostname")
				}
				return nil
			},
		},
		{
			name:         "template with missing environment variable",
			templateData: `ENVVAR="{{ .NonexistentVar }}"`,
			config: &config.Config{
				Hostname: "test-host",
			},
			expectedErr: true,
			errContains: "NonexistentVar",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a unique template file for this test
			templatePath := filepath.Join(templatesDir, "answers.tmpl")
			if err := os.WriteFile(templatePath, []byte(tt.templateData), 0644); err != nil {
				t.Fatal(err)
			}

			// Set up environment
			if err := os.Setenv("TEMPLATE_DIR", templatesDir); err != nil {
				t.Fatal(err)
			}
			defer func() {
				if err := os.Unsetenv("TEMPLATE_DIR"); err != nil {
					t.Fatal(err)
				}
			}()

			// Run any setup function
			if tt.setupFunc != nil {
				if err := tt.setupFunc(); err != nil {
					t.Fatal(err)
				}
			}

			// Define output path if not specified
			outputPath := tt.outputPath
			if outputPath == "" {
				outputPath = filepath.Join(tmpDir, "test-answers.txt")
			}

			// Create generator and generate file
			gen := New(tt.config, outputPath)
			err := gen.Generate()

			// Check error conditions
			if (err != nil) != tt.expectedErr {
				t.Errorf("Generate() error = %v, expectedErr %v", err, tt.expectedErr)
				return
			}

			if tt.errContains != "" && err != nil {
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("Expected error containing %q, got %v", tt.errContains, err)
				}
				return
			}

			// If no error expected, validate the content
			if !tt.expectedErr {
				content, err := os.ReadFile(outputPath)
				if err != nil {
					t.Fatal(err)
				}

				if tt.validateFunc != nil {
					if err := tt.validateFunc(t, string(content)); err != nil {
						t.Errorf("Content validation failed: %v", err)
					}
				}
			}

			// Run any cleanup function
			if tt.cleanupFunc != nil {
				if err := tt.cleanupFunc(); err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}

func TestGenerator_TemplateNotFound(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "alpine-generator-missing-template")
	if err != nil {
		t.Fatal(err)
	}
	defer func(path string) {
		err := os.RemoveAll(path)
		if err != nil {
			t.Fatal(err)
		}
	}(tmpDir)

	if err := os.Setenv("TEMPLATE_DIR", tmpDir); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Unsetenv("TEMPLATE_DIR"); err != nil {
			t.Fatal(err)
		}
	}()

	cfg := &config.Config{
		Hostname: "test-host",
		Username: "testuser",
	}

	gen := New(cfg, filepath.Join(tmpDir, "answers.txt"))
	err = gen.Generate()

	if err == nil {
		t.Error("Expected error when template file is missing")
	}

	if !strings.Contains(err.Error(), "no such file or directory") {
		t.Errorf("Expected 'no such file' error, got: %v", err)
	}
}

func TestGenerator_InvalidTemplateDir(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "alpine-generator-invalid-template")
	if err != nil {
		t.Fatal(err)
	}
	defer func(path string) {
		err := os.RemoveAll(path)
		if err != nil {
			t.Fatal(err)
		}
	}(tmpDir)

	// Create a file instead of a directory
	templatePath := filepath.Join(tmpDir, "templates")
	if err := os.WriteFile(templatePath, []byte("not a directory"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := os.Setenv("TEMPLATE_DIR", templatePath); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Unsetenv("TEMPLATE_DIR"); err != nil {
			t.Fatal(err)
		}
	}()

	cfg := &config.Config{
		Hostname: "test-host",
		Username: "testuser",
	}

	gen := New(cfg, filepath.Join(tmpDir, "answers.txt"))
	err = gen.Generate()

	if err == nil {
		t.Error("Expected error when template directory is invalid")
	}
}

func TestGenerator_OutputFilePermissions(t *testing.T) {
	// Skip if running as root since permission tests won't work
	if os.Getuid() == 0 {
		t.Skip("Skipping permission tests when running as root")
	}

	tmpDir, err := os.MkdirTemp("", "alpine-generator-permissions")
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

	templateContent := "HOSTNAME={{ .Hostname }}"
	if err := os.WriteFile(filepath.Join(templatesDir, "answers.tmpl"), []byte(templateContent), 0644); err != nil {
		t.Fatal(err)
	}

	if err := os.Setenv("TEMPLATE_DIR", templatesDir); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Unsetenv("TEMPLATE_DIR"); err != nil {
			t.Fatal(err)
		}
	}()

	outputFile := filepath.Join(tmpDir, "answers.txt")
	cfg := &config.Config{
		Hostname: "test-host",
		Username: "testuser",
	}

	gen := New(cfg, outputFile)
	if err := gen.Generate(); err != nil {
		t.Fatal(err)
	}

	// Check file permissions
	info, err := os.Stat(outputFile)
	if err != nil {
		t.Fatal(err)
	}

	// Check if permissions are 0600 (rw-------)
	if info.Mode().Perm() != 0600 {
		t.Errorf("Expected file permissions 0600, got %v", info.Mode().Perm())
	}
}
