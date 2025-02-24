package config

import (
	"fmt"
	"os"
	"strings"
)

// Config holds the Alpine Linux installation configuration
type Config struct {
	Hostname     string
	Username     string
	Password     string
	Timezone     string
	Keymap       string
	NetworkIface string
	DiskDevice   string
	Groups       []string
	SSHKey       string
}

// New creates a new Config with default values
func New() *Config {
	return &Config{
		Hostname:     "alpinehost",
		Username:     "alpine",
		Password:     "changeme",
		Timezone:     "UTC",
		Keymap:       "us",
		NetworkIface: "eth0",
		DiskDevice:   "/dev/mmcblk0",
		Groups:       []string{"audio", "video", "netdev"},
	}
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.Hostname == "" {
		return fmt.Errorf("hostname cannot be empty")
	}
	if c.Username == "" {
		return fmt.Errorf("username cannot be empty")
	}
	if c.Password == "" {
		return fmt.Errorf("password cannot be empty")
	}
	if c.DiskDevice == "" {
		return fmt.Errorf("disk device cannot be empty")
	}
	if c.SSHKey != "" {
		keyData, err := os.ReadFile(c.SSHKey)
		if err != nil {
			return fmt.Errorf("failed to read SSH key file: %w", err)
		}
		if !strings.HasPrefix(string(keyData), "ssh-") {
			return fmt.Errorf("invalid SSH public key format")
		}
	}
	return nil
}
