package handlers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/docker/distribution/log"
	"github.com/docker/distribution/registry/api/errcode"
	v2 "github.com/docker/distribution/registry/api/v2"
	"github.com/docker/distribution/registry/datastore"
	"github.com/docker/distribution/registry/storage"
	ghandlers "github.com/gorilla/handlers"
	"gitlab.com/gitlab-org/labkit/errortracking"
)

// importHandler handles http operations on repository imports
type importHandler struct {
	*Context

	datastore.RepositoryReader
}

// importDispatcher takes the request context and builds the
// appropriate handler for handling import requests.
func importDispatcher(ctx *Context, r *http.Request) http.Handler {
	ih := &importHandler{
		Context:          ctx,
		RepositoryReader: datastore.NewRepositoryStore(ctx.App.db),
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
	// * it's actually Migrated
	// * a pre-import is currently in progress
	// * a pre-import failed
	// * another import is currently in progress
	// and communicate appropriately back to the client, as defined in the spec.
	if dbRepo != nil {
		// Horrible hack before !510 is finished to detect pre imported repositories.
		tags, err := ih.Tags(ih.Context, dbRepo)
		if err != nil {
			ih.Errors = append(ih.Errors, errcode.FromUnknownError(err))
			return
		}

		if len(tags) != 0 {
			l.Info("repository already imported, skipping import")
			w.WriteHeader(http.StatusOK)
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

	preImport, err := getPreValue(r)
	if err != nil {
		ih.Errors = append(ih.Errors, errcode.FromUnknownError(err))
	}

	l = l.WithFields(log.Fields{"pre_import": preImport})

	go func() {
		bts, err := storage.NewBlobTransferService(ih.App.driver, ih.App.migrationDriver)
		if err != nil {
			// TODO: We should have a specific error for bad query values.
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

		if preImport {
			err = importer.PreImport(ctx, ih.Repository.Named().Name())
		} else {
			err = importer.Import(ctx, ih.Repository.Named().Name())
		}

		if err != nil {
			l.WithError(err).Error("importing repository")
			errortracking.Capture(err, errortracking.WithContext(ctx), errortracking.WithRequest(r))
		}
	}()

	w.WriteHeader(http.StatusAccepted)
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
