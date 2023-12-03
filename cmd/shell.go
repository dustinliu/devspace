package cmd

import (
	"github.com/dustinliu/devspace/core"
	"github.com/dustinliu/devspace/logging"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(shellCmd)
}

var shellCmd = &cobra.Command{
	Use:   "shell env",
	Short: "spawn a shell in dev environment",
	Run: func(_ *cobra.Command, _ []string) {
		project, err := core.NewProject()
		if err != nil {
			logging.Fatal(err)
		}
		if err = project.Shell(); err != nil {
			logging.Fatal(err)
		}
	},
}
