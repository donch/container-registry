package cmd

import (
	"log"
	"os"

	"github.com/docker/distribution/cmd/internal/release-cli/client"
	"github.com/spf13/cobra"
)

var cngCmd = &cobra.Command{
	Use:   "cng",
	Short: "Manage CNG release",
	Run: func(cmd *cobra.Command, args []string) {
		triggerToken, err := cmd.Flags().GetString("cng-trigger-token")
		if err != nil {
			log.Fatal(err)
		}

		accessToken, err := cmd.Flags().GetString("registry-access-token")
		if err != nil {
			log.Fatal(err)
		}

		version := os.Getenv("CI_COMMIT_TAG")
		if version == "" {
			log.Fatal("Version is empty. Aborting.")
		}

		release, err := readConfig(cmd.Use, version)
		if err != nil {
			log.Fatalf("Error reading config: %v", err)
			return
		}

		gitlabClient := client.NewClient(accessToken)

		err = gitlabClient.SendRequestToDeps(release.ProjectID, triggerToken, release.Ref)
		if err != nil {
			log.Fatalf("Failed to trigger a pipeline in CNG: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(cngCmd)

	cngCmd.Flags().StringP("cng-trigger-token", "", "", "Trigger token for CNG")
	cngCmd.MarkFlagRequired("cng-trigger-token")
}
