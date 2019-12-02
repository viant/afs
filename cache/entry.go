package cache

import "time"

//Entry represents cache entry
type Entry struct {
	URL     string
	ModTime time.Time
	Size    int64
	Data    []byte
}
