package config

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	cfg := New()

	// Test that New() returns a config with expected default values
	expectedConfig := &Config{
		Hostname:     "alpinehost",
		Username:     "alpine",
		Password:     "changeme",
		Timezone:     "UTC",
		Keymap:       "us",
		NetworkIface: "eth0",
		DiskDevice:   "/dev/mmcblk0",
		Groups:       []string{"audio", "video", "netdev"},
	}

	if !reflect.DeepEqual(cfg, expectedConfig) {
		t.Errorf("New() returned unexpected default values\ngot: %+v\nwant: %+v", cfg, expectedConfig)
	}
}

func TestConfig_Validate(t *testing.T) {
	// Create temporary test directory and files
	tmpDir, err := os.MkdirTemp("", "alpine-config-test")
	if err != nil {
		t.Fatal(err)
	}
	defer func(path string) {
		err := os.RemoveAll(path)
		if err != nil {
			t.Fatal(err)
		}
	}(tmpDir)

	// Create test SSH key files
	validKeyPath := filepath.Join(tmpDir, "valid.pub")
	invalidKeyPath := filepath.Join(tmpDir, "invalid.pub")
	nonexistentKeyPath := filepath.Join(tmpDir, "nonexistent.pub")

	err = os.WriteFile(validKeyPath, []byte("ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC0g+ZTxC test@example.com"), 0644)
	if err != nil {
		t.Fatal(err)
	}
	err = os.WriteFile(invalidKeyPath, []byte("invalid-key-data"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name        string
		config      *Config
		wantErr     bool
		errContains string
	}{
		{
			name: "valid minimal config",
			config: &Config{
				Hostname:   "test-host",
				Username:   "testuser",
				Password:   "testpass",
				DiskDevice: "/dev/sda",
			},
			wantErr: false,
		},
		{
			name: "valid full config",
			config: &Config{
				Hostname:     "test-host",
				Username:     "testuser",
				Password:     "testpass",
				Timezone:     "UTC",
				Keymap:       "us",
				NetworkIface: "eth0",
				DiskDevice:   "/dev/sda",
				Groups:       []string{"audio", "video"},
				SSHKey:       validKeyPath,
			},
			wantErr: false,
		},
		{
			name: "empty hostname",
			config: &Config{
				Username:   "testuser",
				Password:   "testpass",
				DiskDevice: "/dev/sda",
			},
			wantErr:     true,
			errContains: "hostname cannot be empty",
		},
		{
			name: "empty username",
			config: &Config{
				Hostname:   "test-host",
				Password:   "testpass",
				DiskDevice: "/dev/sda",
			},
			wantErr:     true,
			errContains: "username cannot be empty",
		},
		{
			name: "empty password",
			config: &Config{
				Hostname:   "test-host",
				Username:   "testuser",
				DiskDevice: "/dev/sda",
			},
			wantErr:     true,
			errContains: "password cannot be empty",
		},
		{
			name: "empty disk device",
			config: &Config{
				Hostname: "test-host",
				Username: "testuser",
				Password: "testpass",
			},
			wantErr:     true,
			errContains: "disk device cannot be empty",
		},
		{
			name: "valid ssh key",
			config: &Config{
				Hostname:   "test-host",
				Username:   "testuser",
				Password:   "testpass",
				DiskDevice: "/dev/sda",
				SSHKey:     validKeyPath,
			},
			wantErr: false,
		},
		{
			name: "invalid ssh key format",
			config: &Config{
				Hostname:   "test-host",
				Username:   "testuser",
				Password:   "testpass",
				DiskDevice: "/dev/sda",
				SSHKey:     invalidKeyPath,
			},
			wantErr:     true,
			errContains: "invalid SSH public key format",
		},
		{
			name: "nonexistent ssh key file",
			config: &Config{
				Hostname:   "test-host",
				Username:   "testuser",
				Password:   "testpass",
				DiskDevice: "/dev/sda",
				SSHKey:     nonexistentKeyPath,
			},
			wantErr:     true,
			errContains: "failed to read SSH key file",
		},
		{
			name: "valid config with special characters",
			config: &Config{
				Hostname:   "test-host-123",
				Username:   "test_user.123",
				Password:   "test@pass!123",
				DiskDevice: "/dev/nvme0n1p1",
			},
			wantErr: false,
		},
		{
			name: "valid config with multiple groups",
			config: &Config{
				Hostname:   "test-host",
				Username:   "testuser",
				Password:   "testpass",
				DiskDevice: "/dev/sda",
				Groups:     []string{"audio", "video", "docker", "wheel", "users"},
			},
			wantErr: false,
		},
		{
			name: "valid config with ed25519 key",
			config: &Config{
				Hostname:   "test-host",
				Username:   "testuser",
				Password:   "testpass",
				DiskDevice: "/dev/sda",
				SSHKey:     validKeyPath, // We'll create this with an ed25519 key
			},
			wantErr: false,
		},
	}

	// Create an ed25519 test key for the last test case
	err = os.WriteFile(validKeyPath, []byte("ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIKrExpXw7vJ4dBU test@example.com"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()

			// Check if error matches expectation
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// If expecting an error, check the error message contains expected string
			if tt.wantErr && tt.errContains != "" && err != nil {
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("Config.Validate() error = %v, want error containing %q", err, tt.errContains)
				}
			}
		})
	}
}

func TestConfig_ValidateWithPermissions(t *testing.T) {
	// Skip this test if running as root since root can bypass permissions
	if os.Getuid() == 0 {
		t.Skip("Skipping permission tests when running as root")
	}

	// Create our temporary test directory
	tmpDir, err := os.MkdirTemp("", "alpine-config-perms-test")
	if err != nil {
		t.Fatal(err)
	}

	// Ensure cleanup happens even if the test fails
	defer func() {
		// Make sure we can delete everything by restoring write permissions
		err := os.Chmod(tmpDir, 0755)
		if err != nil {
			return
		}
		lErr := os.RemoveAll(tmpDir)
		if lErr != nil {
			return
		}
	}()

	// Create the directory for our read-only test
	readOnlyDir := filepath.Join(tmpDir, "readonly")
	if err := os.MkdirAll(readOnlyDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create our test SSH key file while we still have write permissions
	readOnlyKey := filepath.Join(readOnlyDir, "readonly.pub")
	keyContent := "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC0g+ZTxC test@example.com"
	if err := os.WriteFile(readOnlyKey, []byte(keyContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Now that the file exists, make the directory read-only
	if err := os.Chmod(readOnlyDir, 0555); err != nil {
		t.Fatal(err)
	}

	// Run our test cases
	tests := []struct {
		name        string
		config      *Config
		wantErr     bool
		errContains string
	}{
		{
			name: "ssh key in read-only directory",
			config: &Config{
				Hostname:   "test-host",
				Username:   "testuser",
				Password:   "testpass",
				DiskDevice: "/dev/sda",
				SSHKey:     readOnlyKey,
			},
			wantErr: false, // Should succeed because we only need read access
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errContains != "" && err != nil {
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("Config.Validate() error = %v, want error containing %q",
						err, tt.errContains)
				}
			}
		})
	}
}

// TestConfig_DefaultValues ensures that a newly created Config has the expected default values
func TestConfig_DefaultValues(t *testing.T) {
	cfg := New()

	tests := []struct {
		name     string
		got      interface{}
		want     interface{}
		fieldMsg string
	}{
		{
			name:     "default hostname",
			got:      cfg.Hostname,
			want:     "alpinehost",
			fieldMsg: "Hostname field",
		},
		{
			name:     "default username",
			got:      cfg.Username,
			want:     "alpine",
			fieldMsg: "Username field",
		},
		{
			name:     "default password",
			got:      cfg.Password,
			want:     "changeme",
			fieldMsg: "Password field",
		},
		{
			name:     "default timezone",
			got:      cfg.Timezone,
			want:     "UTC",
			fieldMsg: "Timezone field",
		},
		{
			name:     "default keymap",
			got:      cfg.Keymap,
			want:     "us",
			fieldMsg: "Keymap field",
		},
		{
			name:     "default network interface",
			got:      cfg.NetworkIface,
			want:     "eth0",
			fieldMsg: "NetworkIface field",
		},
		{
			name:     "default disk device",
			got:      cfg.DiskDevice,
			want:     "/dev/mmcblk0",
			fieldMsg: "DiskDevice field",
		},
		{
			name:     "default groups",
			got:      cfg.Groups,
			want:     []string{"audio", "video", "netdev"},
			fieldMsg: "Groups field",
		},
		{
			name:     "default ssh key",
			got:      cfg.SSHKey,
			want:     "",
			fieldMsg: "SSHKey field",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !reflect.DeepEqual(tt.got, tt.want) {
				t.Errorf("%s = %v, want %v", tt.fieldMsg, tt.got, tt.want)
			}
		})
	}
}
