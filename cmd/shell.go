package cmd

import (
	"github.com/dustinliu/devspace/core"
	"github.com/dustinliu/devspace/logging"
	"github.com/spf13/cobra"
)

var stop_container bool

func init() {
	rootCmd.AddCommand(shellCmd)
	shellCmd.LocalFlags().BoolVarP(&stop_container, "stop", "s", false, "stop container after shell exit")
}

var shellCmd = &cobra.Command{
	Use:   "shell",
	Short: "spawn a shell in dev environment",
	Run: func(_ *cobra.Command, _ []string) {
		project, err := core.NewProject()
		if err != nil {
			logging.Fatal(err)
		}
		if err = project.Shell(stop_container); err != nil {
			logging.Fatal(err)
		}
	},
}
