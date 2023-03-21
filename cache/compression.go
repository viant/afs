package cache

import (
	"bytes"
	"compress/gzip"
	"io"
)

func compressWithGzip(data []byte) ([]byte, error) {
	buf := new(bytes.Buffer)
	writer := gzip.NewWriter(buf)
	if _, err := io.Copy(writer, bytes.NewReader(data)); err != nil {
		return nil, err
	}
	if err := writer.Flush(); err != nil {
		return nil, err
	}
	return buf.Bytes(), writer.Close()
}

func uncompressWithGzip(data []byte) ([]byte, error) {
	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	reader.Close()
	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, reader)
	data = buf.Bytes()
	return data, err
}
