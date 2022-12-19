package cmd

import (
	"log"

	"github.com/docker/distribution/cmd/internal/release-cli/client"
	"github.com/spf13/cobra"
)

var chartsCmd = &cobra.Command{
	Use:   "charts",
	Short: "Release to Cloud Native GitLab Helm Chart",
	Run: func(cmd *cobra.Command, args []string) {
		client.Init(cmd.Use, nil)
		err := client.SendRequestToDeps(chartsTriggerToken)
		if err != nil {
			log.Fatalf("Failed to trigger a pipeline in Charts: %v", err)
		}
	},
}

var chartsTriggerToken string

func init() {
	releaseCmd.AddCommand(chartsCmd)
	chartsCmd.Flags().StringVar(&chartsTriggerToken, "trigger-token", "", "Trigger token for pipeline trigering")
	chartsCmd.MarkFlagRequired("trigger-token")
}
