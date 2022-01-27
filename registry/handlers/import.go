package handlers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/docker/distribution/log"
	"github.com/docker/distribution/registry/api/errcode"
	v1 "github.com/docker/distribution/registry/api/gitlab/v1"
	v2 "github.com/docker/distribution/registry/api/v2"
	"github.com/docker/distribution/registry/datastore"
	"github.com/docker/distribution/registry/datastore/models"
	"github.com/docker/distribution/registry/internal/migration"
	"github.com/docker/distribution/registry/storage"
	ghandlers "github.com/gorilla/handlers"
	"gitlab.com/gitlab-org/labkit/errortracking"
)

// importHandler handles http operations on repository imports
type importHandler struct {
	*Context

	datastore.RepositoryStore
	preImport bool
}

// importDispatcher takes the request context and builds the
// appropriate handler for handling import requests.
func importDispatcher(ctx *Context, r *http.Request) http.Handler {
	ih := &importHandler{
		Context:         ctx,
		RepositoryStore: datastore.NewRepositoryStore(ctx.App.db),
	}

	ihandler := ghandlers.MethodHandler{}

	if !ctx.readOnly {
		ihandler[http.MethodPut] = http.HandlerFunc(ih.StartRepositoryImport)
	}

	return ihandler
}

const preImportQueryParamKey = "pre"

// StartRepository begins a repository import.
func (ih *importHandler) StartRepositoryImport(w http.ResponseWriter, r *http.Request) {
	l := log.GetLogger(log.WithContext(ih)).WithFields(log.Fields{"repository": ih.Repository.Named().Name()})
	l.Debug("ImportRepository")

	dbRepo, err := ih.FindByPath(ih.Context, ih.Repository.Named().Name())
	if err != nil {
		ih.Errors = append(ih.Errors, errcode.FromUnknownError(err))
		return
	}

	// Do not begin an import for a repository which has already been imported.
	// TODO: When https://gitlab.com/gitlab-org/container-registry/-/issues/510 is
	// complete, we should check the repository import status to determine if:
	// * it's actually Migrated - DONE
	// * a pre-import is currently in progress
	// * a pre-import failed
	// * another import is currently in progress
	// and communicate appropriately back to the client, as defined in the spec.
	if dbRepo != nil {
		if dbRepo.MigrationStatus.OnDatabase() {
			l.Info("repository already imported, skipping import")
			w.WriteHeader(http.StatusOK)
			return
		}

		if dbRepo.MigrationStatus == migration.RepositoryStatusPreImportInProgress {
			detail := v1.ErrorCodePreImportInProgressErrorDetail(ih.Repository)
			ih.Errors = append(ih.Errors, v1.ErrorCodePreImportInProgress.WithDetail(detail))
			return
		}
	}

	validator, ok := ih.Repository.(storage.RepositoryValidator)
	if !ok {
		ih.Errors = append(ih.Errors, errcode.FromUnknownError(fmt.Errorf("repository does not implement RepositoryValidator interface")))
		return
	}

	// check if repository exists in the old storage prefix before attempting import
	exists, err := validator.Exists(ih)
	if err != nil {
		ih.Errors = append(ih.Errors, errcode.FromUnknownError(fmt.Errorf("unable to determine if repository exists on old storage prefix: %w", err)))
		return
	}

	if !exists {
		ih.Errors = append(ih.Errors, v2.ErrorCodeNameUnknown)
		return
	}

	// TODO: We should have a specific error for bad query values.
	ih.preImport, err = getPreValue(r)
	if err != nil {
		ih.Errors = append(ih.Errors, errcode.FromUnknownError(err))
		return
	}

	l = l.WithFields(log.Fields{"pre_import": ih.preImport})

	dbRepo, err = ih.createOrUpdateRepo(ih.Context, dbRepo)
	if err != nil {
		ih.Errors = append(ih.Errors, errcode.FromUnknownError(err))
		return
	}

	go func() {
		bts, err := storage.NewBlobTransferService(ih.App.driver, ih.App.migrationDriver)
		if err != nil {
			ih.Errors = append(ih.Errors, errcode.FromUnknownError(err))
		}

		importer := datastore.NewImporter(
			ih.App.db,
			ih.App.registry,
			datastore.WithBlobTransferService(bts),
			datastore.WithTagConcurrency(ih.App.Config.Migration.TagConcurrency),
		)

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Add parent logger to worker context to preserve request-specific fields.
		l := log.GetLogger(log.WithContext(ih.Context))
		ctx = log.WithLogger(ctx, l)

		ih.runImport(ctx, importer, dbRepo)
		if err != nil {
			l.WithError(err).Error("importing repository")
			errortracking.Capture(err, errortracking.WithContext(ctx), errortracking.WithRequest(r))
		}
	}()

	w.WriteHeader(http.StatusAccepted)
}

func (ih *importHandler) createOrUpdateRepo(ctx context.Context, dbRepo *models.Repository) (*models.Repository, error) {
	// TODO: We should set the migration status in one step if the repository does not exist.
	var err error
	if dbRepo == nil {
		dbRepo, err = ih.CreateByPath(ih.Context, ih.Repository.Named().Name())
		if err != nil {
			return dbRepo, fmt.Errorf("creating repository for import: %w", err)
		}
	}

	if ih.preImport {
		dbRepo.MigrationStatus = migration.RepositoryStatusPreImportInProgress
	} else {
		dbRepo.MigrationStatus = migration.RepositoryStatusImportInProgress
	}

	if err := ih.Update(ih.Context, dbRepo); err != nil {
		return dbRepo, fmt.Errorf("updating migration status before import: %w", err)
	}

	return dbRepo, nil
}

func (ih *importHandler) runImport(ctx context.Context, importer *datastore.Importer, dbRepo *models.Repository) error {
	if ih.preImport {
		if err := importer.PreImport(ctx, dbRepo.Path); err != nil {
			dbRepo.MigrationStatus = migration.RepositoryStatusPreImportFailed
			if err := ih.Update(ctx, dbRepo); err != nil {
				return fmt.Errorf("updating migration status after failed pre import: %w", err)
			}

			return err
		}

		dbRepo.MigrationStatus = migration.RepositoryStatusPreImportComplete
		if err := ih.Update(ctx, dbRepo); err != nil {
			return fmt.Errorf("updating migration status after successful pre import: %w", err)
		}

		return nil
	}

	if err := importer.Import(ctx, dbRepo.Path); err != nil {
		dbRepo.MigrationStatus = migration.RepositoryStatusImportFailed
		if err := ih.Update(ctx, dbRepo); err != nil {
			return fmt.Errorf("updating migration status after failed import: %w", err)
		}

		return err
	}

	dbRepo.MigrationStatus = migration.RepositoryStatusImportComplete
	if err := ih.Update(ctx, dbRepo); err != nil {
		return fmt.Errorf("updating migration status after successful import: %w", err)
	}

	return nil
}

// The API spec for this route only specifies 'true' or 'false', while
// strconv.ParseBool accepts a greater range of string values.
func getPreValue(r *http.Request) (bool, error) {
	preImportValue := r.URL.Query().Get(preImportQueryParamKey)

	switch preImportValue {
	case "true":
		return true, nil
	case "false", "":
		return false, nil
	default:
		return false, fmt.Errorf("pre value must be 'true' or 'false', got %s", preImportValue)
	}
}
