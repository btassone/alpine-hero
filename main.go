package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/spf13/cobra"
)

var osExit = os.Exit

type AlpineConfig struct {
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

var (
	config     AlpineConfig
	outputFile string
)

var rootCmd = &cobra.Command{
	Use:   "alpine-template",
	Short: "Alpine Linux answer file generator",
	Long: `A CLI tool to generate Alpine Linux answer files for automated installation.
This tool helps create the answers file needed for automated Alpine Linux installation.`,
}

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate the answers file",
	Long:  `Generate an answers file based on the provided configuration or default values.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return generateAnswersFile()
	},
}

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate the current configuration",
	Long:  `Check if the current configuration values are valid for Alpine Linux installation.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return validateConfig()
	},
}

func init() {
	// Initialize default configuration
	config = AlpineConfig{
		Hostname:     "alpinehost",
		Username:     "alpine",
		Password:     "changeme",
		Timezone:     "UTC",
		Keymap:       "us",
		NetworkIface: "eth0",
		DiskDevice:   "/dev/mmcblk0",
		Groups:       []string{"audio", "video", "netdev"},
	}

	// Add commands to root command
	rootCmd.AddCommand(generateCmd)
	rootCmd.AddCommand(validateCmd)

	// Add flags to generate command
	generateCmd.Flags().StringVarP(&config.Hostname, "hostname", "n", config.Hostname, "Hostname for the Alpine system")
	generateCmd.Flags().StringVarP(&config.Username, "username", "u", config.Username, "Username for the main user")
	generateCmd.Flags().StringVarP(&config.Password, "password", "p", config.Password, "Password for the main user")
	generateCmd.Flags().StringVarP(&config.Timezone, "timezone", "t", config.Timezone, "Timezone for the system")
	generateCmd.Flags().StringVarP(&config.Keymap, "keymap", "k", config.Keymap, "Keyboard layout")
	generateCmd.Flags().StringVarP(&config.NetworkIface, "interface", "i", config.NetworkIface, "Network interface to configure")
	generateCmd.Flags().StringVarP(&config.DiskDevice, "disk", "d", config.DiskDevice, "Disk device for installation")
	generateCmd.Flags().StringSliceVar(&config.Groups, "groups", config.Groups, "User groups (comma-separated)")
	generateCmd.Flags().StringVar(&config.SSHKey, "ssh-key", "", "Path to SSH public key file")
	generateCmd.Flags().StringVarP(&outputFile, "output", "o", "answers.txt", "Output file path")
}

func getTemplateDir() string {
	if dir := os.Getenv("TEMPLATE_DIR"); dir != "" {
		return dir
	}
	return "templates"
}

func generateAnswersFile() error {
	// Get the template file path using the template directory
	tmplPath := filepath.Join(getTemplateDir(), "answers.tmpl")

	// Parse the template first
	t, err := template.ParseFiles(tmplPath)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	// Validate the output path before attempting to execute template
	if err := validateOutputPath(outputFile); err != nil {
		return err
	}

	// Create or truncate the output file
	f, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	if err := f.Chmod(0600); err != nil {
		_ = f.Close()
		return fmt.Errorf("failed to set file permissions: %w", err)
	}
	defer func(f *os.File) {
		_ = f.Close()
	}(f)

	// Execute the template
	if err := t.Execute(f, config); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	fmt.Printf("Successfully generated answers file: %s\n", outputFile)
	return nil
}

func validateOutputPath(path string) error {
	// Normalize the path first
	cleanPath := filepath.Clean(path)

	// Get absolute path for validation
	absPath, err := filepath.Abs(cleanPath)
	if err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}

	// Handle path traversal first
	if strings.Contains(cleanPath, "..") {
		// Block access to sensitive directories
		normalizedPath := filepath.ToSlash(absPath) + "/"
		if strings.Contains(normalizedPath, "/etc/") ||
			strings.Contains(normalizedPath, "/usr/") ||
			strings.Contains(normalizedPath, "/boot/") ||
			strings.Contains(normalizedPath, "/root/") ||
			(strings.Contains(normalizedPath, "/var/") && !strings.Contains(normalizedPath, "/var/folders/")) {
			return fmt.Errorf("invalid output path: output path not allowed: %s", path)
		}
		return nil
	}

	// If it's a non-traversing relative path, allow it
	if !filepath.IsAbs(cleanPath) {
		return nil
	}

	// Define allowed directories
	allowedPrefixes := []string{
		"/tmp/",
		os.TempDir(),
		filepath.Join(os.Getenv("HOME"), "alpine-template"),
		".", // Current directory
	}

	// Add support for macOS temporary directories
	if tempDir := os.Getenv("TMPDIR"); tempDir != "" {
		allowedPrefixes = append(allowedPrefixes, tempDir)
	}

	// Check if path matches any allowed prefix
	isAllowed := false
	for _, prefix := range allowedPrefixes {
		absPrefix, err := filepath.Abs(prefix)
		if err != nil {
			continue
		}

		if strings.HasPrefix(absPath, absPrefix) {
			isAllowed = true
			break
		}
	}

	// Additional check for temporary directory patterns
	if !isAllowed {
		tempPatterns := []string{
			"/var/folders/", // macOS temp directory pattern
			"/tmp/",
		}
		for _, pattern := range tempPatterns {
			if strings.HasPrefix(absPath, pattern) {
				isAllowed = true
				break
			}
		}
	}

	if !isAllowed {
		return fmt.Errorf("invalid output path: output path not allowed: %s", path)
	}

	// Check parent directory exists and is a directory
	parentDir := filepath.Dir(absPath)
	info, err := os.Stat(parentDir)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("parent directory does not exist: %s", parentDir)
		}
		return fmt.Errorf("cannot access parent directory: %w", err)
	}

	if !info.IsDir() {
		return fmt.Errorf("parent path is not a directory: %s", parentDir)
	}

	return nil
}

func validateConfig() error {
	// Add validation logic here
	if config.Hostname == "" {
		return fmt.Errorf("hostname cannot be empty")
	}
	if config.Username == "" {
		return fmt.Errorf("username cannot be empty")
	}
	if config.Password == "" {
		return fmt.Errorf("password cannot be empty")
	}
	if config.DiskDevice == "" {
		return fmt.Errorf("disk device cannot be empty")
	}
	if config.SSHKey != "" {
		keyData, err := os.ReadFile(config.SSHKey)
		if err != nil {
			return fmt.Errorf("failed to read SSH key file: %w", err)
		}
		if !strings.HasPrefix(string(keyData), "ssh-") {
			return fmt.Errorf("invalid SSH public key format")
		}
	}

	fmt.Println("Configuration validation passed!")
	return nil
}

func main() {
	if pErr := rootCmd.Execute(); pErr != nil {
		_, err := fmt.Fprintf(os.Stderr, "Error: %v\n", pErr)
		if err != nil {
			return
		}
		osExit(1)
	}
}
