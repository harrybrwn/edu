package cmd

import "github.com/spf13/cobra"

var serviceCmd = &cobra.Command{
	Use:    "service",
	Hidden: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

func runService() error {
	return nil
}
