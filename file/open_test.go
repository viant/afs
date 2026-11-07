package file

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOpenURLDoesNotCreateMissingParent(t *testing.T) {
	parent := filepath.Join(t.TempDir(), "${workPath}", "offline")
	reader, err := OpenURL(context.Background(), filepath.Join(parent, "run.yaml"))
	if reader != nil {
		defer reader.Close()
	}
	require.ErrorIs(t, err, os.ErrNotExist)
	_, err = os.Stat(filepath.Dir(parent))
	require.ErrorIs(t, err, os.ErrNotExist)
}

func TestOpenURLPreservesParentPermissions(t *testing.T) {
	parent := t.TempDir()
	require.NoError(t, os.Chmod(parent, 0700))
	filename := filepath.Join(parent, "run.yaml")
	require.NoError(t, os.WriteFile(filename, []byte("existing contents"), 0600))
	before, err := os.Stat(parent)
	require.NoError(t, err)

	reader, err := OpenURL(context.Background(), filename)
	require.NoError(t, err)
	defer reader.Close()
	content, err := io.ReadAll(reader)
	require.NoError(t, err)
	require.Equal(t, "existing contents", string(content))

	after, err := os.Stat(parent)
	require.NoError(t, err)
	require.Equal(t, before.Mode().Perm(), after.Mode().Perm())
}

func TestUploadCreatesMissingParent(t *testing.T) {
	filename := filepath.Join(t.TempDir(), "generated", "offline", "run.yaml")
	require.NoError(t, Upload(context.Background(), filename, 0600, strings.NewReader("generated contents")))
	content, err := os.ReadFile(filename)
	require.NoError(t, err)
	require.Equal(t, "generated contents", string(content))
}
