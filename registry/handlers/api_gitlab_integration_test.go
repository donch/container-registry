//go:build integration
// +build integration

package handlers_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"testing"
	"time"

	"github.com/docker/distribution/reference"
	v1 "github.com/docker/distribution/registry/api/gitlab/v1"
	v2 "github.com/docker/distribution/registry/api/v2"
	"github.com/docker/distribution/registry/handlers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGitlabAPI_RepositoryImport_Put(t *testing.T) {
	rootDir, err := os.MkdirTemp("", "api-repository-import-")
	require.NoError(t, err)
	t.Cleanup(func() {
		os.RemoveAll(rootDir)
	})

	migrationDir := filepath.Join(rootDir, "/new")

	env := newTestEnv(t, withFSDriver(rootDir))
	defer env.Shutdown()

	env.requireDB(t)

	repoPath := "old/repo"
	tagName := "import-tag"

	// Push up a image to the old side of the registry, so we can migrate it below.
	seedRandomSchema2Manifest(t, env, repoPath, putByTag(tagName), writeToFilesystemOnly)

	env2 := newTestEnv(t, withFSDriver(rootDir), withMigrationEnabled, withMigrationRootDirectory(migrationDir))
	defer env2.Shutdown()

	// Start Repository Import.
	repoRef, err := reference.WithName(repoPath)
	require.NoError(t, err)

	importURL, err := env2.builder.BuildGitlabV1RepositoryImportURL(repoRef)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPut, importURL, nil)
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Import should start without error.
	require.Equal(t, http.StatusAccepted, resp.StatusCode)

	// Spin up a non-migartion mode env to test that the repository imported correctly.
	env3 := newTestEnv(t, withFSDriver(migrationDir))
	defer env3.Shutdown()

	tagURL := buildManifestTagURL(t, env3, repoPath, tagName)
	waitForImportSuccess(t, tagURL)

	// Subsequent calls to the same repository should not start another import.
	req, err = http.NewRequest(http.MethodPut, importURL, nil)
	require.NoError(t, err)

	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGitlabAPI_RepositoryImport_Put_ConcurrentTags(t *testing.T) {
	rootDir := t.TempDir()
	migrationDir := filepath.Join(rootDir, "/new")

	env := newTestEnv(t, withFSDriver(rootDir))
	defer env.Shutdown()

	env.requireDB(t)

	repoPath := "old/repo"
	tagTmpl := "import-tag-%d"
	tags := make([]string, 10)

	// Push up a image to the old side of the registry, so we can migrate it below.
	// Push up a series of images to the old side of the registry, so we can
	// test the importer works as expectd when launching multiple goroutines.
	for n := range tags {
		tagName := fmt.Sprintf(tagTmpl, n)
		tags[n] = tagName

		seedRandomSchema2Manifest(t, env, repoPath, putByTag(tagName), writeToFilesystemOnly)
	}

	env2 := newTestEnv(
		t, withFSDriver(rootDir),
		withMigrationEnabled,
		withMigrationRootDirectory(migrationDir),
		withMigrationTagConcurrency(5),
	)
	defer env2.Shutdown()

	// Start Repository Import.
	repoRef, err := reference.WithName(repoPath)
	require.NoError(t, err)

	importURL, err := env2.builder.BuildGitlabV1RepositoryImportURL(repoRef)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPut, importURL, nil)
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Import should start without error.
	require.Equal(t, http.StatusAccepted, resp.StatusCode)

	// Spin up a non-migartion mode env to test that the repository imported correctly.
	env3 := newTestEnv(t, withFSDriver(migrationDir))
	defer env3.Shutdown()

	for _, tag := range tags {
		tagURL := buildManifestTagURL(t, env3, repoPath, tag)
		waitForImportSuccess(t, tagURL)
	}

	// Subsequent calls to the same repository should not start another import.
	req, err = http.NewRequest(http.MethodPut, importURL, nil)
	require.NoError(t, err)

	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGitlabAPI_RepositoryImport_Put_PreImport(t *testing.T) {
	rootDir := t.TempDir()
	migrationDir := filepath.Join(rootDir, "/new")

	env := newTestEnv(t, withFSDriver(rootDir))
	defer env.Shutdown()

	env.requireDB(t)

	repoPath := "old/repo"
	tagName := "import-tag"

	// Push up a image to the old side of the registry, so we can migrate it below.
	deserializedManifest := seedRandomSchema2Manifest(t, env, repoPath, putByTag(tagName), writeToFilesystemOnly)

	env2 := newTestEnv(t, withFSDriver(rootDir), withMigrationEnabled, withMigrationRootDirectory(migrationDir))
	defer env2.Shutdown()

	// Start repository pre import.
	repoRef, err := reference.WithName(repoPath)
	require.NoError(t, err)

	importURL, err := env2.builder.BuildGitlabV1RepositoryImportURL(repoRef, url.Values{"pre": []string{"true"}})
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPut, importURL, nil)
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Pre import should start without error.
	require.Equal(t, http.StatusAccepted, resp.StatusCode)

	// Spin up a non-migartion mode env to test that the repository pre imported correctly.
	env3 := newTestEnv(t, withFSDriver(migrationDir))
	defer env3.Shutdown()

	// The manifest should only be reachable by digest.
	digestURL := buildManifestDigestURL(t, env3, repoPath, deserializedManifest)
	waitForImportSuccess(t, digestURL)

	// The tag should not have been imported.
	tagURL := buildManifestTagURL(t, env3, repoPath, tagName)
	resp, err = http.Get(tagURL)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusNotFound, resp.StatusCode)

	// Subsequent calls to the same repository should start another pre import.
	req, err = http.NewRequest(http.MethodPut, importURL, nil)
	require.NoError(t, err)

	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusAccepted, resp.StatusCode)

        // TODO: Once the notification is done, we should listen for that, rather than sleeping.
	time.Sleep(time.Second * 1)

	// Importing after pre import should succeed.
	importURL, err = env2.builder.BuildGitlabV1RepositoryImportURL(repoRef, url.Values{"pre": []string{"false"}})
	require.NoError(t, err)

	req, err = http.NewRequest(http.MethodPut, importURL, nil)
	require.NoError(t, err)

	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusAccepted, resp.StatusCode)

	waitForImportSuccess(t, tagURL)
}

func TestGitlabAPI_RepositoryImport_Put_PreImportInProgress(t *testing.T) {
	rootDir := t.TempDir()
	migrationDir := filepath.Join(rootDir, "/new")

	env := newTestEnv(t, withFSDriver(rootDir))
	defer env.Shutdown()

	env.requireDB(t)

	repoPath := "old/repo"
	tagName := "import-tag"

	// Push up a image to the old side of the registry, so we can migrate it below.
	seedRandomSchema2Manifest(t, env, repoPath, putByTag(tagName), writeToFilesystemOnly)

	env2 := newTestEnv(
		t, withFSDriver(rootDir),
		withMigrationEnabled,
		withMigrationRootDirectory(migrationDir),
		// Zero causes the importer to block, simulating a long running pre import.
		withMigrationTagConcurrency(0),
	)
	defer env2.Shutdown()

	// Start repository pre import.
	repoRef, err := reference.WithName(repoPath)
	require.NoError(t, err)

	importURL, err := env2.builder.BuildGitlabV1RepositoryImportURL(repoRef, url.Values{"pre": []string{"true"}})
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPut, importURL, nil)
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Pre import should start without error.
	require.Equal(t, http.StatusAccepted, resp.StatusCode)

	// Additonal pre import attemps should fail
	req, err = http.NewRequest(http.MethodPut, importURL, nil)
	require.NoError(t, err)

	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusTooEarly, resp.StatusCode)

	// Additonal import attemps should fail as well
	importURL, err = env2.builder.BuildGitlabV1RepositoryImportURL(repoRef)
	require.NoError(t, err)

	req, err = http.NewRequest(http.MethodPut, importURL, nil)
	require.NoError(t, err)

	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusTooEarly, resp.StatusCode)
}

func waitForImportSuccess(tb testing.TB, u string) {
	require.Eventually(tb, func() bool {
		// TODO: When https://gitlab.com/gitlab-org/container-registry/-/issues/529 is
		// finished, we should listen for the completion notification instead.
		resp, err := http.Get(u)
		if !assert.NoError(tb, err) {
			return false
		}
		defer resp.Body.Close()

		return assert.Equal(tb, http.StatusOK, resp.StatusCode)
	},
		time.Second*10,
		time.Millisecond*200,
	)
}

func TestGitlabAPI_RepositoryImport_Put_RepositoryNotPresentOnOldSide(t *testing.T) {
	rootDir, err := os.MkdirTemp("", "api-repository-import-")
	require.NoError(t, err)
	t.Cleanup(func() {
		os.RemoveAll(rootDir)
	})

	migrationDir := filepath.Join(rootDir, "/new")

	env := newTestEnv(t, withFSDriver(rootDir), withMigrationEnabled, withMigrationRootDirectory(migrationDir))
	defer env.Shutdown()

	env.requireDB(t)

	repoPath := "old/repo"

	// Start Repository Import.
	repoRef, err := reference.WithName(repoPath)
	require.NoError(t, err)

	importURL, err := env.builder.BuildGitlabV1RepositoryImportURL(repoRef)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPut, importURL, nil)
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// We should get a repository not found error
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
	checkBodyHasErrorCodes(t, "repository not found", resp, v2.ErrorCodeNameUnknown)
}

// iso8601MsFormat is a regular expression to validate ISO8601 timestamps with millisecond precision.
var iso8601MsFormat = regexp.MustCompile(`^(?:[0-9]{4}-[0-9]{2}-[0-9]{2})?(?:[ T][0-9]{2}:[0-9]{2}:[0-9]{2})?(?:[.][0-9]{3})`)

func TestGitlabAPI_Repository_Get(t *testing.T) {
	env := newTestEnv(t, disableMirrorFS)
	defer env.Shutdown()
	env.requireDB(t)

	repoName := "bar"
	repoPath := fmt.Sprintf("foo/%s", repoName)
	tagName := "latest"
	repoRef, err := reference.WithName(repoPath)
	require.NoError(t, err)

	// try to get details of non-existing repository
	u, err := env.builder.BuildGitlabV1RepositoryURL(repoRef)
	require.NoError(t, err)

	resp, err := http.Get(u)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
	checkBodyHasErrorCodes(t, "wrong response body error code", resp, v2.ErrorCodeNameUnknown)

	// try getting the details of an "empty" (no tagged artifacts) repository
	seedRandomSchema2Manifest(t, env, repoPath, putByDigest)

	u, err = env.builder.BuildGitlabV1RepositoryURL(repoRef, url.Values{
		"size": []string{"self"},
	})
	require.NoError(t, err)

	resp, err = http.Get(u)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var r handlers.RepositoryAPIResponse
	p, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	err = json.Unmarshal(p, &r)
	require.NoError(t, err)

	require.Equal(t, r.Name, repoName)
	require.Equal(t, r.Path, repoPath)
	require.Zero(t, *r.Size)
	require.NotEmpty(t, r.CreatedAt)
	require.Regexp(t, iso8601MsFormat, r.CreatedAt)
	require.Empty(t, r.UpdatedAt)

	// repeat, but before that push another image, this time tagged
	dm := seedRandomSchema2Manifest(t, env, repoPath, putByTag(tagName))
	var expectedSize int64
	for _, d := range dm.Layers() {
		expectedSize += d.Size
	}

	resp, err = http.Get(u)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	r = handlers.RepositoryAPIResponse{}
	p, err = io.ReadAll(resp.Body)
	require.NoError(t, err)
	err = json.Unmarshal(p, &r)
	require.NoError(t, err)

	require.Equal(t, r.Name, repoName)
	require.Equal(t, r.Path, repoPath)
	require.Equal(t, *r.Size, expectedSize)
	require.NotEmpty(t, r.CreatedAt)
	require.Regexp(t, iso8601MsFormat, r.CreatedAt)
	require.Empty(t, r.UpdatedAt)

	// use invalid `size` query param value
	u, err = env.builder.BuildGitlabV1RepositoryURL(repoRef, url.Values{
		"size": []string{"selfff"},
	})
	require.NoError(t, err)

	resp, err = http.Get(u)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	checkBodyHasErrorCodes(t, "wrong response body error code", resp, v1.ErrorCodeInvalidQueryParamValue)
}
