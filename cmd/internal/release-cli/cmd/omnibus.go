package cmd

import (
	"log"

	"github.com/docker/distribution/cmd/internal/release-cli/client"
	"github.com/spf13/cobra"
)

var omnibusCmd = &cobra.Command{
	Use:   "omnibus",
	Short: "Release to Omnibus GitLab",
	Run: func(cmd *cobra.Command, args []string) {
		client.Init(cmd.Use, nil)
		err := client.SendRequestToDeps(omnibusTriggerToken)
		if err != nil {
			log.Fatalf("Failed to trigger a pipeline in Omnibus: %v", err)
		}
	},
}

var omnibusTriggerToken string

func init() {
	releaseCmd.AddCommand(omnibusCmd)
	omnibusCmd.Flags().StringVar(&omnibusTriggerToken, "trigger-token", "", "Trigger token for pipeline trigering")
	omnibusCmd.MarkFlagRequired("trigger-token")
}
