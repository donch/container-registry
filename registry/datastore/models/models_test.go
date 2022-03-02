package models

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRepository_IsTopLevel(t *testing.T) {
	r := &Repository{Path: "foo"}
	require.True(t, r.IsTopLevel())
	r.Path = "foo/bar"
	require.False(t, r.IsTopLevel())
}
