package cmd

import (
	"github.com/docker/distribution/cmd/internal/release-cli/client"
	"github.com/spf13/cobra"
)

var cngCmd = &cobra.Command{
	Use:   "cng",
	Short: "Release to Cloud Native container images components of GitLab",
	Run: func(cmd *cobra.Command, args []string) {
		client.Init(cmd.Use, nil)
		client.SendRequestToDeps()
	},
}

func init() {
	releaseCmd.AddCommand(cngCmd)
}
