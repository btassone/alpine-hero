package cmd

import (
	"github.com/spf13/cobra"
)

var (
	Version    = "dev"
	BuildTime  = "unknown"
	CommitHash = "none"
)

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Long:  `Display version, build time, and commit hash information for alpine-hero.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Use the command's output writer instead of fmt.Printf
			cmd.Printf("alpine-hero version %s\n", Version)
			cmd.Printf("Built at: %s\n", BuildTime)
			cmd.Printf("Commit: %s\n", CommitHash)
			return nil
		},
	}
}
