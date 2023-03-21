package cache

import "time"

//Cache represent a cache
type Cache struct {
	Items []*Entry
	At    time.Time
}
