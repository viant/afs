package scratchpad

import (
	"context"
	"encoding/json"
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

func TestService_ContextUserIsolatesSharedService(t *testing.T) {
	ctx := context.Background()
	svc := New(WithRootURI("mem://localhost/scratchpad_context/${userID}"))

	aliceCtx := ContextWithUserID(ctx, "alice@example.com")
	bobCtx := ContextWithUserID(ctx, "bob@example.com")

	_, err := svc.Memorize(aliceCtx, &MemorizeInput{Key: "notes/report", Description: "Report", Body: "alice"})
	require.NoError(t, err)
	_, err = svc.Memorize(bobCtx, &MemorizeInput{Key: "notes/report", Description: "Report", Body: "bob"})
	require.NoError(t, err)

	alice, err := svc.Fetch(aliceCtx, "notes/report")
	require.NoError(t, err)
	require.Equal(t, "alice", alice.Body)

	bob, err := svc.Fetch(bobCtx, "notes/report")
	require.NoError(t, err)
	require.Equal(t, "bob", bob.Body)
}

func TestService_UserIDResolutionPrecedence(t *testing.T) {
	ctx := context.Background()
	svc := New(
		WithRootURI("mem://localhost/scratchpad_precedence/${userID}"),
		WithUserID("static"),
		WithUserIDProvider(func(context.Context) string { return "provider" }),
	)

	_, userID, err := svc.ResolveRootURIContext(ContextWithUserID(ctx, "context"))
	require.NoError(t, err)
	require.Equal(t, "context", userID)

	_, userID, err = svc.ResolveRootURIContext(ctx)
	require.NoError(t, err)
	require.Equal(t, "provider", userID)

	staticSvc := New(WithRootURI("mem://localhost/scratchpad_precedence_static/${userID}"), WithUserID("static"))
	_, userID, err = staticSvc.ResolveRootURIContext(ctx)
	require.NoError(t, err)
	require.Equal(t, "static", userID)
}

func TestProvider_OpenArtifactWithContextUser(t *testing.T) {
	ctx := context.Background()
	fs := afs.New()
	root := "file://" + t.TempDir() + "/scratchpad/${userID}"
	sourceURL := "mem://localhost/scratchpad_provider_context/report.txt"
	userCtx := ContextWithUserID(ctx, "devuser")
	require.NoError(t, fs.Upload(ctx, sourceURL, 0o644, strings.NewReader("report content")))

	svc := New(
		WithAFS(fs),
		WithRootURI(root),
		WithAllowedTargetSchemes("mem"),
	)
	_, err := svc.Memorize(userCtx, &MemorizeInput{
		Key:         ArtifactKey("context-a1"),
		Description: "Artifact context-a1",
		Body:        `{"kind":"artifact","artifactId":"context-a1","name":"report.txt","contentType":"text/plain","sourceURL":"` + sourceURL + `"}`,
	})
	require.NoError(t, err)

	Register(
		WithAFS(fs),
		WithRootURI(root),
		WithAllowedTargetSchemes("mem"),
	)
	reader, err := fs.OpenURL(userCtx, "scratchpad://artifact/context-a1")
	require.NoError(t, err)
	defer reader.Close()
	data, err := io.ReadAll(reader)
	require.NoError(t, err)
	require.Equal(t, "report content", string(data))
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

func TestService_RejectsMismatchedNoteUser(t *testing.T) {
	ctx := context.Background()
	fs := afs.New()
	rootTemplate := "mem://localhost/scratchpad_owner/${userID}"
	root, _, err := ResolveRootURI(rootTemplate, "devuser", "", nil)
	require.NoError(t, err)
	require.NoError(t, fs.Create(ctx, root, 0o755, true))

	svc := New(WithAFS(fs), WithRootURI(rootTemplate), WithUserID("devuser"), WithAllowedTargetSchemes("mem"))
	writeNote(t, ctx, fs, root, Note{
		Key:         "notes/bad",
		Description: "Bad",
		Body:        "body",
		UserID:      "other",
		UpdatedAt:   time.Now().UTC().Format(time.RFC3339Nano),
	})
	writeNote(t, ctx, fs, root, Note{
		Key:         ArtifactKey("bad-artifact"),
		Description: "Bad artifact",
		Body:        `{"artifactId":"bad-artifact","sourceURL":"mem://localhost/private"}`,
		UserID:      "other",
		UpdatedAt:   time.Now().UTC().Format(time.RFC3339Nano),
	})

	_, err = svc.Fetch(ctx, "notes/bad")
	require.Error(t, err)
	require.Contains(t, err.Error(), "different user")

	_, err = svc.Append(ctx, &AppendInput{Key: "notes/bad", Body: "append"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "different user")

	_, err = svc.List(ctx)
	require.Error(t, err)
	require.Contains(t, err.Error(), "different user")

	err = svc.Delete(ctx, "notes/bad")
	require.Error(t, err)
	require.Contains(t, err.Error(), "different user")

	_, _, err = svc.OpenArtifact(ctx, "bad-artifact")
	require.Error(t, err)
	require.Contains(t, err.Error(), "different user")
}

func TestService_RejectsMissingNoteUserID(t *testing.T) {
	ctx := context.Background()
	fs := afs.New()
	rootTemplate := "mem://localhost/scratchpad_missing_user/${userID}"
	root, _, err := ResolveRootURI(rootTemplate, "devuser", "", nil)
	require.NoError(t, err)
	require.NoError(t, fs.Create(ctx, root, 0o755, true))

	svc := New(WithAFS(fs), WithRootURI(rootTemplate), WithUserID("devuser"))
	writeNote(t, ctx, fs, root, Note{
		Key:         "notes/missing",
		Description: "Missing user",
		Body:        "body",
		UpdatedAt:   time.Now().UTC().Format(time.RFC3339Nano),
	})

	_, err = svc.Fetch(ctx, "notes/missing")
	require.Error(t, err)
	require.Contains(t, err.Error(), "missing userId")
}

func writeNote(t *testing.T, ctx context.Context, fs afs.Service, root string, note Note) {
	t.Helper()
	data, err := json.Marshal(note)
	require.NoError(t, err)
	require.NoError(t, fs.Upload(ctx, NoteURL(root, note.Key), 0o644, strings.NewReader(string(data))))
}
