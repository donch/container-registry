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

var (
	targetProjectID = flag.String("target-project-id", "", "The project ID of the target project where MR will be created against")
	targetBranch    = flag.String("target-branch", "master", "The branch name to target the MR against")
	targetFilename  = flag.String("target-filename", "bases/environments.yaml", "The name of the file to modify in the target MR")
	sourceProjectID = flag.String("source-project-id", "", "Project ID of the source project to get details from")
	sourceVersion   = flag.String("source-version", "", "The version that will be used in the update MR (e.g. $CI_COMMIT_TAG)")
	gitlabAuthToken = flag.String("gitlab-auth-token", "", "PAT or $CI_JOB_TOKEN to use to authenticate against the GitLab API")
)

type releaser struct {
	targetProjectID string
	targetBranch    string
	targetFilename  string
	sourceProjectID string
	sourceVersion   string
	gitlabClient    *gitlab.Client
}

func main() {
	flag.Parse()

	gitlab, err := gitlab.NewClient(*gitlabAuthToken)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	release := &releaser{
		targetProjectID: *targetProjectID,
		targetBranch:    *targetBranch,
		targetFilename:  *targetFilename,
		sourceProjectID: *sourceProjectID,
		sourceVersion:   *sourceVersion,
		gitlabClient:    gitlab,
	}

	b1, err := release.CreateReleaseBranch("bump-registry-version-pre-gstg")
	if err != nil {
		log.Fatalf("Failed to create pre/gstg branch: %v", err)
	}

	b2, err := release.CreateReleaseBranch("bump-registry-version-prod")
	if err != nil {
		log.Fatalf("Failed to create grpd branch: %v", err)
	}

	file, err := release.GetDecodedTargetFile()
	if err != nil {
		log.Fatalf("Failed to decode target file: %v", err)
	}

	if err := CreateAndCopyRepositoryFile("tmp1", file); err != nil {
		log.Fatalf("Failed to create tmp1 file: %v", err)
	}

	if err := CreateAndCopyRepositoryFile("tmp2", file); err != nil {
		log.Fatalf("Failed to create tmp2 file: %v", err)
	}

	changePreStg, err := release.UpdateRegistryVersion("tmp1", "pre/gstg")
	if err != nil {
		log.Fatalf("Failed to update registry version for pre/gstg: %v", err)
	}

	changeProd, err := release.UpdateRegistryVersion("tmp2", "gprd")
	if err != nil {
		log.Fatalf("Failed to update registry version for gprd: %v", err)
	}

	if err := release.CreateReleaseCommit(changePreStg, "pre/gstg", b1); err != nil {
		log.Fatalf("Failed to create release commit for pre/gstg: %v", err)
	}

	if err := release.CreateReleaseCommit(changeProd, "gprd" ,b2); err != nil {
		log.Fatalf("Failed to create release commit for gprd: %v", err)
	}

	d1, err := release.GetChangelog()
	if err != nil {
		log.Fatalf("Failed to get changelog: %v", err)
	}

	m1, err := release.CreateReleaseMergeRequest("pre/gstg", d1, b1)
	if err != nil {
		log.Fatalf("Failed to create MR for pre/gstg: %v", err)
	}

	fmt.Printf("Created MR for pre/gstg: %s\n", m1.WebURL)

	d2 := fmt.Sprintf("%s but for gprd", m1.WebURL)
	m2, err := release.CreateReleaseMergeRequest("gprd", d2, b2)
	if err != nil {
		log.Fatalf("Failed to create MR for gprd: %v", err)
	}

	fmt.Printf("Created MR for gprd: %s\n", m2.WebURL)
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
	if env == "gprd" {
		for i, line := range lines {
			if strings.Contains(line, "registry_version") {
				if occurs == 3 {
					lines[i] = fmt.Sprintf("        registry_version: %s", r.sourceVersion)
					break
				}
				occurs++
			}
		}
	} else if env == "pre/gstg" {
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

func (r *releaser) GetDecodedTargetFile() ([]byte, error) {
	rfo := &gitlab.GetFileOptions{
		Ref: gitlab.String(r.targetBranch),
	}
	file, _, err := r.gitlabClient.RepositoryFiles.GetFile(r.targetProjectID, r.targetFilename, rfo)
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
		Action:   gitlab.FileAction(gitlab.FileUpdate),
		FilePath: gitlab.String(r.targetFilename),
		Content:  gitlab.String(string(commit)),
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
	mro := &gitlab.CreateMergeRequestOptions{
		Title:        gitlab.String(fmt.Sprintf("Bump container registry to %s (%s)", r.sourceVersion, env)),
		Description:  gitlab.String(description),
		SourceBranch: gitlab.String(b.Name),
		TargetBranch: gitlab.String(r.targetBranch),
		Squash:       gitlab.Bool(true),
		Labels:       &gitlab.Labels{"workflow::ready for review", "team::Delivery", "Service::Container Registry"},
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
