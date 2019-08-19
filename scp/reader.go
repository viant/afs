package scp

import (
	"fmt"
	"io"
	"sync/atomic"
	"time"
)

type reader struct {
	processors int
	reader     io.Reader
	outputChan chan []byte
	closedChan chan bool
	errorChan  chan error
	closed     uint32
}

func (r *reader) sendCloseNotification() {
	for i := 0; i < r.processors; i++ {
		select {
		case r.closedChan <- true:
		case <-time.After(time.Millisecond):
		}
	}

}

func (r *reader) isClosed() bool {
	return atomic.LoadUint32(&r.closed) == 1
}

func (r *reader) close() {
	if atomic.CompareAndSwapUint32(&r.closed, 0, 1) {
		close(r.errorChan)
		close(r.closedChan)
		close(r.outputChan)
	}
}

func (r *reader) read(timeout time.Duration) ([]byte, error) {
	if r.isClosed() {
		return nil, fmt.Errorf("closed")
	}
	select {
	case data := <-r.outputChan:
		return data, nil
	case err := <-r.errorChan:
		return nil, err
	case <-time.Tick(timeout):
		return nil, fmt.Errorf("exceeded timeout %s", timeout)
	case <-r.closedChan:
		return nil, fmt.Errorf("closed")
	}
}

func (r *reader) readInBackground() {
	for {
		var buffer = make([]byte, 4096)
		n, err := r.reader.Read(buffer)
		if err != nil {
			r.closeWithError(err)
			return
		}
		if r.isClosed() {
			return
		}
		if n == 0 {
			continue
		}
		select {
		case r.outputChan <- buffer[:n]:
			continue
		case <-r.closedChan:
			return
		}
	}
}

func (r *reader) closeWithError(err error) {
	if r.isClosed() {
		return
	}
	r.errorChan <- err
	r.sendCloseNotification()
}

func newReader(ioReader io.Reader) *reader {
	const processors = 3
	return &reader{
		reader:     ioReader,
		outputChan: make(chan []byte, processors),
		closedChan: make(chan bool, processors),
		errorChan:  make(chan error, processors),
	}
}
