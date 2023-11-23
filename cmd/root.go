package cmd

import (
	"github.com/dustinliu/devspace/core"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.PersistentFlags().BoolVarP(&core.DebugMode, "debug", "d", false, "debug mode")
}

var (
	debug bool

	rootCmd = &cobra.Command{
		Use:   "devspace [command]",
		Short: "manage development environment",
		PersistentPreRun: func(_ *cobra.Command, _ []string) {
		},
	}
)

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		core.Fatal(err)
	}
}
