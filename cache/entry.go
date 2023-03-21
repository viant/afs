package cache

import (
	"sync"
	"time"
)

//Entry represents cache entry
type Entry struct {
	URL     string
	ModTime time.Time
	Size    int64
	Data    []byte
}

type Entries struct {
	ptr *[]*Entry
	mux sync.Mutex
}

func (e *Entries) Append(entry *Entry) {
	e.mux.Lock()
	*e.ptr = append(*e.ptr, entry)
	e.mux.Unlock()
}

func NewEntries(ptr *[]*Entry) *Entries {
	return &Entries{ptr: ptr}
}
