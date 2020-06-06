package cmd

import "github.com/spf13/cobra"

var watchCmd = &cobra.Command{
	Use:    "watch",
	Hidden: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}
