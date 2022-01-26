package v1

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/docker/distribution"
	"github.com/docker/distribution/registry/api/errcode"
)

const errGroup = "gitlab.api.v1"

// ErrorCodeInvalidQueryParamValue is returned when the value of a query parameter is invalid.
var ErrorCodeInvalidQueryParamValue = errcode.Register(errGroup, errcode.ErrorDescriptor{
	Value:          "INVALID_QUERY_PARAMETER_VALUE",
	Message:        "invalid query parameter value",
	Description:    "The value of a request query parameter is invalid",
	HTTPStatusCode: http.StatusBadRequest,
})

func InvalidQueryParamValueErrorDetail(key string, validValues []string) string {
	return fmt.Sprintf("the '%s' query parameter must be set to one of: %s", key, strings.Join(validValues, ", "))
}

// ErrorCodePreImportInProgress is returned when a repository is already pre importing.
var ErrorCodePreImportInProgress = errcode.Register(errGroup, errcode.ErrorDescriptor{
	Value:          "PRE_IMPORT_IN_PROGRESS",
	Message:        "request cannot happen concurrently with pre import",
	Description:    "The repository is being pre imported",
	HTTPStatusCode: http.StatusTooEarly,
})

func ErrorCodePreImportInProgressErrorDetail(repo distribution.Repository) string {
	return fmt.Sprintf("repository path %s", repo.Named().Name())
}
