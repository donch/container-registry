package client

import (
	"encoding/base64"
	"fmt"
	"log"

	"os"
	"strings"

	"github.com/docker/distribution/cmd/internal/release-cli/configuration"
	"github.com/xanzy/go-gitlab"
)

type Release struct {
	name          string
	token         string
	projectID     string
	version       string
	stage         string
	ref           string
	paths         []configuration.Path
	mrTitle       string
	branchName    string
	commitMessage string
}

type Issue struct {
	title     string
	template  string
	token     string
	projectID string
	version   string
	ref       string
}

var release *Release
var issue *Issue
var gtlb *gitlab.Client

func Init(cmdName string, opts map[string]string) {
	var err error

	version := configuration.GetVersion()
	authToken := configuration.GetAuthToken()

	switch cmdName {
	case "k8s":
		cfg := configuration.GetK8sEnvConfig()
		for _, s := range cfg.Stages {
			if s.StageName == opts["stage"] {
				release = &Release{
					name:          cmdName,
					projectID:     s.ProjectID,
					token:         authToken,
					version:       version,
					ref:           s.Ref,
					paths:         s.Paths,
					mrTitle:       s.MRTitle,
					branchName:    s.BranchName,
					commitMessage: s.CommitMessage,
				}
			}
		}
	case "gdk":
		cfg := configuration.GetGDKEnvConfig()
		release = &Release{
			name:          cmdName,
			projectID:     cfg.ProjectID,
			token:         authToken,
			version:       version,
			ref:           cfg.Ref,
			paths:         cfg.Paths,
			mrTitle:       cfg.MRTitle,
			branchName:    cfg.BranchName,
			commitMessage: cfg.CommitMessage,
		}
	case "cng":
		cfg := configuration.GetCNGEnvConfig()
		release = &Release{
			name:      cmdName,
			projectID: cfg.ProjectID,
			token:     authToken,
			ref:       cfg.Ref,
		}
	case "omnibus":
		cfg := configuration.GetOmnibusEnvConfig()
		release = &Release{
			name:      cmdName,
			projectID: cfg.ProjectID,
			token:     authToken,
			ref:       cfg.Ref,
		}
	case "charts":
		cfg := configuration.GetChartsEnvConfig()
		release = &Release{
			name:      cmdName,
			projectID: cfg.ProjectID,
			token:     authToken,
			ref:       cfg.Ref,
		}
	case "issue":
		cfg := configuration.GetIssueConfig()
		issue = &Issue{
			title:     cfg.Title,
			template:  cfg.Template,
			projectID: cfg.ProjectID,
			token:     authToken,
			version:   version,
			ref:       cfg.Ref,
		}
	default:
	}

	gtlb, err = gitlab.NewClient(authToken)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
}

func CreateReleasePlan(labels *gitlab.Labels, changelog string) (*gitlab.Issue, error) {
	ico := &gitlab.CreateIssueOptions{
		Title:       gitlab.String(fmt.Sprintf("%s for %s", issue.title, strings.TrimSuffix(issue.version, "-gitlab"))),
		Description: gitlab.String(changelog),
		Labels:      labels,
	}

	issue, _, err := gtlb.Issues.CreateIssue(issue.projectID, ico)
	if err != nil {
		return nil, err
	}

	return issue, nil
}

func getDecodedFile(fileName string, pid string, ref string) ([]byte, error) {
	rfo := &gitlab.GetFileOptions{
		Ref: gitlab.String(ref),
	}
	file, _, err := gtlb.RepositoryFiles.GetFile(pid, fileName, rfo)
	if err != nil {
		return nil, err
	}

	dec, err := base64.StdEncoding.DecodeString(file.Content)
	if err != nil {
		return nil, err
	}

	return dec, nil
}

func UpdateIssueDescription(changelog string) ([]byte, error) {
	file, err := getDecodedFile(issue.template, issue.projectID, issue.ref)
	if err != nil {
		log.Fatalf("Failed to decode file: %v", err)
	}

	fileName, err := createAndCopyRepositoryFile(file)
	if err != nil {
		log.Fatalf("Failed to create file: %v", err)
	}

	lines, err := readFromTempFile(fileName)
	if err != nil {
		return nil, err
	}

	for i, line := range lines {
		if strings.Contains(line, "[copy changelog here]") {
			lines[i] = changelog
			break
		}
	}

	out := strings.Join(lines, "\n")
	updatedFile, err := loadFileChange(fileName, out)
	if err != nil {
		return nil, err
	}

	return updatedFile, nil
}

func UpdateAllPaths(branch *gitlab.Branch) error {
	var changes []byte

	for i := range release.paths {
		file, err := getDecodedFile(release.paths[i].Filename, release.projectID, release.ref)
		if err != nil {
			log.Fatalf("Failed to decode file: %v", err)
		}

		fileName, err := createAndCopyRepositoryFile(file)
		if err != nil {
			log.Fatalf("Failed to create file: %v", err)
		}

		if release.name == "k8s" {
			changes, err = updateK8sVersion(fileName)
			if err != nil {
				log.Fatalf("Failed to update registry version: %v", err)
			}
		} else if release.name == "gdk" {
			changes, err = updateGDKVersion(fileName)
			if err != nil {
				log.Fatalf("Failed to update registry version: %v", err)
			}
		}

		if err := CreateReleaseCommit(changes, branch, release.paths[i].Filename); err != nil {
			log.Fatalf("Failed to create release commit: %v", err)
		}
	}

	return nil
}

func CreateReleaseCommit(commit []byte, branch *gitlab.Branch, fileName string) error {
	aco := &gitlab.CommitActionOptions{
		Action:   gitlab.FileAction(gitlab.FileUpdate),
		FilePath: gitlab.String(fileName),
		Content:  gitlab.String(string(commit)),
	}
	co := &gitlab.CreateCommitOptions{
		Branch:        gitlab.String(branch.Name),
		CommitMessage: gitlab.String(fmt.Sprintf("%s to %s", release.commitMessage, strings.TrimSuffix(release.version, "-gitlab"))),
		Actions:       []*gitlab.CommitActionOptions{aco},
	}
	_, _, err := gtlb.Commits.CreateCommit(release.projectID, co)
	return err
}

func CreateReleaseMergeRequest(description string, b *gitlab.Branch, labels *gitlab.Labels) (*gitlab.MergeRequest, error) {
	mro := &gitlab.CreateMergeRequestOptions{
		Title:        gitlab.String(release.mrTitle),
		Description:  gitlab.String(description),
		SourceBranch: gitlab.String(b.Name),
		TargetBranch: gitlab.String(release.ref),
		Squash:       gitlab.Bool(true),
		Labels:       labels,
	}
	mr, _, err := gtlb.MergeRequests.CreateMergeRequest(release.projectID, mro)
	if err != nil {
		return nil, err
	}

	return mr, nil
}

func CreateReleaseBranch() (*gitlab.Branch, error) {
	bo := &gitlab.CreateBranchOptions{
		Branch: gitlab.String(fmt.Sprintf("%s-%s", release.branchName, strings.TrimSuffix(release.version, "-gitlab"))),
		Ref:    gitlab.String(release.ref),
	}
	b, _, err := gtlb.Branches.CreateBranch(release.projectID, bo)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func GetChangelog() (string, error) {
	pid := os.Getenv("CI_PROJECT_ID")
	version := os.Getenv("CI_COMMIT_TAG")

	tag, _, err := gtlb.Tags.GetTag(pid, version)
	if err != nil {
		return "", err
	}

	return tag.Commit.Message, nil
}

func SendRequestToDeps(triggerToken string) error {
	rpto := &gitlab.RunPipelineTriggerOptions{
		Ref:       gitlab.String(release.ref),
		Token:     gitlab.String(triggerToken),
		Variables: map[string]string{"DEPS_PIPELINE": "true"},
	}

	_, _, err := gtlb.PipelineTriggers.RunPipelineTrigger(release.projectID, rpto)
	if err != nil {
		return err
	}
	return nil
}
