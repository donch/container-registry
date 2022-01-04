// Package v1 provides API routes for GitLab specific features, which exposes
// additional functionality beyond that of the distribution v2 API. These routes
// depend on functionality provided by the metadata database.
package v1

import (
	"fmt"
	"regexp"

	"github.com/gorilla/mux"
)

// Route is the name and path pair of a GitLab v1 API route.
type Route struct{ Name, Path string }

// Base is the API route under which all other GitLab v1 API routes are found.
var Base = Route{
	Name: "base",
	Path: "/gitlab/v1/",
}

// RouteRegex provides a regexp which can be used to determine if a string
// is a GitLab v1 API route.
var RouteRegex = regexp.MustCompile(fmt.Sprintf("^%s.*", Base.Path))

// Router returns a new *mux.Router for the Gitlab v1 API.
func Router() *mux.Router {
	router := mux.NewRouter()
	router.StrictSlash(true)

	router.Path(Base.Path).Name(Base.Name)

	return router
}
