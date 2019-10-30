package base

import (
	"github.com/pkg/errors"
	"github.com/viant/afs/option"
	"io"
	"sync"
)

//RangeHeader represents a range header
const (
	//RangeHeader represents a range header
	RangeHeader     = "Range"
	RangeHeaderTmpl = "bytes=%d-%d"
)

type streamReader struct {
	*option.Stream
	mux         *sync.Mutex
	reader      io.ReadSeeker
	readSoFar   int
	chunk       []byte
	chunkSize   int
	chunkIndex  int
	chunkOffset int
}

func (r *streamReader) getRange() (from, to int) {
	from = (r.chunkIndex * r.PartSize)
	r.chunkIndex++
	to = r.chunkIndex * r.PartSize

	if to > r.Stream.Size {
		to = r.Stream.Size
	}
	r.chunkOffset = 0
	return from, to
}

func (r *streamReader) byteToCopy(destSize int) int {
	toRead := destSize
	if toRead > r.PartSize {
		toRead = r.PartSize
	}
	if toRead > r.chunkSize {
		toRead = r.chunkSize
	}
	remaining := (r.PartSize - r.chunkOffset)
	if toRead > remaining {
		toRead = remaining
	}

	if r.readSoFar+toRead > r.Stream.Size {
		toRead = r.Stream.Size - r.readSoFar
	}
	return toRead
}

func (r *streamReader) Size() int64 {
	return int64(r.Stream.Size)
}

func (r *streamReader) Read(dest []byte) (n int, err error) {
	readSoFar := 0
	destRemaining := len(dest)
begin:
	if r.readSoFar >= r.Stream.Size {
		return 0, io.EOF
	}
	if r.chunkOffset == 0 {
		from, to := r.getRange()
		if _, err := r.reader.Seek(int64(from), io.SeekStart); err != nil {
			return 0, errors.Wrapf(err, "failed to move to position: %v", from)
		}
		expectReadSize := (to - from)
		if r.chunkSize, err = r.reader.Read(r.chunk[:expectReadSize]); err != nil {
			return 0, err
		}
		if r.chunkSize != expectReadSize {
			return 0, errors.Errorf("range error, expected: %v, but had: %v", expectReadSize, r.chunkSize)
		}
	}

	copyCount := r.byteToCopy(destRemaining)
	destOffset := len(dest) - destRemaining
	copy(dest[destOffset:], r.chunk[r.chunkOffset:r.chunkOffset+copyCount])
	r.chunkOffset += copyCount
	destRemaining -= copyCount
	readSoFar += copyCount
	r.readSoFar += copyCount
	if destRemaining > 0 && r.readSoFar < r.Stream.Size {
		r.chunkOffset = 0
		goto begin
	}
	return readSoFar, nil
}

func (r *streamReader) Close() error {
	return nil
}

//ReadAt added reader at interface
func (r *streamReader) ReadAt(dest []byte, off int64) (n int, err error) {
	r.mux.Lock()
	defer r.mux.Unlock()
	if _, err = r.reader.Seek(off, io.SeekStart); err != nil {
		return n, err
	}
	return r.Read(dest)
}

func NewStreamReader(stream *option.Stream, reedSeeker io.ReadSeeker) io.ReadCloser {
	return &streamReader{
		Stream: stream,
		mux:    &sync.Mutex{},
		reader: reedSeeker,
		chunk:  make([]byte, stream.PartSize),
	}
}
