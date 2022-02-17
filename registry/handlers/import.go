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
	"github.com/docker/distribution/registry/handlers/internal/metrics"
	"github.com/docker/distribution/registry/internal/migration"
	"github.com/docker/distribution/registry/storage"
	ghandlers "github.com/gorilla/handlers"
	"gitlab.com/gitlab-org/labkit/correlation"
	"gitlab.com/gitlab-org/labkit/errortracking"
)

// importHandler handles http operations on repository imports
type importHandler struct {
	*Context

	datastore.RepositoryStore
	preImport bool
	timeout   time.Duration
}

// importDispatcher takes the request context and builds the
// appropriate handler for handling import requests.
func importDispatcher(ctx *Context, r *http.Request) http.Handler {
	ih := &importHandler{
		Context:         ctx,
		RepositoryStore: datastore.NewRepositoryStore(ctx.App.db),
		timeout:         ctx.App.Config.Migration.ImportTimeout,
	}

	ihandler := ghandlers.MethodHandler{}

	if !ctx.readOnly {
		ihandler[http.MethodPut] = ih.maxConcurrentImportsMiddleware(http.HandlerFunc(ih.StartRepositoryImport))
	}

	return ihandler
}

const preImportQueryParamKey = "pre"

// StartRepository begins a repository import.
func (ih *importHandler) StartRepositoryImport(w http.ResponseWriter, r *http.Request) {
	// acquire semaphore
	ih.acquireImportSemaphore()

	defer func() {
		if len(ih.Errors) > 0 {
			// make sure we release the resource if this handler returned an error
			ih.releaseImportSemaphore()
		}
	}()

	l := log.GetLogger(log.WithContext(ih)).WithFields(log.Fields{
		"repository":             ih.Repository.Named().Name(),
		"current_import_count":   len(ih.importSemaphore),
		"max_concurrent_imports": cap(ih.importSemaphore),
		"delay_s":                ih.App.Config.Migration.TestSlowImport.Seconds(),
		"tag_concurrency":        ih.App.Config.Migration.TagConcurrency,
	})
	l.Debug("ImportRepository")

	if ih.App.Config.Migration.TestSlowImport > 0 {
		l.Warn("testing slow import, this should never happen in production")
	}

	dbRepo, err := ih.FindByPath(ih.Context, ih.Repository.Named().Name())
	if err != nil {
		ih.Errors = append(ih.Errors, errcode.FromUnknownError(err))
		return
	}

	// TODO: We should have a specific error for bad query values.
	// https://gitlab.com/gitlab-org/container-registry/-/issues/587
	ih.preImport, err = getPreValue(r)
	if err != nil {
		ih.Errors = append(ih.Errors, errcode.FromUnknownError(err))
		return
	}

	if ih.preImport {
		ih.timeout = ih.App.Config.Migration.PreImportTimeout
	}

	l = l.WithFields(log.Fields{"pre_import": ih.preImport, "timeout_s": ih.timeout.Seconds()})

	// Set up metrics reporting
	report := metrics.Import()
	if ih.preImport {
		report = metrics.PreImport()
	}

	shouldImport, err := ih.shouldImport(dbRepo)
	if err != nil {
		ih.Errors = append(ih.Errors, err)

		report(false, err)
		return
	}

	if !shouldImport {
		l.Info("repository already imported, skipping import")
		w.WriteHeader(http.StatusOK)

		report(false, nil)
		return
	}

	dbRepo, err = ih.createOrUpdateRepo(ih.Context, dbRepo)
	if err != nil {
		err = errcode.FromUnknownError(err)
		ih.Errors = append(ih.Errors, err)

		report(false, err)
		return
	}

	bts, err := storage.NewBlobTransferService(ih.App.driver, ih.App.migrationDriver)
	if err != nil {
		err = errcode.FromUnknownError(err)
		ih.Errors = append(ih.Errors, err)

		report(false, err)
		return
	}

	go func() {
		defer ih.releaseImportSemaphore()

		importer := datastore.NewImporter(
			ih.App.db,
			ih.App.registry,
			datastore.WithBlobTransferService(bts),
			datastore.WithTagConcurrency(ih.App.Config.Migration.TagConcurrency),
			// This should ALWAYS be set to zero during production.
			datastore.WithTestSlowImport(ih.App.Config.Migration.TestSlowImport),
		)

		ctx, cancel := context.WithTimeout(context.Background(), ih.timeout)
		defer cancel()

		// ensure correlation ID is forwarded to the notifier
		ctx = correlation.ContextWithCorrelation(ctx, correlation.ExtractFromContext(ih.Context))

		// Add parent logger to worker context to preserve request-specific fields.
		l := log.GetLogger(log.WithContext(ih.Context))
		ctx = log.WithLogger(ctx, l)

		err = ih.runImport(ctx, importer, dbRepo)
		if err != nil {
			l.WithError(err).Error("importing repository")
			errortracking.Capture(err, errortracking.WithContext(ctx), errortracking.WithRequest(r))

			report(true, err)
		}

		report(true, nil)
		ih.sendImportNotification(ctx, dbRepo, err)
	}()

	w.WriteHeader(http.StatusAccepted)
}

func (ih *importHandler) shouldImport(dbRepo *models.Repository) (bool, error) {
	if dbRepo != nil {
		switch status := dbRepo.MigrationStatus; {
		// Do not begin an import or pre import for a repository which has already been imported.
		case status.OnDatabase():
			return false, nil

		// Do not begin an import or pre import with a repository that already has
		//	an import or pre import operation ongoing.
		case status == migration.RepositoryStatusPreImportInProgress:
			detail := v1.ErrorCodePreImportInProgressErrorDetail(ih.Repository)
			return false, v1.ErrorCodePreImportInProgress.WithDetail(detail)

		case status == migration.RepositoryStatusImportInProgress:
			detail := v1.ErrorCodeImportInProgressErrorDetail(ih.Repository)
			return false, v1.ErrorCodeImportInProgress.WithDetail(detail)

		// Do not begin an import for a repository that failed to pre import, allow
		// additional pre import attempts.
		case status == migration.RepositoryStatusPreImportFailed && !ih.preImport:
			detail := v1.ErrorCodePreImportFailedErrorDetail(ih.Repository)
			return false, v1.ErrorCodePreImportInFailed.WithDetail(detail)
		}
	}

	validator, ok := ih.Repository.(storage.RepositoryValidator)
	if !ok {
		return false, errcode.FromUnknownError(fmt.Errorf("repository does not implement RepositoryValidator interface"))
	}

	// check if repository exists in the old storage prefix before attempting import
	exists, err := validator.Exists(ih)
	if err != nil {
		return false, errcode.FromUnknownError(fmt.Errorf("unable to determine if repository exists on old storage prefix: %w", err))
	}

	if !exists {
		return false, v2.ErrorCodeNameUnknown
	}

	return true, nil
}

func (ih *importHandler) createOrUpdateRepo(ctx context.Context, dbRepo *models.Repository) (*models.Repository, error) {
	var status migration.RepositoryStatus
	if ih.preImport {
		status = migration.RepositoryStatusPreImportInProgress
	} else {
		status = migration.RepositoryStatusImportInProgress
	}

	var err error
	if dbRepo == nil {
		dbRepo, err = ih.CreateByPath(ih.Context, ih.Repository.Named().Name(), datastore.WithMigrationStatus(status))
		if err != nil {
			return dbRepo, fmt.Errorf("creating repository for import: %w", err)
		}
	} else {
		dbRepo.MigrationStatus = status
		if err := ih.Update(ih.Context, dbRepo); err != nil {
			return dbRepo, fmt.Errorf("updating migration status before import: %w", err)
		}
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

func (ih *importHandler) sendImportNotification(ctx context.Context, dbRepo *models.Repository, err error) {
	if ih.App.importNotifier == nil {
		return
	}

	importNotification := &migration.Notification{
		Name:   dbRepo.Name,
		Path:   dbRepo.Path,
		Status: string(dbRepo.MigrationStatus),
		// TODO: replace with migration_error when https://gitlab.com/gitlab-org/container-registry/-/issues/566 is done
		Detail: getImportDetail(ih.preImport, err),
	}

	if err := ih.App.importNotifier.Notify(ctx, importNotification); err != nil {
		log.GetLogger(log.WithContext(ih)).WithError(err).Error("failed to send import notification")
		errortracking.Capture(err, errortracking.WithContext(ctx))
	}
}

func getImportDetail(preImport bool, err error) string {
	if err != nil {
		return err.Error()
	}

	if preImport {
		return "pre import completed successfully"
	}

	return "import completed successfully"
}

// maxConcurrentImportsMiddleware is a middleware that checks the configured `maxconcurrentimports`
// and does not allow requests to begin an import if the limit has been reached.
func (ih *importHandler) maxConcurrentImportsMiddleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: add saturation metrics https://gitlab.com/gitlab-org/container-registry/-/issues/586
		// the capacity is equivalent to the `maxconcurrentimports` value
		capacity := cap(ih.importSemaphore)
		// the length of the semaphore tells us how many resources are being currently used
		length := len(ih.importSemaphore)

		if length < capacity {
			handler.ServeHTTP(w, r)
			return
		}

		log.GetLogger(log.WithContext(ih.Context)).WithFields(log.Fields{
			"repository":             ih.Repository.Named().Name(),
			"max_concurrent_imports": capacity,
		}).Warn("import has been rate limited")

		detail := v1.ErrorCodeImportRateLimitedDetail(ih.Repository)
		ih.Errors = append(ih.Errors, v1.ErrorCodeImportRateLimited.WithDetail(detail))
		return
	})
}

func (ih *importHandler) acquireImportSemaphore() {
	ih.importSemaphore <- struct{}{}
}

func (ih *importHandler) releaseImportSemaphore() {
	<-ih.importSemaphore
}
