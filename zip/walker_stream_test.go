package zip

import (
	"archive/zip"
	"bytes"
	"context"
	"errors"
	"github.com/viant/afs/option"
	"github.com/viant/afs/storage"
	"io"
	"os"
	"testing"
)

// readerAtNoRead wraps bytes.Reader to implement io.ReaderAt and io.ReadCloser
// and tracks if Read was ever called (it should not be in streaming path).
type readerAtNoRead struct {
	*bytes.Reader
	readCalled int
}

func (r *readerAtNoRead) Read(p []byte) (int, error) {
	r.readCalled++
	return 0, errors.New("unexpected Read on ReaderAt")
}

func (r *readerAtNoRead) Close() error { return nil }

type testOpener struct{ rc io.ReadCloser }

func (o *testOpener) Open(ctx context.Context, _ storage.Object, _ ...storage.Option) (io.ReadCloser, error) {
	return o.rc, nil
}
func (o *testOpener) OpenURL(ctx context.Context, _ string, _ ...storage.Option) (io.ReadCloser, error) {
	return o.rc, nil
}

func makeZip(t *testing.T, files map[string]string) []byte {
	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)
	for name, content := range files {
		f, err := w.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err = f.Write([]byte(content)); err != nil {
			t.Fatal(err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func TestWalker_ReaderAtWithSize_Streams(t *testing.T) {
	ctx := context.Background()
	data := makeZip(t, map[string]string{"a.txt": "hello", "b/b.txt": "world"})
	rac := &readerAtNoRead{Reader: bytes.NewReader(data)}
	w := newWalker(&testOpener{rc: rac})

	// Provide size via option.Size to enable ReaderAt path
	var visited int
	err := w.Walk(ctx, "mem://ignored.zip", func(ctx context.Context, baseURL string, parent string, info os.FileInfo, r io.Reader) (bool, error) {
		if !info.IsDir() && r != nil {
			// Read some bytes to ensure entry reading works
			var tmp [2]byte
			_, _ = r.Read(tmp[:])
		}
		visited++
		return true, nil
	}, option.Size(len(data)))
	if err != nil {
		t.Fatalf("walk failed: %v", err)
	}
	if rac.readCalled != 0 {
		t.Fatalf("unexpected Read on ReaderAt, called %d times", rac.readCalled)
	}
	if visited == 0 {
		t.Fatalf("expected to visit files")
	}
}
