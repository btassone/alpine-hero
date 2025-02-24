package cmd

import (
	"github.com/spf13/cobra"
)

func newValidateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "validate",
		Short: "Validate the current configuration",
		Long:  `Check if the current configuration values are valid for Alpine Linux installation.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cfg.Validate()
		},
	}
}
