package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/btassone/alpine-hero/internal/config"
)

// Generator handles the generation of Alpine Linux answer files
type Generator struct {
	config *config.Config
	output string
}

// New creates a new Generator instance
func New(cfg *config.Config, output string) *Generator {
	return &Generator{
		config: cfg,
		output: output,
	}
}

// Generate creates the answer file based on the configuration
func (g *Generator) Generate() error {
	tmplPath := filepath.Join(getTemplateDir(), "answers.tmpl")

	t, err := template.ParseFiles(tmplPath)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	if err := validateOutputPath(g.output); err != nil {
		return err
	}

	f, err := os.Create(g.output)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	if err := f.Chmod(0600); err != nil {
		_ = f.Close()
		return fmt.Errorf("failed to set file permissions: %w", err)
	}
	defer f.Close()

	if err := t.Execute(f, g.config); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	fmt.Printf("Successfully generated answers file: %s\n", g.output)
	return nil
}

func getTemplateDir() string {
	if dir := os.Getenv("TEMPLATE_DIR"); dir != "" {
		return dir
	}
	return "templates"
}

// validateOutputPath ensures the output path is safe and valid
func validateOutputPath(path string) error {
	cleanPath := filepath.Clean(path)
	absPath, err := filepath.Abs(cleanPath)
	if err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}

	if strings.Contains(cleanPath, "..") {
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

	if !filepath.IsAbs(cleanPath) {
		return nil
	}

	allowedPrefixes := []string{
		"/tmp/",
		os.TempDir(),
		filepath.Join(os.Getenv("HOME"), "alpine-template"),
		".",
	}

	if tempDir := os.Getenv("TMPDIR"); tempDir != "" {
		allowedPrefixes = append(allowedPrefixes, tempDir)
	}

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

	if !isAllowed {
		tempPatterns := []string{
			"/var/folders/",
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
