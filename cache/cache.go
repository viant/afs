package cache

import "time"

//Cache represent a cache
type Cache struct {
	URL   string
	Items []*Entry
	At    time.Time
}
