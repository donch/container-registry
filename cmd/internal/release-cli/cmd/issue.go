package cmd

import (
	"fmt"
	"log"

	"github.com/docker/distribution/cmd/internal/release-cli/client"
	"github.com/spf13/cobra"
	"github.com/xanzy/go-gitlab"
)

var issueCmd = &cobra.Command{
	Use:   "issue",
	Short: "Create a Release Plan issue",
	Run: func(cmd *cobra.Command, args []string) {
		client.Init(cmd.Use, nil)

		labels := &gitlab.Labels{
			"devops::package",
			"group::container registry",
			"section::ops",
			"type::maintenance",
			"maintenance::dependency",
			"Category:Container Registry",
			"backend",
			"golang",
			"workflow::in dev",
		}

		changelog, err := client.GetChangelog()
		if err != nil {
			log.Fatalf("Failed to get changelog: %v", err)
		}

		updatedDescription, err := client.UpdateIssueDescription(changelog)
		if err != nil {
			log.Fatalf("Failed to copy the changelog: %v", err)
		}

		issue, err := client.CreateReleasePlan(labels, string(updatedDescription))
		if err != nil {
			log.Fatalf("Failed to create the Release Plan issue: %v", err)
		}

		fmt.Printf("Created Release Plan at: %s\n", issue.WebURL)
	},
}

func init() {
	releaseCmd.AddCommand(issueCmd)
}
