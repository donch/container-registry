package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/docker/distribution/log"
	"github.com/docker/distribution/registry/api/errcode"
	v1 "github.com/docker/distribution/registry/api/gitlab/v1"
	v2 "github.com/docker/distribution/registry/api/v2"
	"github.com/docker/distribution/registry/datastore"
	"github.com/docker/distribution/registry/datastore/models"
	"github.com/gorilla/handlers"
)

type repositoryHandler struct {
	*Context
}

func repositoryDispatcher(ctx *Context, _ *http.Request) http.Handler {
	repositoryHandler := &repositoryHandler{
		Context: ctx,
	}

	return handlers.MethodHandler{
		http.MethodGet: http.HandlerFunc(repositoryHandler.GetRepository),
	}
}

type RepositoryAPIResponse struct {
	Name      string `json:"name"`
	Path      string `json:"path"`
	Size      *int64 `json:"size_bytes,omitempty"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at,omitempty"`
}

const (
	sizeQueryParamKey                      = "size"
	sizeQueryParamSelfValue                = "self"
	sizeQueryParamSelfWithDescendantsValue = "self_with_descendants"
)

var sizeQueryParamValidValues = []string{
	sizeQueryParamSelfValue,
	sizeQueryParamSelfWithDescendantsValue,
}

func isQueryParamValueValid(value string, validValues []string) bool {
	for _, v := range validValues {
		if value == v {
			return true
		}
	}
	return false
}

func sizeQueryParamValue(r *http.Request) string {
	return r.URL.Query().Get(sizeQueryParamKey)
}

// timeToString converts a time.Time to a ISO 8601 with millisecond precision string. This is the standard format used
// across GitLab applications.
func timeToString(t time.Time) string {
	return t.UTC().Format("2006-01-02T15:04:05.000Z07:00")
}

func (h *repositoryHandler) GetRepository(w http.ResponseWriter, r *http.Request) {
	l := log.GetLogger(log.WithContext(h)).WithFields(log.Fields{"path": h.Repository.Named().Name()})
	l.Debug("GetRepository")

	var withSize bool
	sizeVal := sizeQueryParamValue(r)
	if sizeVal != "" {
		if !isQueryParamValueValid(sizeVal, sizeQueryParamValidValues) {
			detail := v1.InvalidQueryParamValueErrorDetail(sizeQueryParamKey, sizeQueryParamValidValues)
			h.Errors = append(h.Errors, v1.ErrorCodeInvalidQueryParamValue.WithDetail(detail))
			return
		}
		withSize = true
	}

	store := datastore.NewRepositoryStore(h.db)
	repo, err := store.FindByPath(h.Context, h.Repository.Named().Name())
	if err != nil {
		h.Errors = append(h.Errors, errcode.FromUnknownError(err))
		return
	}
	if repo == nil {
		// If the caller is requesting the aggregated size of repository `foo/bar` including descendants, it might be
		// the case that `foo/bar` does not exist but there is an e.g. `foo/bar/car`. In such case we should not raise a
		// 404 if the base repository (`foo/bar`) does not exist. This is required to allow retrieving the Project level
		// usage when there is no "root" repository for such project but there is at least one sub-repository.
		if !withSize || sizeVal != sizeQueryParamSelfWithDescendantsValue {
			h.Errors = append(h.Errors, v2.ErrorCodeNameUnknown)
			return
		}
		// If this is the case, we need to find the corresponding top-level namespace. That must exist. If not we
		// throw a 404 Not Found here.
		repo = &models.Repository{Path: h.Repository.Named().Name()}
		ns := datastore.NewNamespaceStore(h.db)
		n, err := ns.FindByName(h.Context, repo.TopLevelPathSegment())
		if err != nil {
			h.Errors = append(h.Errors, errcode.FromUnknownError(err))
			return
		}
		if n == nil {
			h.Errors = append(h.Errors, v2.ErrorCodeNameUnknown)
			return
		}
		// path and namespace ID are the two required parameters for the queries in repositoryStore.SizeWithDescendants,
		// so we must fill those. We also fill the name for consistency on the response.
		repo.NamespaceID = n.ID
		repo.Name = repo.Path[strings.LastIndex(repo.Path, "/")+1:]
	}

	resp := RepositoryAPIResponse{
		Name: repo.Name,
		Path: repo.Path,
	}
	if !repo.CreatedAt.IsZero() {
		resp.CreatedAt = timeToString(repo.CreatedAt)
	}
	if repo.UpdatedAt.Valid {
		resp.UpdatedAt = timeToString(repo.UpdatedAt.Time)
	}

	var size int64
	if withSize {
		switch sizeVal {
		case sizeQueryParamSelfValue:
			size, err = store.Size(h.Context.Context, repo)
		case sizeQueryParamSelfWithDescendantsValue:
			size, err = store.SizeWithDescendants(h.Context.Context, repo)
		}
		if err != nil {
			h.Errors = append(h.Errors, errcode.FromUnknownError(err))
			return
		}
		resp.Size = &size
	}

	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)

	if err := enc.Encode(resp); err != nil {
		h.Errors = append(h.Errors, errcode.FromUnknownError(err))
		return
	}
}
