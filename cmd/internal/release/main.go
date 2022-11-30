package main

import (
	"encoding/base64"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/xanzy/go-gitlab"
)

const (
	EnvGDK     = "gdk"
	EnvPreGstg = "pre/gstg"
	EnvGprd    = "gprd"
)

var (
	targetProjectID       = flag.String("target-project-id", "", "The project ID of the target project where MR will be created against")
	targetBranch          = flag.String("target-branch", "master", "The branch name to target the MR against")
	targetFilenames       = flag.String("target-filenames", "", "List of the files to modify in the target MR, separated by a comma")
	sourceProjectID       = flag.String("source-project-id", "", "Project ID of the source project to get details from")
	sourceVersion         = flag.String("source-version", "", "The version that will be used in the update MR (e.g. $CI_COMMIT_TAG)")
	gitlabAuthToken       = flag.String("gitlab-auth-token", "", "PAT or $CI_JOB_TOKEN to use to authenticate against the GitLab API")
	k8sWorkloads          = flag.Bool("k8s-workloads", false, "specify if the release is for K8s Workloads")
	gdk                   = flag.Bool("gdk", false, "specify if the release is for the GDK")
	issueTemplateFilename = flag.String("template-filename", ".gitlab/issue_templates/Release Plan.md", "The path to the issue template to apply")
	newIssue              = flag.Bool("new-issue", false, "specify whether to create an issue using the template-filename")
)

type releaser struct {
	targetProjectID       string
	targetBranch          string
	targetFilenames       []string
	sourceProjectID       string
	sourceVersion         string
	issueTemplateFilename string
	gitlabClient          *gitlab.Client
}

func main() {
	flag.Parse()
	files := strings.Split(*targetFilenames, ",")

	gitlab, err := gitlab.NewClient(*gitlabAuthToken)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	release := &releaser{
		targetProjectID: *targetProjectID,
		targetBranch:    *targetBranch,
		targetFilenames: files,
		sourceProjectID: *sourceProjectID,
		sourceVersion:   *sourceVersion,
		gitlabClient:    gitlab,
	}

	if *newIssue {
		release.createReleaseIssue()
	} else {
		release.createReleaseMRs()
	}
}

func (r *releaser) createReleaseMRs() {
	if *k8sWorkloads {
		b1, err := r.CreateReleaseBranch("bump-registry-version-pre-gstg")
		if err != nil {
			log.SetPrefix("[k8s-version-bump-pre/gstg]: ")
			log.Fatalf("Failed to create branch: %v", err)
		}

		b2, err := r.CreateReleaseBranch("bump-registry-version-prod")
		if err != nil {
			log.SetPrefix("[k8s-version-bump-gprd]: ")
			log.Fatalf("Failed to create branch: %v", err)
		}

		file, err := r.GetDecodedTargetFile(r.targetFilenames[0])
		if err != nil {
			log.Fatalf("Failed to decode target file: %v", err)
		}

		if err := CreateAndCopyRepositoryFile("tmp1", file); err != nil {
			log.Fatalf("Failed to create tmp1 file: %v", err)
		}

		if err := CreateAndCopyRepositoryFile("tmp2", file); err != nil {
			log.Fatalf("Failed to create tmp2 file: %v", err)
		}

		changePreStg, err := r.UpdateRegistryVersion("tmp1", EnvPreGstg)
		if err != nil {
			log.SetPrefix("[k8s-version-bump-pre/gstg]: ")
			log.Fatalf("Failed to update registry version: %v", err)
		}

		changeProd, err := r.UpdateRegistryVersion("tmp2", EnvGprd)
		if err != nil {
			log.SetPrefix("[k8s-version-bump-gprd]: ")
			log.Fatalf("Failed to update registry version for gprd: %v", err)
		}

		if err := r.CreateReleaseCommit(changePreStg, EnvPreGstg, b1); err != nil {
			log.SetPrefix("[k8s-version-bump-pre/gstg]: ")
			log.Fatalf("Failed to create release commit: %v", err)
		}

		if err := r.CreateReleaseCommit(changeProd, EnvGprd, b2); err != nil {
			log.SetPrefix("[k8s-version-bump-gprd]: ")
			log.Fatalf("Failed to create release commit: %v", err)
		}

		d1, err := r.GetChangelog()
		if err != nil {
			log.Fatalf("Failed to get changelog: %v", err)
		}

		m1, err := r.CreateReleaseMergeRequest(EnvPreGstg, d1, b1)
		if err != nil {
			log.SetPrefix("[k8s-version-bump-pre/gstg]: ")
			log.Fatalf("Failed to create MR: %v", err)
		}

		fmt.Printf("Created MR for pre/gstg: %s\n", m1.WebURL)

		d2 := fmt.Sprintf("%s but for gprd", m1.WebURL)

		m2, err := r.CreateReleaseMergeRequest(EnvGprd, d2, b2)
		if err != nil {
			log.SetPrefix("[k8s-version-bump-gprd]: ")
			log.Fatalf("Failed to create MR: %v", err)
		}

		fmt.Printf("Created MR for gprd: %s\n", m2.WebURL)
	} else if *gdk {
		log.SetPrefix("[k8s-version-bump-gdk]: ")

		b1, err := r.CreateReleaseBranch("bump-registry-version")
		if err != nil {
			log.Fatalf("Failed to create branch: %v", err)
		}

		for i := range r.targetFilenames {
			file, err := r.GetDecodedTargetFile(r.targetFilenames[i])
			if err != nil {
				log.Fatalf("Failed to decode target file: %v", err)
			}

			if err := CreateAndCopyRepositoryFile("tmp", file); err != nil {
				log.Fatalf("Failed to create tmp file: %v", err)
			}

			changeGdk, err := r.UpdateRegistryVersion("tmp", EnvGDK)
			if err != nil {
				log.Fatalf("Failed to update registry version: %v", err)
			}

			if err := r.CreateReleaseCommit(changeGdk, EnvGDK, b1); err != nil {
				log.Fatalf("Failed to create release commit: %v", err)
			}
		}
		d1, err := r.GetChangelog()
		if err != nil {
			log.Fatalf("Failed to get changelog: %v", err)
		}
		m1, err := r.CreateReleaseMergeRequest(EnvGDK, d1, b1)
		if err != nil {
			log.Fatalf("Failed to create MR: %v", err)
		}

		fmt.Printf("Created MR for GDK: %s\n", m1.WebURL)
	}
}

func (r *releaser) createReleaseIssue() {
	log.SetPrefix("[create-release-issue]: ")

	changelog, err := r.GetChangelog()
	if err != nil {
		log.Fatalf("Failed to get changelog: %v", err)
	}

	file, err := r.GetDecodedTargetFile(r.issueTemplateFilename)
	if err != nil {
		log.Fatalf("Failed to decode Issue Template file: %v", err)
	}

	if err := CreateAndCopyRepositoryFile("tmp1", file); err != nil {
		log.Fatalf("Failed to create tmp1 file: %v", err)
	}

	updatedDescription, err := r.UpdateIssueDescription("tmp1", changelog)
	if err != nil {
		log.Fatalf("Failed to copy the changelog to the issue template: %v", err)
	}

	issue, err := r.CreateReleasePlan(string(updatedDescription))
	if err != nil {
		log.Fatalf("Failed to create the Release Plan issue: %v", err)
	}

	fmt.Printf("Created Release Plan at: %s\n", issue.WebURL)
}

func (r *releaser) CreateReleasePlan(changelog string) (*gitlab.Issue, error) {
	ico := &gitlab.CreateIssueOptions{
		Title:       gitlab.String(fmt.Sprintf("Release %s", strings.TrimSuffix(r.sourceVersion, "-gitlab"))),
		Description: gitlab.String(fmt.Sprintf("%s", changelog)),
		Labels: &gitlab.Labels{
			"devops::package",
			"group::container registry",
			"section::ops",
			"type::maintenance",
			"maintenance::dependency",
			"Category:Container Registry",
			"backend",
			"golang",
			"workflow::in dev",
		},
	}

	issue, _, err := r.gitlabClient.Issues.CreateIssue(r.targetProjectID, ico)
	if err != nil {
		return nil, err
	}

	return issue, nil
}

func (r *releaser) UpdateIssueDescription(fileName string, changelog string) ([]byte, error) {
	f, err := os.ReadFile(fileName)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(f), "\n")
	for i, line := range lines {
		if strings.Contains(line, "[copy changelog here]") {
			lines[i] = fmt.Sprintf(changelog)
			break
		}
	}

	out := strings.Join(lines, "\n")
	err = os.WriteFile(fileName, []byte(out), 0644)
	if err != nil {
		return nil, err
	}

	cng, err := os.ReadFile(fileName)
	if err != nil {
		return nil, err
	}

	return cng, nil
}

func (r *releaser) UpdateRegistryVersion(fileName string, env string) ([]byte, error) {
	f, err := os.ReadFile(fileName)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(f), "\n")
	occurs := 1

	// note: replacing registry_version is dependent on the
	// the bases/environments.yaml found on the k8s workloads project
	// https://gitlab.com/gitlab-com/gl-infra/k8s-workloads/gitlab-com/-/blob/master/bases/environments.yaml
	// if that file changes this script may break
	if env == EnvGprd {
		for i, line := range lines {
			if strings.Contains(line, "registry_version") {
				if occurs == 3 {
					lines[i] = fmt.Sprintf("        registry_version: %s", r.sourceVersion)
					break
				}
				occurs++
			}
		}
	} else if env == EnvPreGstg {
			for i, line := range lines {
				if strings.Contains(line, "registry_version") {
					if occurs <= 2 {
						lines[i] = fmt.Sprintf("        registry_version: %s", r.sourceVersion)
						occurs++
					} else {
						break
					}
			}
		}
	} else if env == EnvGDK && fileName == "support/docker-registry" {
			for i, line := range lines {
				if strings.Contains(line, "registry_image:-registry.gitlab.com/gitlab-org/build/cng") {
					lines[i] = fmt.Sprintf("      \"${registry_image:-registry.gitlab.com/gitlab-org/build/cng/gitlab-container-registry:%s}\"", r.sourceVersion)
					break
				}
			}
		} else if env == EnvGDK && fileName == "lib/gdk/config.rb" {
			for i, line := range lines {
				if strings.Contains(line, "registry.gitlab.com/gitlab-org/build/cng/gitlab-container-registry:") {
					lines[i + 1] = fmt.Sprintf("        '%s'", r.sourceVersion)
					break
				}
			}
	}

	out := strings.Join(lines, "\n")
	err = os.WriteFile(fileName, []byte(out), 0644)
	if err != nil {
		return nil, err
	}

	cng, err := os.ReadFile(fileName)
	if err != nil {
		return nil, err
	}

	return cng, nil
}

func (r *releaser) GetDecodedTargetFile(fileName string) ([]byte, error) {
	rfo := &gitlab.GetFileOptions{
		Ref: gitlab.String(r.targetBranch),
	}
	file, _, err := r.gitlabClient.RepositoryFiles.GetFile(r.targetProjectID, fileName, rfo)
	if err != nil {
		return nil, err
	}

	dec, err := base64.StdEncoding.DecodeString(file.Content)
	if err != nil {
		return nil, err
	}

	return dec, nil

}

func CreateAndCopyRepositoryFile(fileName string, dec []byte) error {
	f, err := os.Create(fileName)
	if err != nil {
		return err
	}

	defer f.Close()

	if _, err := f.Write(dec); err != nil {
		return err
	}

	if err := f.Sync(); err != nil {
		return err
	}

	return nil
}

func (r *releaser) CreateReleaseCommit(commit []byte, env string, branch *gitlab.Branch) error {
	aco := &gitlab.CommitActionOptions{
		Action:  gitlab.FileAction(gitlab.FileUpdate),
		Content: gitlab.String(string(commit)),
	}
	co := &gitlab.CreateCommitOptions{
		Branch:        gitlab.String(branch.Name),
		CommitMessage: gitlab.String(fmt.Sprintf("%s: Bump Container Registry to %s", env, r.sourceVersion)),
		Actions:       []*gitlab.CommitActionOptions{aco},
	}
	_, _, err := r.gitlabClient.Commits.CreateCommit(r.targetProjectID, co)
	return err
}

func (r *releaser) CreateReleaseMergeRequest(env string, description string, b *gitlab.Branch) (*gitlab.MergeRequest, error) {
	labels := &gitlab.Labels{"workflow::ready for review", "team::Delivery", "Service::Container Registry"}

	if env == EnvGDK {
		labels = &gitlab.Labels{"workflow::ready for review", "group::container registry", "devops::package"}
	}

	mro := &gitlab.CreateMergeRequestOptions{
		Title:        gitlab.String(fmt.Sprintf("Bump container registry to %s (%s)", r.sourceVersion, env)),
		Description:  gitlab.String(description),
		SourceBranch: gitlab.String(b.Name),
		TargetBranch: gitlab.String(r.targetBranch),
		Squash:       gitlab.Bool(true),
		Labels:       labels,
	}
	mr, _, err := r.gitlabClient.MergeRequests.CreateMergeRequest(r.targetProjectID, mro)
	if err != nil {
		return nil, err
	}

	return mr, nil
}

func (r *releaser) CreateReleaseBranch(name string) (*gitlab.Branch, error) {
	bo := &gitlab.CreateBranchOptions{
		Branch: gitlab.String(fmt.Sprintf("%s-%s", name, r.sourceVersion)),
		Ref:    gitlab.String(r.targetBranch),
	}
	b, _, err := r.gitlabClient.Branches.CreateBranch(r.targetProjectID, bo)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func (r *releaser) GetChangelog() (string, error) {
	tag, _, err := r.gitlabClient.Tags.GetTag(r.sourceProjectID, r.sourceVersion)
	if err != nil {
		return "", err
	}

	return tag.Commit.Message, nil
}