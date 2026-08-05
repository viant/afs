package scratchpad

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/viant/afs"
	"github.com/viant/afs/storage"
)

type faultAFS struct {
	afs.Service
	existsFn   func(string) (bool, error)
	listFn     func(string) ([]storage.Object, error)
	createFn   func(string) error
	uploadFn   func(string) error
	downloadFn func(string) ([]byte, error)
	deleteFn   func(string) error
	openURLFn  func(string) (io.ReadCloser, error)
}

func (f *faultAFS) Exists(ctx context.Context, URL string, options ...storage.Option) (bool, error) {
	if f.existsFn != nil {
		return f.existsFn(URL)
	}
	return f.Service.Exists(ctx, URL, options...)
}

func (f *faultAFS) List(ctx context.Context, URL string, options ...storage.Option) ([]storage.Object, error) {
	if f.listFn != nil {
		return f.listFn(URL)
	}
	return f.Service.List(ctx, URL, options...)
}

func (f *faultAFS) Create(ctx context.Context, URL string, mode os.FileMode, isDir bool, options ...storage.Option) error {
	if f.createFn != nil {
		return f.createFn(URL)
	}
	return f.Service.Create(ctx, URL, mode, isDir, options...)
}

func (f *faultAFS) Upload(ctx context.Context, URL string, mode os.FileMode, reader io.Reader, options ...storage.Option) error {
	if f.uploadFn != nil {
		return f.uploadFn(URL)
	}
	return f.Service.Upload(ctx, URL, mode, reader, options...)
}

func (f *faultAFS) DownloadWithURL(ctx context.Context, URL string, options ...storage.Option) ([]byte, error) {
	if f.downloadFn != nil {
		return f.downloadFn(URL)
	}
	return f.Service.DownloadWithURL(ctx, URL, options...)
}

func (f *faultAFS) Delete(ctx context.Context, URL string, options ...storage.Option) error {
	if f.deleteFn != nil {
		return f.deleteFn(URL)
	}
	return f.Service.Delete(ctx, URL, options...)
}

func (f *faultAFS) OpenURL(ctx context.Context, URL string, options ...storage.Option) (io.ReadCloser, error) {
	if f.openURLFn != nil {
		return f.openURLFn(URL)
	}
	return f.Service.OpenURL(ctx, URL, options...)
}

type scratchpadStorageError struct {
	operation string
}

func (e *scratchpadStorageError) Error() string {
	return "storage " + e.operation + " failed"
}

func TestServiceStorageErrorsPreserveCauses(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		name        string
		wantMessage string
		invoke      func(*faultAFS, *Service, string, error) error
	}{
		{
			name:        "memorize root exists",
			wantMessage: "scratchpad storage unavailable",
			invoke: func(fs *faultAFS, service *Service, root string, cause error) error {
				fs.existsFn = func(string) (bool, error) { return false, cause }
				_, err := service.Memorize(ctx, &MemorizeInput{Key: "note/key", Description: "Note", Body: "body"})
				return err
			},
		},
		{
			name:        "memorize root create",
			wantMessage: "scratchpad storage unavailable",
			invoke: func(fs *faultAFS, service *Service, root string, cause error) error {
				fs.existsFn = func(string) (bool, error) { return false, nil }
				fs.createFn = func(string) error { return cause }
				_, err := service.Memorize(ctx, &MemorizeInput{Key: "note/key", Description: "Note", Body: "body"})
				return err
			},
		},
		{
			name:        "append root exists",
			wantMessage: "scratchpad storage unavailable",
			invoke: func(fs *faultAFS, service *Service, root string, cause error) error {
				fs.existsFn = func(string) (bool, error) { return false, cause }
				_, err := service.Append(ctx, &AppendInput{Key: "note/key", Body: "body"})
				return err
			},
		},
		{
			name:        "append note exists",
			wantMessage: "read scratchpad note failed",
			invoke: func(fs *faultAFS, service *Service, root string, cause error) error {
				fs.existsFn = func(URL string) (bool, error) {
					if URL == root {
						return true, nil
					}
					return false, cause
				}
				_, err := service.Append(ctx, &AppendInput{Key: "note/key", Body: "body"})
				return err
			},
		},
		{
			name:        "list root exists",
			wantMessage: "list scratchpad notes failed",
			invoke: func(fs *faultAFS, service *Service, root string, cause error) error {
				fs.existsFn = func(string) (bool, error) { return false, cause }
				_, err := service.List(ctx)
				return err
			},
		},
		{
			name:        "list entries",
			wantMessage: "list scratchpad notes failed",
			invoke: func(fs *faultAFS, service *Service, root string, cause error) error {
				fs.existsFn = func(string) (bool, error) { return true, nil }
				fs.listFn = func(string) ([]storage.Object, error) { return nil, cause }
				_, err := service.List(ctx)
				return err
			},
		},
		{
			name:        "delete note exists",
			wantMessage: "read scratchpad note failed",
			invoke: func(fs *faultAFS, service *Service, root string, cause error) error {
				fs.existsFn = func(string) (bool, error) { return false, cause }
				return service.Delete(ctx, "note/key")
			},
		},
		{
			name:        "fetch note exists",
			wantMessage: "read scratchpad note failed",
			invoke: func(fs *faultAFS, service *Service, root string, cause error) error {
				fs.existsFn = func(string) (bool, error) { return false, cause }
				_, err := service.Fetch(ctx, "note/key")
				return err
			},
		},
		{
			name:        "read note",
			wantMessage: "read scratchpad note failed",
			invoke: func(fs *faultAFS, service *Service, root string, cause error) error {
				fs.existsFn = func(string) (bool, error) { return true, nil }
				fs.downloadFn = func(string) ([]byte, error) { return nil, cause }
				_, err := service.Fetch(ctx, "note/key")
				return err
			},
		},
		{
			name:        "write note",
			wantMessage: `write scratchpad note "note/key" failed`,
			invoke: func(fs *faultAFS, service *Service, root string, cause error) error {
				fs.existsFn = func(string) (bool, error) { return true, nil }
				fs.uploadFn = func(string) error { return cause }
				_, err := service.Memorize(ctx, &MemorizeInput{Key: "note/key", Description: "Note", Body: "body"})
				return err
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fs := &faultAFS{Service: afs.New()}
			service := New(WithAFS(fs), WithRootURI("mem://localhost/scratchpad_errors/${userID}"), WithUserID("tester"))
			root, _, err := service.ResolveRootURI()
			require.NoError(t, err)
			cause := &scratchpadStorageError{operation: test.name}

			err = test.invoke(fs, service, root, cause)
			require.EqualError(t, err, test.wantMessage+": "+cause.Error())
			require.True(t, errors.Is(err, cause))
			var actual *scratchpadStorageError
			require.True(t, errors.As(err, &actual))
			require.Same(t, cause, actual)
		})
	}
}

func TestNoteNotFoundErrorFormatsAvailableContext(t *testing.T) {
	cause := errors.New("storage failure: complete cause text")
	tests := []struct {
		name string
		err  *NoteNotFoundError
		want string
	}{
		{
			name: "all fields",
			err:  &NoteNotFoundError{Key: "note/key", URI: "mem://localhost/note.json", Cause: cause},
			want: `scratchpad note not found (key="note/key", uri="mem://localhost/note.json"): storage failure: complete cause text`,
		},
		{
			name: "empty key",
			err:  &NoteNotFoundError{URI: "mem://localhost/note.json", Cause: cause},
			want: `scratchpad note not found (uri="mem://localhost/note.json"): storage failure: complete cause text`,
		},
		{
			name: "empty cause",
			err:  &NoteNotFoundError{Key: "note/key", URI: "mem://localhost/note.json"},
			want: `scratchpad note not found (key="note/key", uri="mem://localhost/note.json")`,
		},
		{
			name: "only key",
			err:  &NoteNotFoundError{Key: "note/key"},
			want: `scratchpad note not found (key="note/key")`,
		},
		{
			name: "no context",
			err:  &NoteNotFoundError{},
			want: `scratchpad note not found`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.EqualError(t, test.err, test.want)
			require.ErrorIs(t, test.err, ErrNoteNotFound)
		})
	}
}

func TestServiceFetchNotFoundFromExistsIncludesKeyAndURI(t *testing.T) {
	const key = "note/key"
	fs := &faultAFS{
		Service: afs.New(),
		existsFn: func(string) (bool, error) {
			return false, nil
		},
	}
	service := New(WithAFS(fs), WithRootURI("mem://localhost/scratchpad_not_found_exists/${userID}"), WithUserID("tester"))
	root, _, err := service.ResolveRootURI()
	require.NoError(t, err)
	wantURI := NoteURL(root, key)

	_, err = service.Fetch(context.Background(), key)
	require.EqualError(t, err, `scratchpad note not found (key="note/key", uri="`+wantURI+`")`)
	require.ErrorIs(t, err, ErrNoteNotFound)
	require.NotErrorIs(t, err, os.ErrNotExist)
	var notFound *NoteNotFoundError
	require.ErrorAs(t, err, &notFound)
	require.Equal(t, key, notFound.Key)
	require.Equal(t, wantURI, notFound.URI)
	require.Nil(t, notFound.Cause)
}

func TestServiceFetchNotFoundPreservesStorageCause(t *testing.T) {
	const key = "note/key"
	pathErr := &os.PathError{Op: "open", Path: "/scratchpad/note.json", Err: os.ErrNotExist}
	cause := fmt.Errorf("download scratchpad note: %w", pathErr)
	fs := &faultAFS{
		Service: afs.New(),
		existsFn: func(string) (bool, error) {
			return true, nil
		},
		downloadFn: func(string) ([]byte, error) {
			return nil, cause
		},
	}
	service := New(WithAFS(fs), WithRootURI("mem://localhost/scratchpad_not_found/${userID}"), WithUserID("tester"))
	root, _, err := service.ResolveRootURI()
	require.NoError(t, err)
	wantURI := NoteURL(root, key)

	_, err = service.Fetch(context.Background(), key)
	require.EqualError(t, err, `scratchpad note not found (key="note/key", uri="`+wantURI+`"): `+cause.Error())
	require.ErrorIs(t, err, ErrNoteNotFound)
	require.ErrorIs(t, err, os.ErrNotExist)
	var notFound *NoteNotFoundError
	require.ErrorAs(t, err, &notFound)
	require.Equal(t, key, notFound.Key)
	require.Equal(t, wantURI, notFound.URI)
	require.Same(t, cause, notFound.Cause)
	var actual *os.PathError
	require.ErrorAs(t, err, &actual)
	require.Same(t, pathErr, actual)
}

func TestServiceDeleteStorageErrorPreservesCause(t *testing.T) {
	cause := &scratchpadStorageError{operation: "delete note"}
	fs := &faultAFS{
		Service:  afs.New(),
		existsFn: func(string) (bool, error) { return false, nil },
		deleteFn: func(string) error { return cause },
	}
	service := New(WithAFS(fs), WithRootURI("mem://localhost/scratchpad_delete_errors/${userID}"), WithUserID("tester"))

	err := service.Delete(context.Background(), "note/key")
	require.EqualError(t, err, "delete scratchpad note failed: "+cause.Error())
	require.True(t, errors.Is(err, cause))
	var actual *scratchpadStorageError
	require.True(t, errors.As(err, &actual))
	require.Same(t, cause, actual)
}

func TestServiceOpenArtifactStorageErrorPreservesCause(t *testing.T) {
	cause := &scratchpadStorageError{operation: "open artifact source"}
	fs := &faultAFS{Service: afs.New()}
	service := New(WithAFS(fs), WithRootURI("mem://localhost/scratchpad_artifact_errors/${userID}"), WithUserID("tester"))
	_, err := service.Memorize(context.Background(), &MemorizeInput{
		Key:         ArtifactKey("a1"),
		Description: "Artifact a1",
		Body:        `{"artifactId":"a1","sourceURL":"mem://localhost/private/a1"}`,
	})
	require.NoError(t, err)
	fs.openURLFn = func(string) (io.ReadCloser, error) { return nil, cause }

	artifact, reader, err := service.OpenArtifact(context.Background(), "a1")
	require.Nil(t, artifact)
	require.Nil(t, reader)
	require.EqualError(t, err, "read artifact source failed: "+cause.Error())
	require.True(t, errors.Is(err, cause))
	var actual *scratchpadStorageError
	require.True(t, errors.As(err, &actual))
	require.Same(t, cause, actual)
}

func TestServiceStorageErrorWrappersPreserveContextCauses(t *testing.T) {
	tests := []struct {
		name        string
		newContext  func() (context.Context, context.CancelFunc)
		wantCause   error
		wantMessage string
		invoke      func(context.Context, *faultAFS, *Service) error
	}{
		{
			name: "cancellation during delete",
			newContext: func() (context.Context, context.CancelFunc) {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx, cancel
			},
			wantCause:   context.Canceled,
			wantMessage: "delete scratchpad note failed",
			invoke: func(ctx context.Context, fs *faultAFS, service *Service) error {
				fs.existsFn = func(string) (bool, error) { return false, nil }
				fs.deleteFn = func(string) error { return ctx.Err() }
				return service.Delete(ctx, "note/key")
			},
		},
		{
			name: "deadline during artifact open",
			newContext: func() (context.Context, context.CancelFunc) {
				return context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
			},
			wantCause:   context.DeadlineExceeded,
			wantMessage: "read artifact source failed",
			invoke: func(ctx context.Context, fs *faultAFS, service *Service) error {
				fs.openURLFn = func(string) (io.ReadCloser, error) { return nil, ctx.Err() }
				_, _, err := service.OpenArtifact(ctx, "a1")
				return err
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fs := &faultAFS{Service: afs.New()}
			service := New(WithAFS(fs), WithRootURI("mem://localhost/scratchpad_context_errors/${userID}"), WithUserID("tester"))
			_, err := service.Memorize(context.Background(), &MemorizeInput{
				Key:         ArtifactKey("a1"),
				Description: "Artifact a1",
				Body:        `{"artifactId":"a1","sourceURL":"mem://localhost/private/a1"}`,
			})
			require.NoError(t, err)
			ctx, cancel := test.newContext()
			defer cancel()
			require.ErrorIs(t, ctx.Err(), test.wantCause)

			err = test.invoke(ctx, fs, service)
			require.EqualError(t, err, test.wantMessage+": "+test.wantCause.Error())
			require.True(t, errors.Is(err, test.wantCause))
		})
	}
}

func TestServiceListMissingRootRemainsEmpty(t *testing.T) {
	pathErr := &os.PathError{Op: "readdir", Path: "/scratchpad", Err: os.ErrNotExist}
	cause := fmt.Errorf("list scratchpad root: %w", pathErr)
	fs := &faultAFS{
		Service:  afs.New(),
		existsFn: func(string) (bool, error) { return true, nil },
		listFn:   func(string) ([]storage.Object, error) { return nil, cause },
	}
	service := New(WithAFS(fs), WithRootURI("mem://localhost/scratchpad_list_not_found/${userID}"), WithUserID("tester"))

	output, err := service.List(context.Background())
	require.NoError(t, err)
	require.Empty(t, output.Entries)
}

func TestProviderExistsReturnsArbitraryFailures(t *testing.T) {
	cause := &scratchpadStorageError{operation: "exists"}
	fs := &faultAFS{
		Service: afs.New(),
		existsFn: func(string) (bool, error) {
			return false, cause
		},
	}
	service := New(WithAFS(fs), WithRootURI("mem://localhost/scratchpad_provider_errors/${userID}"), WithUserID("tester"))
	provider := &storager{service: service, baseURL: "scratchpad://note"}

	exists, err := provider.Exists(context.Background(), "note/key")
	require.False(t, exists)
	require.EqualError(t, err, "read scratchpad note failed: "+cause.Error())
	require.True(t, errors.Is(err, cause))
	var actual *scratchpadStorageError
	require.True(t, errors.As(err, &actual))
	require.Same(t, cause, actual)
}

func TestProviderExistsKeepsNotFoundSemantics(t *testing.T) {
	tests := []struct {
		name       string
		existsFn   func(string) (bool, error)
		downloadFn func(string) ([]byte, error)
	}{
		{
			name:     "exists false",
			existsFn: func(string) (bool, error) { return false, nil },
		},
		{
			name:     "download not found after exists",
			existsFn: func(string) (bool, error) { return true, nil },
			downloadFn: func(string) ([]byte, error) {
				return nil, fmt.Errorf("download failed: %w", os.ErrNotExist)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fs := &faultAFS{Service: afs.New(), existsFn: test.existsFn, downloadFn: test.downloadFn}
			service := New(WithAFS(fs), WithRootURI("mem://localhost/scratchpad_provider_not_found/${userID}"), WithUserID("tester"))
			provider := &storager{service: service, baseURL: "scratchpad://note"}

			exists, err := provider.Exists(context.Background(), "note/key")
			require.NoError(t, err)
			require.False(t, exists)
		})
	}
}
