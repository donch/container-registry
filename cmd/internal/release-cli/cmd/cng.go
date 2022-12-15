package cmd

import (
	"log"

	"github.com/docker/distribution/cmd/internal/release-cli/client"
	"github.com/spf13/cobra"
)

var cngCmd = &cobra.Command{
	Use:   "cng",
	Short: "Release to Cloud Native container images components of GitLab",
	Run: func(cmd *cobra.Command, args []string) {
		client.Init(cmd.Use, nil)
		err := client.SendRequestToDeps(cngTriggerToken)
		if err != nil {
			log.Fatalf("Failed to trigger a pipeline in CNG: %v", err)
		}
	},
}

var cngTriggerToken string

func init() {
	releaseCmd.AddCommand(cngCmd)
	cngCmd.Flags().StringVar(&cngTriggerToken, "trigger-token", "", "Trigger token for pipeline trigering")
	cngCmd.MarkFlagRequired("trigger-token")
}
