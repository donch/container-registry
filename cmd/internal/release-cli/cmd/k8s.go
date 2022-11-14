package cmd

import (
	"fmt"
	"log"

	"github.com/docker/distribution/cmd/internal/release-cli/client"
	"github.com/spf13/cobra"
	"github.com/xanzy/go-gitlab"
)

var k8sCmd = &cobra.Command{
	Use:   "k8s",
	Short: "Release to Kubernetes Workload configurations for GitLab.com",

	Run: func(cmd *cobra.Command, args []string) {
		opts := map[string]string{"stage": stage}

		client.Init(cmd.Use, opts)

		labels := &gitlab.Labels{
			"workflow::ready for review",
			"team::Delivery",
			"Service::Container Registry",
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

var stage string

func init() {
	releaseCmd.AddCommand(k8sCmd)
	k8sCmd.Flags().StringVar(&stage, "stage", "", "Stage in the environment")
	k8sCmd.MarkFlagRequired("stage")
}
