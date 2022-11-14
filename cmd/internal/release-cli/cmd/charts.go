package cmd

import (
	"github.com/docker/distribution/cmd/internal/release-cli/client"
	"github.com/spf13/cobra"
)

var chartsCmd = &cobra.Command{
	Use:   "charts",
	Short: "Release to Cloud Native GitLab Helm Chart",
	Run: func(cmd *cobra.Command, args []string) {
		client.Init(cmd.Use, nil)
		client.SendRequestToDeps()
	},
}

func init() {
	releaseCmd.AddCommand(chartsCmd)
}
