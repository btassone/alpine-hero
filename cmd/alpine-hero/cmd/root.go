package cmd

import (
	"github.com/btassone/alpine-hero/internal/config"
	"github.com/spf13/cobra"
)

var (
	// Shared configuration that will be used across commands
	cfg        *config.Config
	outputFile string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "alpine-hero",
	Short: "Alpine Linux answer file generator",
	Long: `A CLI tool to generate Alpine Linux answer files for automated installation.
This tool helps create the answers file needed for automated Alpine Linux installation.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cfg = config.New()

	// Add all subcommands
	rootCmd.AddCommand(newGenerateCmd())
	rootCmd.AddCommand(newValidateCmd())
	rootCmd.AddCommand(newVersionCmd())
}
