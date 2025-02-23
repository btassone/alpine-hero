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
	generateCmd.Flags().StringVarP(&outputFile, "output", "o", "answers.txt", "Output file path")
}

func getTemplateDir() string {
	if dir := os.Getenv("TEMPLATE_DIR"); dir != "" {
		return dir
	}
	return "templates"
}

func generateAnswersFile() error {
	// Validate the output path first
	if err := validateOutputPath(outputFile); err != nil {
		return fmt.Errorf("invalid output path: %w", err)
	}

	// Get the template file path using the template directory
	tmplPath := filepath.Join(getTemplateDir(), "answers.tmpl")

	// Parse the template
	t, err := template.ParseFiles(tmplPath)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	// Create or truncate the output file
	f, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
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
	// Check if path is absolute and normalize it
	cleanPath := filepath.Clean(path)
	if filepath.IsAbs(cleanPath) {
		// For absolute paths, ensure they're within allowed directories
		// You might want to customize this based on your security requirements
		allowedPrefixes := []string{
			"/tmp/",
			os.TempDir(),
			filepath.Join(os.Getenv("HOME"), "alpine-template"),
			".", // Current directory
		}

		allowed := false
		for _, prefix := range allowedPrefixes {
			absPrefix, err := filepath.Abs(prefix)
			if err != nil {
				continue
			}
			if strings.HasPrefix(cleanPath, absPrefix) {
				allowed = true
				break
			}
		}

		if !allowed {
			return fmt.Errorf("output path not allowed: %s", path)
		}
	}

	// Check parent directory exists and is writable
	parentDir := filepath.Dir(cleanPath)
	if info, err := os.Stat(parentDir); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("parent directory does not exist: %s", parentDir)
		}
		return fmt.Errorf("cannot access parent directory: %s", err)
	} else if !info.IsDir() {
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

	fmt.Println("Configuration validation passed!")
	return nil
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		osExit(1)
	}
}
