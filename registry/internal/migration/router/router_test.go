package router

import (
	"context"
	"fmt"
	"testing"

	"github.com/docker/distribution"
	"github.com/docker/distribution/reference"
	"github.com/docker/distribution/registry/datastore/models"
	"github.com/docker/distribution/registry/internal/migration"
	"github.com/docker/distribution/registry/storage"
	"github.com/docker/distribution/registry/storage/driver/inmemory"
	"github.com/docker/distribution/testutil"
	"github.com/stretchr/testify/require"
)

// Simple mocks for the migration router.
type repoFinder struct {
	dbRepo *models.Repository
	err    error
}

func (f *repoFinder) FindByPath(_ context.Context, _ string) (*models.Repository, error) {
	return f.dbRepo, f.err
}

func Test_MigrationStatus(t *testing.T) {
	var tests = []struct {
		name           string
		finder         RepoFinder
		path           string
		repoOnFS       bool
		expectedStatus migration.Status
		expectErr      bool
	}{
		{
			name:           "Old Repo",
			finder:         &repoFinder{}, // Repo not on database, no error
			path:           "old/repo",
			repoOnFS:       true,
			expectedStatus: migration.StatusOldRepo,
		},
		{
			name:           "New Repo",
			finder:         &repoFinder{}, // Repo not on database, no error
			path:           "new/repo",
			repoOnFS:       false,
			expectedStatus: migration.StatusNewRepo,
		},
		{
			name:           "Finding Repository on Database Error",
			finder:         &repoFinder{nil, fmt.Errorf("repoFinder test error")},
			path:           "broken/repo",
			repoOnFS:       false,
			expectedStatus: migration.StatusError,
			expectErr:      true,
		},
		{
			name: "Native Repo",
			finder: &repoFinder{dbRepo: &models.Repository{
				Path:            "native/repo",
				MigrationStatus: migration.RepositoryStatusNative,
			}},
			path:           "native/repo",
			repoOnFS:       false,
			expectedStatus: migration.StatusOnDatabase,
		},
		{
			name: "Imported Repo",
			finder: &repoFinder{dbRepo: &models.Repository{
				Path:            "imported/repo",
				MigrationStatus: migration.RepositoryStatusImportComplete,
			}},
			path:           "imported/repo",
			repoOnFS:       true,
			expectedStatus: migration.StatusOnDatabase,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := createRegistry(t)
			repo := makeRepository(t, registry, tt.path)

			if tt.repoOnFS {
				_, err := testutil.UploadRandomSchema2Image(repo)
				require.NoError(t, err)
			}

			router := &Router{tt.finder}

			status, err := router.MigrationStatus(context.Background(), repo)
			require.True(t, tt.expectErr == (err != nil))
			require.Equal(t, tt.expectedStatus, status)
		})
	}
}

func createRegistry(t *testing.T) distribution.Namespace {
	ctx := context.Background()

	registry, err := storage.NewRegistry(ctx, inmemory.New())
	if err != nil {
		t.Fatalf("Failed to construct namespace")
	}
	return registry
}

func makeRepository(t *testing.T, registry distribution.Namespace, name string) distribution.Repository {
	ctx := context.Background()

	// Initialize a dummy repository
	named, err := reference.WithName(name)
	if err != nil {
		t.Fatalf("Failed to parse name %s:  %v", name, err)
	}

	repo, err := registry.Repository(ctx, named)
	if err != nil {
		t.Fatalf("Failed to construct repository: %v", err)
	}
	return repo
}
