package cmd

import (
	"github.com/btassone/alpine-hero/internal/generator"
	"github.com/spf13/cobra"
)

func newGenerateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate the answers file",
		Long:  `Generate an answers file based on the provided configuration or default values.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			gen := generator.New(cfg, outputFile)
			return gen.Generate()
		},
	}

	// Add flags specific to generate command
	cmd.Flags().StringVarP(&cfg.Hostname, "hostname", "n", cfg.Hostname, "Hostname for the Alpine system")
	cmd.Flags().StringVarP(&cfg.Username, "username", "u", cfg.Username, "Username for the main user")
	cmd.Flags().StringVarP(&cfg.Password, "password", "p", cfg.Password, "Password for the main user")
	cmd.Flags().StringVarP(&cfg.Timezone, "timezone", "t", cfg.Timezone, "Timezone for the system")
	cmd.Flags().StringVarP(&cfg.Keymap, "keymap", "k", cfg.Keymap, "Keyboard layout")
	cmd.Flags().StringVarP(&cfg.NetworkIface, "interface", "i", cfg.NetworkIface, "Network interface to configure")
	cmd.Flags().StringVarP(&cfg.DiskDevice, "disk", "d", cfg.DiskDevice, "Disk device for installation")
	cmd.Flags().StringSliceVar(&cfg.Groups, "groups", cfg.Groups, "User groups (comma-separated)")
	cmd.Flags().StringVar(&cfg.SSHKey, "ssh-key", "", "Path to SSH public key file")
	cmd.Flags().StringVarP(&outputFile, "output", "o", "answers.txt", "Output file path")

	return cmd
}
