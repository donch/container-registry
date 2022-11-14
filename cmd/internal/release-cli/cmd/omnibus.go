package cmd

import (
	"github.com/docker/distribution/cmd/internal/release-cli/client"
	"github.com/spf13/cobra"
)

var omnibusCmd = &cobra.Command{
	Use:   "omnibus",
	Short: "Release to Omnibus GitLab",
	Run: func(cmd *cobra.Command, args []string) {
		client.Init(cmd.Use, nil)
		client.SendRequestToDeps()
	},
}

func init() {
	releaseCmd.AddCommand(omnibusCmd)
}
