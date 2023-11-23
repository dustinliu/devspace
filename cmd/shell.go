package cmd

import "github.com/spf13/cobra"

var shellCmd = &cobra.Command{
	Use:   "devspace shell env",
	Short: "spawn a shell in dev environment",
	Run: func(cmd *cobra.Command, args []string) {
		// Do Stuff Here
	},
}
