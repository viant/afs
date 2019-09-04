package option

import (
	"sync/atomic"
)

//Page represents a page
type Page struct {
	counter uint32
	limit   int
	offset  int
}

//ShallSkip returns true if item needs to be skipped
func (p *Page) ShallSkip() bool {
	if p.limit == 0 {
		return false
	}
	return int(atomic.LoadUint32(&p.counter)) < p.offset
}

//MaxResult returns max results or zero
func (p *Page) MaxResult() int64 {
	if p.offset > 0 {
		return 0
	}
	return int64(p.limit)
}

//HasReachedLimit returns true if limit has been reaced
func (p *Page) HasReachedLimit() bool {
	if p.limit == 0 {
		return false
	}
	return int(atomic.LoadUint32(&p.counter)) >= p.limit
}

//Increment increment counter
func (p *Page) Increment() int {
	return int(atomic.AddUint32(&p.counter, 1))
}

//NewPage returns a page
func NewPage(offset, limit int) *Page {
	return &Page{
		offset: offset,
		limit:  limit,
	}
}
