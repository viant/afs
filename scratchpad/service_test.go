package scratchpad

import (
	"context"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/viant/afs"
)

func TestService_MemorizeFetchListAppend(t *testing.T) {
	ctx := context.Background()
	svc := New(
		WithRootURI("mem://localhost/scratchpad_service/${userID}"),
		WithUserID("team/user@example.com"),
		WithNow(func() time.Time { return time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC) }),
	)

	memorized, err := svc.Memorize(ctx, &MemorizeInput{Key: "notes/report", Description: "Report", Body: "first"})
	require.NoError(t, err)
	require.Equal(t, "notes/report", memorized.Key)
	require.Equal(t, "2026-07-06T12:00:00Z", memorized.UpdatedAt)

	appended, err := svc.Append(ctx, &AppendInput{Key: "notes/report", Body: "second"})
	require.NoError(t, err)
	require.False(t, appended.Created)

	fetched, err := svc.Fetch(ctx, "notes/report")
	require.NoError(t, err)
	require.Equal(t, "first\n\n---\n\nsecond", fetched.Body)

	listed, err := svc.List(ctx)
	require.NoError(t, err)
	require.Len(t, listed.Entries, 1)
	require.Equal(t, "notes/report", listed.Entries[0].Key)
}

func TestService_SharedFileBackedRootAcrossClients(t *testing.T) {
	ctx := context.Background()
	root := "file://" + t.TempDir() + "/scratchpad/${userID}"

	writer := New(WithRootURI(root), WithUserID("devuser"))
	_, err := writer.Memorize(ctx, &MemorizeInput{Key: "artifact/a1", Description: "Artifact", Body: `{"sourceURL":"file:///tmp/a1.txt"}`})
	require.NoError(t, err)

	reader := New(WithRootURI(root), WithUserID("devuser"))
	fetched, err := reader.Fetch(ctx, "artifact/a1")
	require.NoError(t, err)
	require.Contains(t, fetched.Body, "file:///tmp/a1.txt")
}

func TestProvider_OpenArtifact(t *testing.T) {
	ctx := context.Background()
	fs := afs.New()
	root := "file://" + t.TempDir() + "/scratchpad/${userID}"
	sourceURL := "mem://localhost/scratchpad_provider/report.txt"
	require.NoError(t, fs.Upload(ctx, sourceURL, 0o644, strings.NewReader("report content")))

	svc := New(
		WithAFS(fs),
		WithRootURI(root),
		WithUserID("devuser"),
		WithAllowedTargetSchemes("mem"),
	)
	_, err := svc.Memorize(ctx, &MemorizeInput{
		Key:         ArtifactKey("a1"),
		Description: "Artifact a1",
		Body:        `{"kind":"artifact","artifactId":"a1","name":"report.txt","contentType":"text/plain","sourceURL":"` + sourceURL + `"}`,
	})
	require.NoError(t, err)

	Register(
		WithAFS(fs),
		WithRootURI(root),
		WithUserID("devuser"),
		WithAllowedTargetSchemes("mem"),
	)
	reader, err := fs.OpenURL(ctx, "scratchpad://artifact/a1")
	require.NoError(t, err)
	defer reader.Close()
	data, err := io.ReadAll(reader)
	require.NoError(t, err)
	require.Equal(t, "report content", string(data))
}

func TestProvider_RejectsDisallowedArtifactTargetScheme(t *testing.T) {
	ctx := context.Background()
	fs := afs.New()
	root := "file://" + t.TempDir() + "/scratchpad/${userID}"
	svc := New(WithAFS(fs), WithRootURI(root), WithUserID("devuser"), WithAllowedTargetSchemes("file"))
	_, err := svc.Memorize(ctx, &MemorizeInput{
		Key:         ArtifactKey("a1"),
		Description: "Artifact a1",
		Body:        `{"artifactId":"a1","sourceURL":"mem://localhost/private"}`,
	})
	require.NoError(t, err)

	Register(WithAFS(fs), WithRootURI(root), WithUserID("devuser"), WithAllowedTargetSchemes("file"))
	_, err = fs.OpenURL(ctx, "scratchpad://artifact/a1")
	require.Error(t, err)
	require.Contains(t, err.Error(), `source scheme "mem" is not allowed`)
}
