package afs

import (
	"context"
	"io"
	"os"
	"sync"
	"sync/atomic"

	"github.com/viant/afs/option"
	"github.com/viant/afs/storage"
)

// NewWriter creates an upload writer
func (s *service) NewWriter(ctx context.Context, URL string, mode os.FileMode, options ...storage.Option) (io.WriteCloser, error) {
	empty := &option.Empty{}
	option.Assign(options, &empty)

	manager, err := s.manager(ctx, URL, options)
	if err != nil {
		return nil, err
	}
	if provider, ok := manager.(storage.WriterProvider); ok {
		return provider.NewWriter(ctx, URL, mode, options...)
	}
	return &writer{
		ctx:         ctx,
		url:         URL,
		allowEmpty:  empty.Allowed,
		mode:        mode,
		options:     options,
		uploader:    s,
		opened:      false,
		doneChannel: make(chan bool),
		err:         nil,
	}, nil
}

// A writer writes an object to destination
type writer struct {
	ctx         context.Context
	url         string
	mode        os.FileMode
	options     []storage.Option
	uploader    storage.Uploader
	mutex       sync.RWMutex
	opened      bool
	allowEmpty  bool
	writer      *io.PipeWriter
	doneChannel chan bool
	closeDone   sync.Once
	err         error
	written     int64
}

func (w *writer) markDone() {
	w.closeDone.Do(func() {
		close(w.doneChannel)
	})
}

func (w *writer) open() error {
	pipeReader, pipeWriter := io.Pipe()
	w.writer = pipeWriter
	w.opened = true
	go w.monitorCancel()
	go func() {
		defer w.markDone()
		if err := w.uploader.Upload(w.ctx, w.url, w.mode, pipeReader, w.options...); err != nil {
			w.setError(err)
			pipeReader.CloseWithError(err)
			return
		}
	}()
	return nil
}

// Write appends to pipe writer
func (w *writer) Write(p []byte) (n int, err error) {
	if err := w.error(); err != nil {
		return 0, err
	}
	if !w.opened {
		if err := w.open(); err != nil {
			return 0, err
		}
	}
	n, err = w.writer.Write(p)
	atomic.AddInt64(&w.written, int64(n))
	if err != nil {
		w.setError(err)
		if err == context.Canceled || err == context.DeadlineExceeded {
			return n, err
		}
	}
	return n, err
}

// Close completes the write operation and flushes any buffered data.
//
// When Write has already opened the pipe, the upload goroutine owns closing
// doneChannel. If written is still 0 and empty uploads are disallowed, Close
// must wait for that goroutine instead of closing doneChannel again — the old
// path panicked with "close of closed channel".
func (w *writer) Close() error {
	// nothing was written quit
	if atomic.LoadInt64(&w.written) == 0 && !w.allowEmpty {
		if w.opened {
			if w.writer != nil {
				_ = w.writer.Close()
			}
			<-w.doneChannel
			return w.error()
		}
		w.markDone()
		return nil
	}
	if !w.opened {
		if err := w.open(); err != nil {
			return err
		}
	}
	// Closing either the read or write causes the entire pipe to close.
	if err := w.writer.Close(); err != nil {
		return err
	}
	<-w.doneChannel
	return w.err
}

func (w *writer) monitorCancel() {
	select {
	case <-w.ctx.Done():
		w.setError(w.ctx.Err())
	case <-w.doneChannel:
	}
}

func (w *writer) error() error {
	w.mutex.RLock()
	result := w.err
	w.mutex.RUnlock()
	return result
}

func (w *writer) setError(err error) {
	if err == nil {
		return
	}
	w.mutex.Lock()
	w.err = err
	w.mutex.Unlock()
}
