package migration

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_CodePathVal_String(t *testing.T) {
	require.Equal(t, "", UnknownCodePath.String())
	require.Equal(t, "old", OldCodePath.String())
	require.Equal(t, "new", NewCodePath.String())
}

func Test_migrationContext_WithCodePath(t *testing.T) {
	ctx := context.WithValue(context.Background(), "foo", "bar")
	mc := WithCodePath(ctx, OldCodePath)

	v, ok := mc.Value(CodePathKey).(CodePathVal)
	require.True(t, ok)
	require.Equal(t, OldCodePath, v)

	s, ok := mc.Value("foo").(string)
	require.True(t, ok)
	require.Equal(t, "bar", s)
}

func Test_migrationContext_CodePath(t *testing.T) {
	require.Equal(t, UnknownCodePath, CodePath(context.Background()))

	mc := WithCodePath(context.Background(), OldCodePath)
	require.Equal(t, OldCodePath, CodePath(mc))

	mc = WithCodePath(context.Background(), NewCodePath)
	require.Equal(t, NewCodePath, CodePath(mc))
}
