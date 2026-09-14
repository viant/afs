package scratchpad

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/viant/afs"
	"github.com/viant/afs/storage"
)

// The service-scoped factory seam must retain scratchpad's note persistence,
// contextual roots and artifact fetch policy; it is not a plain-file provider.
func TestScopedProviderNotesAndArtifacts(t *testing.T) {
	ctx := context.Background()
	backend := afs.New()
	cfg := Config{FS: backend, RootURITemplate: "file://" + t.TempDir() + "/${userID}", AllowedTargetSchemes: []string{"mem"}}
	svc := afs.NewWithProviders(map[string]afs.Provider{Scheme: func(...storage.Option) (storage.Manager, error) { return Provider(cfg) }})
	notes := New(WithConfig(cfg))
	for _, user := range []string{"alice", "bob"} {
		userCtx := ContextWithUserID(ctx, user)
		require.NoError(t, svc.Upload(userCtx, "scratchpad://note/report", 0600, strings.NewReader(user)))
		note, err := notes.Fetch(userCtx, "report")
		require.NoError(t, err)
		require.Equal(t, user, note.Body)
	}
	for _, user := range []string{"alice", "bob"} {
		data, err := svc.DownloadWithURL(ContextWithUserID(ctx, user), "scratchpad://note/report")
		require.NoError(t, err)
		require.Equal(t, user, string(data))
	}
	userCtx := ContextWithUserID(ctx, "alice")
	source := "mem://localhost/scoped_scratchpad/report.txt"
	require.NoError(t, backend.Upload(ctx, source, 0600, strings.NewReader("artifact bytes")))
	require.NoError(t, svc.Upload(userCtx, "scratchpad://artifact/report", 0600, strings.NewReader(`{"artifactId":"report","sourceURL":"`+source+`"}`)))
	data, err := svc.DownloadWithURL(userCtx, "scratchpad://artifact/report")
	require.NoError(t, err)
	require.Equal(t, "artifact bytes", string(data))
	require.NoError(t, svc.Upload(userCtx, "scratchpad://artifact/blocked", 0600, strings.NewReader(`{"artifactId":"blocked","sourceURL":"https://example.invalid/secret"}`)))
	_, err = svc.OpenURL(userCtx, "scratchpad://artifact/blocked")
	require.Error(t, err)
	require.Contains(t, err.Error(), "not allowed")
	_, err = svc.OpenURL(ContextWithUserID(ctx, "bob"), "scratchpad://artifact/report")
	require.Error(t, err)
}
