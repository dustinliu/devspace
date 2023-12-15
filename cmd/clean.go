package cmd

import (
	"github.com/dustinliu/devspace/core"
	"github.com/dustinliu/devspace/logging"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(cleanCmd)
}

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "clean unused docker images and containers",
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
