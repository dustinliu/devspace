package cmd

import (
	"github.com/dustinliu/devspace/core"
	"github.com/dustinliu/devspace/logging"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(buildCmd)
}

var buildCmd = &cobra.Command{
	Use:   "build [flags]",
	Short: "build development environment",
	Run:   build,
}

func build(_ *cobra.Command, _ []string) {
	project, err := core.NewProject()
	if err != nil {
		logging.Fatal(err)
	}

	if err := project.Build(); err != nil {
		logging.Fatal(err)
	}
}
