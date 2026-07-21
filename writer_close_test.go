package afs

import (
	"context"
	"io"
	"os"
	"testing"

	"github.com/viant/afs/storage"
)

type discardUploader struct{}

func (discardUploader) Upload(ctx context.Context, URL string, mode os.FileMode, reader io.Reader, options ...storage.Option) error {
	_, err := io.Copy(io.Discard, reader)
	return err
}

// Write may open the upload pipe while written stays 0 (empty/failed write).
// Close must not panic with "close of closed channel" in that case.
func TestWriterCloseOpenedZeroWrittenNoPanic(t *testing.T) {
	w := &writer{
		ctx:         context.Background(),
		url:         "mem://localhost/test",
		uploader:    discardUploader{},
		doneChannel: make(chan bool),
		allowEmpty:  false,
	}
	if _, err := w.Write(nil); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("second Close: %v", err)
	}
}

func TestWriterCloseNeverOpened(t *testing.T) {
	w := &writer{
		ctx:         context.Background(),
		url:         "mem://localhost/test",
		uploader:    discardUploader{},
		doneChannel: make(chan bool),
		allowEmpty:  false,
	}
	if err := w.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("second Close: %v", err)
	}
}
