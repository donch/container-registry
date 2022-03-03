//go:build integration
// +build integration

package handlers_test

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sync"
	"testing"
	"time"

	"github.com/docker/distribution/manifest/schema2"
	"github.com/docker/distribution/reference"
	v1 "github.com/docker/distribution/registry/api/gitlab/v1"
	v2 "github.com/docker/distribution/registry/api/v2"
	"github.com/docker/distribution/registry/handlers"
	"github.com/docker/distribution/registry/internal/migration"
	"github.com/stretchr/testify/require"
)

var waitForever = time.Duration(math.MaxInt64)

func TestGitlabAPI_RepositoryImport_Get(t *testing.T) {
	rootDir := t.TempDir()
	migrationDir := filepath.Join(rootDir, "/new")

	repoPath := "old/repo"
	tagName := "import-tag"

	mockedImportNotifSrv := newMockImportNotification(t, repoPath)

	env := newTestEnv(t,
		withFSDriver(rootDir),
		withMigrationEnabled,
		withMigrationRootDirectory(migrationDir),
		withImportNotification(mockImportNotificationServer(t, mockedImportNotifSrv)))
	defer env.Shutdown()

	env.requireDB(t)

	// Push up a image to the old side of the registry, so we can migrate it below.
	seedRandomSchema2Manifest(t, env, repoPath, putByTag(tagName), writeToFilesystemOnly)

	repoRef, err := reference.WithName(repoPath)
	require.NoError(t, err)

	importURL, err := env.builder.BuildGitlabV1RepositoryImportURL(repoRef)
	require.NoError(t, err)

	// Before starting the import.
	req, err := http.NewRequest(http.MethodGet, importURL, nil)
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Import should return 404.
	require.Equal(t, http.StatusNotFound, resp.StatusCode)

	// Start Repository Import.
	req, err = http.NewRequest(http.MethodPut, importURL, nil)
	require.NoError(t, err)

	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Import should start without error.
	require.Equal(t, http.StatusAccepted, resp.StatusCode)

	mockedImportNotifSrv.waitForImportNotification(
		t, repoPath, string(migration.RepositoryStatusImportComplete), "final import completed successfully", 2*time.Second,
	)

	req, err = http.NewRequest(http.MethodGet, importURL, nil)
	require.NoError(t, err)

	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Import completed successfully.
	require.Equal(t, http.StatusOK, resp.StatusCode)

	b, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var s handlers.RepositoryImportStatus
	err = json.Unmarshal(b, &s)
	require.NoError(t, err)

	expectedStatus := handlers.RepositoryImportStatus{
		Name:   repositoryName(repoPath),
		Path:   repoPath,
		Status: migration.RepositoryStatusImportComplete,
	}

	require.Equal(t, expectedStatus, s)
}

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

	mockedImportNotifSrv := newMockImportNotification(t, repoPath)

	env2 := newTestEnv(t,
		withFSDriver(rootDir),
		withMigrationEnabled,
		withMigrationRootDirectory(migrationDir),
		withImportNotification(mockImportNotificationServer(t, mockedImportNotifSrv)))
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

	mockedImportNotifSrv.waitForImportNotification(
		t, repoPath, string(migration.RepositoryStatusImportComplete), "final import completed successfully", 2*time.Second,
	)

	// Subsequent calls to the same repository should not start another import.
	req, err = http.NewRequest(http.MethodPut, importURL, nil)
	require.NoError(t, err)

	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGitlabAPI_RepositoryPreImport_Put_PreImportTimeout(t *testing.T) {
	rootDir := t.TempDir()

	migrationDir := filepath.Join(rootDir, "/new")

	repoPath := "old/repo"
	tagName := "import-tag"

	mockedImportNotifSrv := newMockImportNotification(t, repoPath)

	env := newTestEnv(t,
		withFSDriver(rootDir),
		withMigrationEnabled,
		withMigrationRootDirectory(migrationDir),
		withMigrationPreImportTimeout(time.Millisecond),
		withImportNotification(mockImportNotificationServer(t, mockedImportNotifSrv)))
	defer env.Shutdown()

	env.requireDB(t)

	// Push up a image to the old side of the registry, so we can migrate it below.
	seedRandomSchema2Manifest(t, env, repoPath, putByTag(tagName), writeToFilesystemOnly)

	// Start Repository Import.
	repoRef, err := reference.WithName(repoPath)
	require.NoError(t, err)

	importURL, err := env.builder.BuildGitlabV1RepositoryImportURL(repoRef, url.Values{"import_type": []string{"pre"}})
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPut, importURL, nil)
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Import should start without error.
	require.Equal(t, http.StatusAccepted, resp.StatusCode)

	// pre import timed out but notification is sent anyway.
	mockedImportNotifSrv.waitForImportNotification(
		t, repoPath, string(migration.RepositoryStatusPreImportFailed),
		"updating migration status after failed pre import: updating repository: context deadline exceeded", 2*time.Second,
	)
}

func TestGitlabAPI_RepositoryImport_Put_ImportTimeout(t *testing.T) {
	rootDir := t.TempDir()

	migrationDir := filepath.Join(rootDir, "/new")

	repoPath := "old/repo"
	tagName := "import-tag"

	mockedImportNotifSrv := newMockImportNotification(t, repoPath)

	env := newTestEnv(t,
		withFSDriver(rootDir),
		withMigrationEnabled,
		withMigrationRootDirectory(migrationDir),
		withMigrationImportTimeout(time.Millisecond),
		withImportNotification(mockImportNotificationServer(t, mockedImportNotifSrv)))
	defer env.Shutdown()

	env.requireDB(t)

	// Push up a image to the old side of the registry, so we can migrate it below.
	seedRandomSchema2Manifest(t, env, repoPath, putByTag(tagName), writeToFilesystemOnly)

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

	// Import should start without error.
	require.Equal(t, http.StatusAccepted, resp.StatusCode)

	// final import timed out but notification is sent anyway.
	mockedImportNotifSrv.waitForImportNotification(
		t, repoPath, string(migration.RepositoryStatusImportFailed),
		"updating migration status after failed final import: updating repository: context deadline exceeded", 2*time.Second,
	)
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

	mockedImportNotifSrv := newMockImportNotification(t, repoPath)

	env2 := newTestEnv(
		t, withFSDriver(rootDir),
		withMigrationEnabled,
		withMigrationRootDirectory(migrationDir),
		withMigrationTagConcurrency(5),
		withImportNotification(mockImportNotificationServer(t, mockedImportNotifSrv)),
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

	mockedImportNotifSrv.waitForImportNotification(
		t, repoPath, string(migration.RepositoryStatusImportComplete), "final import completed successfully", 2*time.Second,
	)

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
	seedRandomSchema2Manifest(t, env, repoPath, putByTag(tagName), writeToFilesystemOnly)

	mockedImportNotifSrv := newMockImportNotification(t, repoPath)

	env2 := newTestEnv(t,
		withFSDriver(rootDir),
		withMigrationEnabled,
		withMigrationRootDirectory(migrationDir),
		withImportNotification(mockImportNotificationServer(t, mockedImportNotifSrv)))
	defer env2.Shutdown()

	// Start repository pre import.
	repoRef, err := reference.WithName(repoPath)
	require.NoError(t, err)

	importURL, err := env2.builder.BuildGitlabV1RepositoryImportURL(repoRef, url.Values{"import_type": []string{"pre"}})
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

	mockedImportNotifSrv.waitForImportNotification(
		t, repoPath, string(migration.RepositoryStatusPreImportComplete), "pre import completed successfully", 2*time.Second,
	)

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

	mockedImportNotifSrv.waitForImportNotification(
		t, repoPath, string(migration.RepositoryStatusPreImportComplete), "pre import completed successfully", 2*time.Second,
	)

	// Final import after pre import should succeed.
	importURL, err = env2.builder.BuildGitlabV1RepositoryImportURL(repoRef, url.Values{"import_type": []string{"final"}})
	require.NoError(t, err)

	req, err = http.NewRequest(http.MethodPut, importURL, nil)
	require.NoError(t, err)

	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusAccepted, resp.StatusCode)

	mockedImportNotifSrv.waitForImportNotification(
		t, repoPath, string(migration.RepositoryStatusImportComplete), "final import completed successfully", 2*time.Second,
	)
}

func TestGitlabAPI_RepositoryImport_PreImportInProgress(t *testing.T) {
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
		// Simulate a long running import.
		withMigrationTestSlowImport(waitForever),
	)
	defer env2.Shutdown()

	// Start repository pre import.
	repoRef, err := reference.WithName(repoPath)
	require.NoError(t, err)

	importURL, err := env2.builder.BuildGitlabV1RepositoryImportURL(repoRef, url.Values{"import_type": []string{"pre"}})
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

	// Import GET should return appropriate status
	req, err = http.NewRequest(http.MethodGet, importURL, nil)
	require.NoError(t, err)

	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	b, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var s handlers.RepositoryImportStatus
	err = json.Unmarshal(b, &s)
	require.NoError(t, err)

	expectedStatus := handlers.RepositoryImportStatus{
		Name:   repositoryName(repoPath),
		Path:   repoPath,
		Status: migration.RepositoryStatusPreImportInProgress,
	}

	require.Equal(t, expectedStatus, s)
}

func TestGitlabAPI_RepositoryImport_ImportInProgress(t *testing.T) {
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
		// Simulate a long running import.
		withMigrationTestSlowImport(waitForever),
	)
	defer env2.Shutdown()

	// Start repository import.
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

	// Additonal import attemps should fail
	req, err = http.NewRequest(http.MethodPut, importURL, nil)
	require.NoError(t, err)

	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusConflict, resp.StatusCode)

	// Pre import attemps should fail as well
	importURL, err = env2.builder.BuildGitlabV1RepositoryImportURL(repoRef, url.Values{"import_type": []string{"pre"}})
	require.NoError(t, err)

	req, err = http.NewRequest(http.MethodPut, importURL, nil)
	require.NoError(t, err)

	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusConflict, resp.StatusCode)

	// Import GET should return appropriate status
	req, err = http.NewRequest(http.MethodGet, importURL, nil)
	require.NoError(t, err)

	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	b, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var s handlers.RepositoryImportStatus
	err = json.Unmarshal(b, &s)
	require.NoError(t, err)

	expectedStatus := handlers.RepositoryImportStatus{
		Name:   repositoryName(repoPath),
		Path:   repoPath,
		Status: migration.RepositoryStatusImportInProgress,
	}

	require.Equal(t, expectedStatus, s)
}

func TestGitlabAPI_RepositoryImport_Put_PreImportFailed(t *testing.T) {
	rootDir := t.TempDir()
	migrationDir := filepath.Join(rootDir, "/new")

	env := newTestEnv(t, withFSDriver(rootDir))
	defer env.Shutdown()

	env.requireDB(t)

	repoPath := "notags/repo"

	// Push up a image to the old side of the registry, but do not push any tags,
	// the pre import will start without error, but the actual pre import will fail.
	seedRandomSchema2Manifest(t, env, repoPath, putByDigest, writeToFilesystemOnly)

	mockedImportNotifSrv := newMockImportNotification(t, repoPath)

	env2 := newTestEnv(t,
		withFSDriver(rootDir),
		withMigrationEnabled,
		withMigrationRootDirectory(migrationDir),
		withImportNotification(mockImportNotificationServer(t, mockedImportNotifSrv)))
	defer env2.Shutdown()

	repoRef, err := reference.WithName(repoPath)
	require.NoError(t, err)

	preImportURL, err := env2.builder.BuildGitlabV1RepositoryImportURL(repoRef, url.Values{"import_type": []string{"pre"}})
	require.NoError(t, err)

	importURL, err := env2.builder.BuildGitlabV1RepositoryImportURL(repoRef)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPut, preImportURL, nil)
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusAccepted, resp.StatusCode)

	mockedImportNotifSrv.waitForImportNotification(
		t, repoPath, string(migration.RepositoryStatusPreImportFailed),
		"pre importing tagged manifests: reading tags: unknown repository name=notags/repo", 2*time.Second,
	)

	// Subsequent import attempts fail.
	req, err = http.NewRequest(http.MethodPut, importURL, nil)
	require.NoError(t, err)

	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusFailedDependency, resp.StatusCode)

	// Starting a pre import after a failed pre import attempt should succeed.
	req, err = http.NewRequest(http.MethodPut, preImportURL, nil)
	require.NoError(t, err)

	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusAccepted, resp.StatusCode)

	// It's good to wait here as well, otherwise the test env will close the
	// connection to the database before the import goroutine is finished, logging
	// an error that could be misleading.
	mockedImportNotifSrv.waitForImportNotification(
		t, repoPath, string(migration.RepositoryStatusPreImportFailed),
		"pre importing tagged manifests: reading tags: unknown repository name=notags/repo", 2*time.Second,
	)

	req, err = http.NewRequest(http.MethodGet, importURL, nil)
	require.NoError(t, err)

	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	b, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var s handlers.RepositoryImportStatus
	err = json.Unmarshal(b, &s)
	require.NoError(t, err)

	expectedStatus := handlers.RepositoryImportStatus{
		Name:   repositoryName(repoPath),
		Path:   repoPath,
		Status: migration.RepositoryStatusPreImportFailed,
	}

	require.Equal(t, expectedStatus, s)
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

func TestGitlabAPI_RepositoryImport_Migration_PreImportInProgress(t *testing.T) {
	rootDir := t.TempDir()
	migrationDir := filepath.Join(rootDir, "/new")

	env := newTestEnv(
		t, withFSDriver(rootDir),
		withMigrationEnabled,
		withMigrationRootDirectory(migrationDir),
		// Simulate a long running pre import.
		withMigrationTestSlowImport(waitForever),
	)
	defer env.Shutdown()

	env.requireDB(t)

	repoPath := "old/repo"
	tagName := "import-tag"

	// Push up a image to the old side of the registry, so we can migrate it below.
	seedRandomSchema2Manifest(t, env, repoPath, putByTag(tagName), writeToFilesystemOnly)

	// Start repository pre import.
	repoRef, err := reference.WithName(repoPath)
	require.NoError(t, err)

	importURL, err := env.builder.BuildGitlabV1RepositoryImportURL(repoRef, url.Values{"pre": []string{"true"}})
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPut, importURL, nil)
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Pre import should start without error.
	require.Equal(t, http.StatusAccepted, resp.StatusCode)

	// Pull the Manifest and ensure that the migration path header indicates
	// that the old code path is still taken during pre import for reads.
	tagURL := buildManifestTagURL(t, env, repoPath, tagName)
	require.NoError(t, err)

	resp, err = http.Get(tagURL)
	require.NoError(t, err)

	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Equal(t, migration.OldCodePath.String(), resp.Header.Get("Gitlab-Migration-Path"))

	// Push up a image to the same repository and ensure that the migration path header
	// indicates that the old code path is still taken during pre import for writes.
	deserializedManifest := seedRandomSchema2Manifest(t, env, repoPath)
	manifestDigestURL := buildManifestDigestURL(t, env, repoPath, deserializedManifest)

	resp = putManifest(t, "putting manifest no error", manifestDigestURL, schema2.MediaTypeManifest, deserializedManifest.Manifest)
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	require.Equal(t, migration.OldCodePath.String(), resp.Header.Get("Gitlab-Migration-Path"))
}

func TestGitlabAPI_RepositoryImport_Migration_ImportInProgress(t *testing.T) {
	rootDir := t.TempDir()
	migrationDir := filepath.Join(rootDir, "/new")

	env := newTestEnv(
		t, withFSDriver(rootDir),
		withMigrationEnabled,
		withMigrationRootDirectory(migrationDir),
		// Simulate a long running import.
		withMigrationTestSlowImport(waitForever),
	)
	defer env.Shutdown()

	env.requireDB(t)

	repoPath := "old/repo"
	tagName := "import-tag"

	// Push up a image to the old side of the registry, so we can migrate it below.
	seedRandomSchema2Manifest(t, env, repoPath, putByTag(tagName), writeToFilesystemOnly)

	// Start repository import.
	repoRef, err := reference.WithName(repoPath)
	require.NoError(t, err)

	importURL, err := env.builder.BuildGitlabV1RepositoryImportURL(repoRef)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPut, importURL, nil)
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Import should start without error.
	require.Equal(t, http.StatusAccepted, resp.StatusCode)

	// Pull the Manifest and ensure that the migration path header indicates
	// that the old code path is still taken during import for reads.
	tagURL := buildManifestTagURL(t, env, repoPath, tagName)
	require.NoError(t, err)

	resp, err = http.Get(tagURL)
	require.NoError(t, err)

	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Equal(t, migration.OldCodePath.String(), resp.Header.Get("Gitlab-Migration-Path"))

	// Push up a image to the same repository and ensure that the migration path header
	// indicates that the old code path is still taken during import for writes.
	deserializedManifest := seedRandomSchema2Manifest(t, env, repoPath)
	manifestDigestURL := buildManifestDigestURL(t, env, repoPath, deserializedManifest)

	resp = putManifest(t, "putting manifest no error", manifestDigestURL, schema2.MediaTypeManifest, deserializedManifest.Manifest)
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	require.Equal(t, migration.OldCodePath.String(), resp.Header.Get("Gitlab-Migration-Path"))
}

func TestGitlabAPI_RepositoryImport_Migration_PreImportComplete(t *testing.T) {
	rootDir := t.TempDir()
	migrationDir := filepath.Join(rootDir, "/new")

	repoPath := "old/repo"
	tagName := "import-tag"

	mockedImportNotifSrv := newMockImportNotification(t, repoPath)

	env := newTestEnv(t,
		withFSDriver(rootDir),
		withMigrationEnabled,
		withMigrationRootDirectory(migrationDir),
		withImportNotification(mockImportNotificationServer(t, mockedImportNotifSrv)),
	)
	defer env.Shutdown()

	env.requireDB(t)

	// Push up a image to the old side of the registry, so we can migrate it below.
	seedRandomSchema2Manifest(t, env, repoPath, putByTag(tagName), writeToFilesystemOnly)

	// Start repository pre import.
	repoRef, err := reference.WithName(repoPath)
	require.NoError(t, err)

	importURL, err := env.builder.BuildGitlabV1RepositoryImportURL(repoRef, url.Values{"import_type": []string{"pre"}})
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPut, importURL, nil)
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Pre import should complete without error.
	require.Equal(t, http.StatusAccepted, resp.StatusCode)

	mockedImportNotifSrv.waitForImportNotification(
		t, repoPath, string(migration.RepositoryStatusPreImportComplete), "pre import completed successfully", 2*time.Second,
	)

	// Pull the Manifest and ensure that the migration path header indicates
	// that the old code path is still taken after pre import for reads.
	tagURL := buildManifestTagURL(t, env, repoPath, tagName)
	require.NoError(t, err)

	resp, err = http.Get(tagURL)
	require.NoError(t, err)

	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Equal(t, migration.OldCodePath.String(), resp.Header.Get("Gitlab-Migration-Path"))

	// Push up a image to the same repository and ensure that the migration path header
	// indicates that the old code path is still taken after pre import for writes.
	deserializedManifest := seedRandomSchema2Manifest(t, env, repoPath)
	manifestDigestURL := buildManifestDigestURL(t, env, repoPath, deserializedManifest)

	resp = putManifest(t, "putting manifest no error", manifestDigestURL, schema2.MediaTypeManifest, deserializedManifest.Manifest)
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	require.Equal(t, migration.OldCodePath.String(), resp.Header.Get("Gitlab-Migration-Path"))
}

func TestGitlabAPI_RepositoryImport_Migration_ImportComplete(t *testing.T) {
	rootDir := t.TempDir()
	migrationDir := filepath.Join(rootDir, "/new")

	repoPath := "old/repo"
	tagName := "import-tag"

	mockedImportNotifSrv := newMockImportNotification(t, repoPath)

	env := newTestEnv(t,
		withFSDriver(rootDir),
		withMigrationEnabled,
		withMigrationRootDirectory(migrationDir),
		withImportNotification(mockImportNotificationServer(t, mockedImportNotifSrv)),
	)
	defer env.Shutdown()

	env.requireDB(t)

	// Push up a image to the old side of the registry, so we can migrate it below.
	seedRandomSchema2Manifest(t, env, repoPath, putByTag(tagName), writeToFilesystemOnly)

	// Start repository import.
	repoRef, err := reference.WithName(repoPath)
	require.NoError(t, err)

	importURL, err := env.builder.BuildGitlabV1RepositoryImportURL(repoRef)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPut, importURL, nil)
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Import should complete without error.
	require.Equal(t, http.StatusAccepted, resp.StatusCode)

	mockedImportNotifSrv.waitForImportNotification(
		t, repoPath, string(migration.RepositoryStatusImportComplete), "final import completed successfully", 2*time.Second,
	)

	// Pull the Manifest and ensure that the migration path header indicates
	// that the new code path is taken after import for reads.
	tagURL := buildManifestTagURL(t, env, repoPath, tagName)
	require.NoError(t, err)

	resp, err = http.Get(tagURL)
	require.NoError(t, err)

	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Equal(t, migration.NewCodePath.String(), resp.Header.Get("Gitlab-Migration-Path"))

	// Push up a image to the same repository and ensure that the migration path header
	// indicates that the new code path is taken after import for writes.
	deserializedManifest := seedRandomSchema2Manifest(t, env, repoPath)
	manifestDigestURL := buildManifestDigestURL(t, env, repoPath, deserializedManifest)

	resp = putManifest(t, "putting manifest no error", manifestDigestURL, schema2.MediaTypeManifest, deserializedManifest.Manifest)
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	require.Equal(t, migration.NewCodePath.String(), resp.Header.Get("Gitlab-Migration-Path"))
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

func TestGitlabAPI_RepositoryImport_MaxConcurrentImports(t *testing.T) {
	rootDir := t.TempDir()

	migrationDir := filepath.Join(rootDir, "/new")

	repoPathTemplate := "old/repo-%d"
	tagName := "import-tag"
	repoCount := 5

	allRepoPaths := generateOldRepoPaths(t, repoPathTemplate, repoCount)

	mockedImportNotifSrv := newMockImportNotification(t, allRepoPaths...)

	env := newTestEnv(t,
		withFSDriver(rootDir),
		withMigrationEnabled,
		withMigrationRootDirectory(migrationDir),
		withImportNotification(mockImportNotificationServer(t, mockedImportNotifSrv)),
		// only allow a maximum of 3 imports at a time
		withMigrationMaxConcurrentImports(3))
	t.Cleanup(env.Shutdown)

	env.requireDB(t)

	seedMultipleFSManifestsWithTag(t, env, tagName, allRepoPaths)

	attemptImportFn := func(count int, expectedStatus int, waitForNotif bool) {
		repoPath := fmt.Sprintf(repoPathTemplate, count)
		// Start Repository Import.
		repoRef, err := reference.WithName(repoPath)
		require.NoError(t, err)

		importURL, err := env.builder.BuildGitlabV1RepositoryImportURL(repoRef, url.Values{"import_type": []string{"pre"}})
		require.NoError(t, err)

		req, err := http.NewRequest(http.MethodPut, importURL, nil)
		require.NoError(t, err)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equalf(t, expectedStatus, resp.StatusCode, "repo path: %q", repoPath)

		if waitForNotif {
			mockedImportNotifSrv.waitForImportNotification(
				t, repoPath, string(migration.RepositoryStatusPreImportComplete), "pre import completed successfully", 5*time.Second,
			)
		}
	}

	wg := &sync.WaitGroup{}

	for i := 0; i < repoCount; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			if i < 3 {
				// expect first 3 imports to succeed
				attemptImportFn(i, http.StatusAccepted, true)
			} else {
				// let the first 3 request go first
				time.Sleep(10 * time.Millisecond)
				attemptImportFn(i, http.StatusTooManyRequests, false)
			}
		}(i)
	}

	wg.Wait()

	wg2 := &sync.WaitGroup{}

	// attempt to pre import again should succeed
	for i := 3; i < 5; i++ {
		wg2.Add(1)
		go func(i int) {
			defer wg2.Done()

			attemptImportFn(i, http.StatusAccepted, true)
		}(i)
	}

	wg2.Wait()
}

func TestGitlabAPI_RepositoryImport_MaxConcurrentImports_IsZero(t *testing.T) {
	rootDir := t.TempDir()

	migrationDir := filepath.Join(rootDir, "/new")

	env := newTestEnv(t,
		withFSDriver(rootDir),
		withMigrationEnabled,
		withMigrationRootDirectory(migrationDir),
		// Explicitly set to 0 so no imports would be allowed.
		withMigrationMaxConcurrentImports(0),
	)
	t.Cleanup(env.Shutdown)

	env.requireDB(t)

	repoPath := "old/repo"
	tagName := "import-tag"

	// Push up a image to the old side of the registry, so we can migrate it below.
	seedRandomSchema2Manifest(t, env, repoPath, putByTag(tagName), writeToFilesystemOnly)

	repoRef, err := reference.WithName(repoPath)
	require.NoError(t, err)

	attemptImportFn := func(preImport bool) {
		urlValues := url.Values{}
		if preImport {
			urlValues.Set("import_type", "pre")
		}

		importURL, err := env.builder.BuildGitlabV1RepositoryImportURL(repoRef, urlValues)
		require.NoError(t, err)

		req, err := http.NewRequest(http.MethodPut, importURL, nil)
		require.NoError(t, err)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		// should be rate limited
		require.Equal(t, http.StatusTooManyRequests, resp.StatusCode)
	}

	// attempt pre import
	attemptImportFn(true)

	// attempt import
	attemptImportFn(false)
}

func TestGitlabAPI_RepositoryImport_MaxConcurrentImports_OneByOne(t *testing.T) {
	rootDir := t.TempDir()

	migrationDir := filepath.Join(rootDir, "/new")

	repoPathTemplate := "old/repo-%d"
	tagName := "import-tag"

	allRepoPaths := generateOldRepoPaths(t, repoPathTemplate, 2)

	mockedImportNotifSrv := newMockImportNotification(t, allRepoPaths...)

	env := newTestEnv(t,
		withFSDriver(rootDir),
		withMigrationEnabled,
		withMigrationRootDirectory(migrationDir),
		withImportNotification(mockImportNotificationServer(t, mockedImportNotifSrv)),
		// only allow 1 import at a time
		withMigrationMaxConcurrentImports(1))
	t.Cleanup(env.Shutdown)

	env.requireDB(t)

	seedMultipleFSManifestsWithTag(t, env, tagName, allRepoPaths)

	repoPath1 := fmt.Sprintf(repoPathTemplate, 0)
	repoRef1, err := reference.WithName(repoPath1)
	require.NoError(t, err)

	importURL, err := env.builder.BuildGitlabV1RepositoryImportURL(repoRef1, url.Values{"import_type": []string{"pre"}})
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPut, importURL, nil)
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// repoPath1 should be accepted
	require.Equal(t, http.StatusAccepted, resp.StatusCode)

	repoPath2 := fmt.Sprintf(repoPathTemplate, 1)

	repoRef2, err := reference.WithName(repoPath2)
	require.NoError(t, err)

	importURL2, err := env.builder.BuildGitlabV1RepositoryImportURL(repoRef2, url.Values{"import_type": []string{"pre"}})
	require.NoError(t, err)

	req2, err := http.NewRequest(http.MethodPut, importURL2, nil)
	require.NoError(t, err)

	resp2, err := http.DefaultClient.Do(req2)
	require.NoError(t, err)
	defer resp2.Body.Close()

	// repoPath2 should be rate limited
	require.Equal(t, http.StatusTooManyRequests, resp2.StatusCode)

	// wait for repoPath1 import notification
	mockedImportNotifSrv.waitForImportNotification(
		t, repoPath1, string(migration.RepositoryStatusPreImportComplete), "pre import completed successfully", 2*time.Second,
	)

	// attempt second import for repoPath2 should succeed
	req3, err := http.NewRequest(http.MethodPut, importURL2, nil)
	require.NoError(t, err)

	resp3, err := http.DefaultClient.Do(req3)
	require.NoError(t, err)
	defer resp3.Body.Close()

	// should be accepted
	require.Equal(t, http.StatusAccepted, resp3.StatusCode)

	// second import should succeed
	mockedImportNotifSrv.waitForImportNotification(
		t, repoPath2, string(migration.RepositoryStatusPreImportComplete), "pre import completed successfully", 2*time.Second,
	)
}

func TestGitlabAPI_RepositoryImport_MaxConcurrentImports_ErrorShouldNotBlockLimits(t *testing.T) {
	rootDir := t.TempDir()

	migrationDir := filepath.Join(rootDir, "/new")

	repoPathTemplate := "old/repo-%d"
	tagName := "import-tag"

	allRepoPaths := generateOldRepoPaths(t, repoPathTemplate, 2)

	mockedImportNotifSrv := newMockImportNotification(t, allRepoPaths...)

	env := newTestEnv(t,
		withFSDriver(rootDir),
		withMigrationEnabled,
		withMigrationRootDirectory(migrationDir),
		withImportNotification(mockImportNotificationServer(t, mockedImportNotifSrv)),
		// only allow 1 import at a time
		withMigrationMaxConcurrentImports(1))
	t.Cleanup(env.Shutdown)

	env.requireDB(t)

	seedMultipleFSManifestsWithTag(t, env, tagName, allRepoPaths)

	repoPath1 := fmt.Sprintf(repoPathTemplate, 0)
	repoRef1, err := reference.WithName(repoPath1)
	require.NoError(t, err)

	// using an invalid value for `import_type` should raise an error and not allow the import to proceed
	importURL, err := env.builder.BuildGitlabV1RepositoryImportURL(repoRef1, url.Values{"import_type": []string{"invalid"}})
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPut, importURL, nil)
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// repoPath1 fail
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	checkBodyHasErrorCodes(t, "wrong response body error code", resp, v1.ErrorCodeInvalidQueryParamValue)

	repoPath2 := fmt.Sprintf(repoPathTemplate, 1)

	repoRef2, err := reference.WithName(repoPath2)
	require.NoError(t, err)

	importURL2, err := env.builder.BuildGitlabV1RepositoryImportURL(repoRef2, url.Values{"import_type": []string{"pre"}})
	require.NoError(t, err)

	req2, err := http.NewRequest(http.MethodPut, importURL2, nil)
	require.NoError(t, err)

	resp2, err := http.DefaultClient.Do(req2)
	require.NoError(t, err)
	defer resp2.Body.Close()

	// repoPath2 should be accepted because the first request failed
	require.Equal(t, http.StatusAccepted, resp2.StatusCode)

	// second import should succeed
	mockedImportNotifSrv.waitForImportNotification(
		t, repoPath2, string(migration.RepositoryStatusPreImportComplete), "pre import completed successfully", 2*time.Second,
	)
}
