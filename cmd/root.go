package cmd

import (
	"github.com/dustinliu/devspace/logging"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.PersistentFlags().BoolVarP(&logging.DebugMode, "debug", "d", false, "debug mode")
}

var rootCmd = &cobra.Command{
	Use:   "devspace [command]",
	Short: "manage development environment",
	PersistentPreRun: func(_ *cobra.Command, _ []string) {
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		logging.Fatal(err)
	}
}
