package cmd

import (
	"fmt"
	"log"

	"github.com/docker/distribution/cmd/internal/release-cli/client"
	"github.com/spf13/cobra"
	"github.com/xanzy/go-gitlab"
)

var gdkCmd = &cobra.Command{
	Use:   "gdk",
	Short: "Release to GitLab Development Kit",
	Run: func(cmd *cobra.Command, args []string) {
		client.Init(cmd.Use, nil)

		labels := &gitlab.Labels{
			"workflow::ready for review",
			"group::container registry",
			"devops::package",
		}

		branch, err := client.CreateReleaseBranch()
		if err != nil {
			log.Fatalf("Failed to create branch: %v", err)
		}

		if err := client.UpdateAllPaths(branch); err != nil {
			log.Fatalf("Failed to update files: %v", err)
		}

		desc, err := client.GetChangelog()
		if err != nil {
			log.Fatalf("Failed to get changelog: %v", err)
		}

		mr, err := client.CreateReleaseMergeRequest(desc, branch, labels)
		if err != nil {
			log.Fatalf("Failed to create MR: %v", err)
		}

		fmt.Printf("Created MR: %s\n", mr.WebURL)
	},
}

func init() {
	releaseCmd.AddCommand(gdkCmd)
}
