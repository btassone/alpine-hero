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
