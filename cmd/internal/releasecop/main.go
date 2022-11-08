package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/xanzy/go-gitlab"
)

var (
	targetProjectID = flag.String(
		"target-project-id",
		"",
		"The project ID of the target project to monitor released items",
	)
	gitlabAuthToken = flag.String(
		"gitlab-auth-token",
		"",
		"PAT or $CI_JOB_TOKEN to use to authenticate against the GitLab API",
	)
	webhookURL = flag.String(
		"slack-webhook",
		"",
		"Slack incoming webhook URL to notify updates to",
	)
)

type Release struct {
	name           string
	projectID      int
	version        string
	authorUsername string
	branchName     string
	whoUpdates     string
}

type SlackClient struct {
	WebHookUrl string
	UserName   string
	Channel    string
}

type SimpleSlackRequest struct {
	Text      string
	IconEmoji string
}

type SlackMessage struct {
	Username  string `json:"username,omitempty"`
	IconEmoji string `json:"icon_emoji,omitempty"`
	Channel   string `json:"channel,omitempty"`
	Text      string `json:"text,omitempty"`
}

func main() {
	flag.Parse()

	gitlabClient, err := gitlab.NewClient(*gitlabAuthToken)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	latestRelease, err := GetLatestReleaseTag(gitlabClient, *targetProjectID)
	if err != nil || latestRelease == "" {
		log.Fatalf("Failed to find latest Release: %v", err)
	}

	releases := []Release{
		{
			name:           "cng",
			projectID:      4359271,
			version:        latestRelease,
			authorUsername: "gitlab-dependency-bot",
			whoUpdates:     "dependency bot",
		},
		{
			name:           "charts",
			projectID:      3828396,
			version:        latestRelease,
			authorUsername: "gitlab-dependency-bot",
			whoUpdates:     "dependency bot",
		},
		{
			name:           "omnibus",
			projectID:      20699,
			version:        latestRelease,
			authorUsername: "gitlab-dependency-bot",
			whoUpdates:     "dependency bot",
		},
		{
			name:       "k8s-workloads",
			projectID:  12547113,
			version:    latestRelease,
			branchName: "bump-registry-version-pre-gstg",
			whoUpdates: "internal tool",
		},
		{
			name:       "k8s-workloads",
			projectID:  12547113,
			version:    latestRelease,
			branchName: "bump-registry-version-prod",
			whoUpdates: "internal tool",
		},
	}

	releaseIssue, err := GetReleaseIssue(gitlabClient, *targetProjectID, latestRelease)
	if err != nil || releaseIssue == nil {
		log.Fatalf(
			"Failed to find an issue with search term `Release %v`: %v",
			strings.TrimSuffix(latestRelease, "-gitlab"),
			err,
		)
	}

	if releaseIssue.State == "closed" {
		fmt.Printf("There is already a release issue at %s but it is closed. Exiting...", releaseIssue.WebURL)
		return
	}

	for _, release := range releases {
		if isReleaseMRLinked(gitlabClient, release, releaseIssue) {
			continue
		} else {
			err = LinkMRToRelease(gitlabClient, release, releaseIssue)
			if err != nil {
				log.Fatalf("Failed to link %v version bump merge request to the release issue: %v", release.name, err)
			}
		}
	}

	hasNoRelatedMRs, err := HasNoRelatedMRs(gitlabClient, releaseIssue)
	if err != nil {
		log.Fatalf("Failed to find related merge requests: %v", err)
	}

	hasOpenMRs, err := HasRelatedMRsOpened(gitlabClient, releaseIssue)
	if err != nil {
		log.Fatalf("Failed to find related merge requests: %v", err)
	}

	if hasOpenMRs || hasNoRelatedMRs {
		fmt.Printf("Found no related merge requests or still to be merged. Exiting...")
		return
	}

	hasVerificationLabel, err := hasLabel(gitlabClient, releaseIssue, "~workflow::verification")
	if err != nil {
		log.Fatalf("Failed to find a label on release issue: %v", err)
	}

	if hasVerificationLabel {
		fmt.Printf("Awaiting human verification on Release %s. Exiting...", latestRelease)
		return
	} else {
		err = ApplyLabelOnReleaseIssue(gitlabClient, releaseIssue, "~workflow::verification")
		if err != nil {
			log.Fatalf("Failed to update the release issue: %v ", err)
		}
	}

	err = PostNoteOnReleaseIssue(
		gitlabClient,
		releaseIssue,
		"Hey human 👋 It looks like all version bumps MRs got merged 🎉 I need your help to verify these, apply the ~workflow::production label and close this issue",
	)
	if err != nil {
		log.Fatalf("Failed to post note on the release issue: %v", err)
	}

	fmt.Printf("Sucessfully posted a note at %s\n", releaseIssue.WebURL)

	messageSlack := fmt.Sprintf("Hey human 👋 It looks like all version bumps MRs got merged 🎉 I need your help to verify these, apply the ~workflow::production label and close the <%s|release issue>", releaseIssue.WebURL)
	notifyInSlack(messageSlack, "#g_container-registry", *webhookURL)
}

func GetLatestReleaseTag(gitlabClient *gitlab.Client, pid string) (string, error) {
	releases, _, err := gitlabClient.Releases.ListReleases(pid, &gitlab.ListReleasesOptions{})
	if err != nil {
		return "", err
	}

	if len(releases) == 0 {
		var errNoReleasesFound = errors.New("could not find any releases")
		return "", errNoReleasesFound
	}

	return releases[0].TagName, nil
}

func GetReleaseIssue(gitlabClient *gitlab.Client, pid string, latestRelease string) (*gitlab.Issue, error) {
	opts := &gitlab.ListProjectIssuesOptions{
		Search: gitlab.String(fmt.Sprintf(
			"Release %s",
			strings.TrimSuffix(latestRelease, "-gitlab")),
		),
	}

	issues, _, err := gitlabClient.Issues.ListProjectIssues(pid, opts)
	if err != nil {
		return nil, err
	}

	if len(issues) == 0 {
		return nil, nil
	}

	return issues[0], nil
}

func HasRelatedMRsOpened(gitlabClient *gitlab.Client, releaseIssue *gitlab.Issue) (bool, error) {
	mrs, _, err := gitlabClient.Issues.ListMergeRequestsRelatedToIssue(*targetProjectID, releaseIssue.IID, &gitlab.ListMergeRequestsRelatedToIssueOptions{})
	if err != nil {
		return false, err
	}

	for i := range mrs {
		if mrs[i].State == "opened" {
			return true, nil
		}
	}

	return false, nil
}

func HasNoRelatedMRs(gitlabClient *gitlab.Client, releaseIssue *gitlab.Issue) (bool, error) {
	mrs, _, err := gitlabClient.Issues.ListMergeRequestsRelatedToIssue(*targetProjectID, releaseIssue.IID, &gitlab.ListMergeRequestsRelatedToIssueOptions{})
	if err != nil {
		return false, err
	}

	if len(mrs) == 0 {
		return true, nil
	}

	return false, nil
}

func hasLabel(gitlabClient *gitlab.Client, releaseIssue *gitlab.Issue, label string) (bool, error) {
	issue, _, err := gitlabClient.Issues.GetIssue(*targetProjectID, releaseIssue.IID)
	if err != nil {
		return false, err
	}

	if contains(issue.Labels, label) {
		return true, nil
	}

	return false, nil
}

func ApplyLabelOnReleaseIssue(gitlabClient *gitlab.Client, releaseIssue *gitlab.Issue, label string) error {
	opts := &gitlab.UpdateIssueOptions{
		Labels: &gitlab.Labels{label},
	}
	_, _, err := gitlabClient.Issues.UpdateIssue(*targetProjectID, releaseIssue.IID, opts)
	if err != nil {
		return err
	}

	return nil
}

func PostNoteOnReleaseIssue(gitlabClient *gitlab.Client, releaseIssue *gitlab.Issue, note string) error {
	opts := &gitlab.CreateIssueNoteOptions{
		Body: gitlab.String(
			note,
		),
	}

	_, _, err := gitlabClient.Notes.CreateIssueNote(*targetProjectID, releaseIssue.IID, opts)
	if err != nil {
		return err
	}

	return nil
}

func LinkMRToRelease(gitlabClient *gitlab.Client, release Release, releaseIssue *gitlab.Issue) error {
	opts := &gitlab.ListProjectMergeRequestsOptions{}
	if release.whoUpdates == "dependency bot" {
		opts = &gitlab.ListProjectMergeRequestsOptions{
			State:          gitlab.String("opened"),
			Search:         gitlab.String(fmt.Sprintf("Update gitlab-org/container-registry %s", release.version)),
			AuthorUsername: gitlab.String(release.authorUsername),
		}
	} else {
		opts = &gitlab.ListProjectMergeRequestsOptions{
			State:        gitlab.String("opened"),
			Search:       gitlab.String(fmt.Sprintf("Bump Container Registry to %s", release.version)),
			SourceBranch: gitlab.String(release.branchName),
		}
	}
	mrs, _, err := gitlabClient.MergeRequests.ListProjectMergeRequests(release.projectID, opts)
	if err != nil {
		return err
	}

	if len(mrs) == 0 {
		fmt.Printf("Found no merge requests in %s matching criteria\n", release.name)
		return nil
	}
	mr := mrs[0]

	mrOpts := &gitlab.CreateMergeRequestNoteOptions{
		Body: gitlab.String(fmt.Sprintf("Relating to %s", releaseIssue.WebURL)),
	}
	_, _, err = gitlabClient.Notes.CreateMergeRequestNote(release.projectID, mr.IID, mrOpts)
	if err != nil {
		return err
	}

	return nil
}

func isReleaseMRLinked(gitlabClient *gitlab.Client, release Release, releaseIssue *gitlab.Issue) bool {
	mrs, _, err := gitlabClient.Issues.ListMergeRequestsRelatedToIssue(*targetProjectID, releaseIssue.IID, &gitlab.ListMergeRequestsRelatedToIssueOptions{})
	if err != nil {
		return false
	}

	for i := range mrs {
		if release.name == "k8s-workloads" {
			if strings.Contains(mrs[i].SourceBranch, release.branchName) {
				return true
			}
		} else if strings.Contains(mrs[i].WebURL, release.name) {
			return true
		}
	}
	return false
}

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}

func (sc SlackClient) SendSlackNotification(sr SimpleSlackRequest) error {
	slackRequest := SlackMessage{
		Text:      sr.Text,
		Username:  sc.UserName,
		IconEmoji: sr.IconEmoji,
		Channel:   sc.Channel,
	}
	return sc.sendHttpRequest(slackRequest)
}

func (sc SlackClient) sendHttpRequest(slackRequest SlackMessage) error {
	slackBody, _ := json.Marshal(slackRequest)
	req, err := http.NewRequest(http.MethodPost, sc.WebHookUrl, bytes.NewBuffer(slackBody))
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		return err
	}
	if buf.String() != "ok" {
		return errors.New("Non-ok response returned from Slack")
	}
	return nil
}

func notifyInSlack(message string, channel string, webhookURL string) {
	sc := SlackClient{
		WebHookUrl: webhookURL,
		UserName:   "release_cop",
		Channel:    channel,
	}
	sr := SimpleSlackRequest{
		Text:      message,
		IconEmoji: ":robot_face:",
	}
	err := sc.SendSlackNotification(sr)
	if err != nil {
		log.Fatal(err)
	}
}
