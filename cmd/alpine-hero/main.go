package main

import (
	"fmt"
	"os"

	"github.com/btassone/alpine-hero/internal/config"
	"github.com/btassone/alpine-hero/internal/generator"
	"github.com/spf13/cobra"
)

var (
	cfg        *config.Config
	outputFile string
)

var rootCmd = &cobra.Command{
	Use:   "alpine-hero",
	Short: "Alpine Linux answer file generator",
	Long: `A CLI tool to generate Alpine Linux answer files for automated installation.
This tool helps create the answers file needed for automated Alpine Linux installation.`,
}

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate the answers file",
	Long:  `Generate an answers file based on the provided configuration or default values.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		gen := generator.New(cfg, outputFile)
		return gen.Generate()
	},
}

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate the current configuration",
	Long:  `Check if the current configuration values are valid for Alpine Linux installation.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cfg.Validate()
	},
}

func init() {
	cfg = config.New()

	rootCmd.AddCommand(generateCmd)
	rootCmd.AddCommand(validateCmd)

	generateCmd.Flags().StringVarP(&cfg.Hostname, "hostname", "n", cfg.Hostname, "Hostname for the Alpine system")
	generateCmd.Flags().StringVarP(&cfg.Username, "username", "u", cfg.Username, "Username for the main user")
	generateCmd.Flags().StringVarP(&cfg.Password, "password", "p", cfg.Password, "Password for the main user")
	generateCmd.Flags().StringVarP(&cfg.Timezone, "timezone", "t", cfg.Timezone, "Timezone for the system")
	generateCmd.Flags().StringVarP(&cfg.Keymap, "keymap", "k", cfg.Keymap, "Keyboard layout")
	generateCmd.Flags().StringVarP(&cfg.NetworkIface, "interface", "i", cfg.NetworkIface, "Network interface to configure")
	generateCmd.Flags().StringVarP(&cfg.DiskDevice, "disk", "d", cfg.DiskDevice, "Disk device for installation")
	generateCmd.Flags().StringSliceVar(&cfg.Groups, "groups", cfg.Groups, "User groups (comma-separated)")
	generateCmd.Flags().StringVar(&cfg.SSHKey, "ssh-key", "", "Path to SSH public key file")
	generateCmd.Flags().StringVarP(&outputFile, "output", "o", "answers.txt", "Output file path")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
