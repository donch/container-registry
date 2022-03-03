package router

import (
	"context"
	"errors"
	"fmt"

	"github.com/docker/distribution"
	"github.com/docker/distribution/registry/datastore/models"
	"github.com/docker/distribution/registry/internal/migration"
	"github.com/docker/distribution/registry/storage"
)

// RepoFinder finds a repository on the database by path.
type RepoFinder interface {
	FindByPath(ctx context.Context, path string) (*models.Repository, error)
}

// Router determines the appropriate migration status of a repository based off of
// the repository's migration status, and presence on the database, and old prefix.
// See https://gitlab.com/gitlab-org/container-registry/-/issues/374#routing-1
// for an in-depth explanation.
type Router struct {
	RepoFinder
}

// MigrationStatus returns the migration status of a repository with the given path.
func (r *Router) MigrationStatus(ctx context.Context, repo distribution.Repository) (migration.Status, error) {
	dbRepo, err := r.FindByPath(ctx, repo.Named().Name())
	if err != nil {
		return migration.StatusError, fmt.Errorf("finding repository in database: %w", err)
	}

	if dbRepo != nil {
		if dbRepo.MigrationStatus.OnDatabase() {
			return migration.StatusOnDatabase, nil
		}

		if dbRepo.MigrationStatus == migration.RepositoryStatusImportInProgress {
			return migration.StatusImportInProgress, nil
		}
	}

	// We didn't find the repository in the database, but we need to check the
	// old prefix on the filesystem to see if this a new repository or not.
	validator, ok := repo.(storage.RepositoryValidator)
	if !ok {
		return migration.StatusError, errors.New("repository does not implement RepositoryValidator interface")
	}

	exists, err := validator.Exists(ctx)
	if err != nil {
		return migration.StatusError, fmt.Errorf("checking for existence of repository on filesystem: %w", err)
	}

	if exists {
		return migration.StatusOldRepo, nil
	}

	return migration.StatusNewRepo, nil
}
