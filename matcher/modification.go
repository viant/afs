package matcher

import (
	"github.com/viant/afs/option"
	"os"
	"time"
)

//Modification represents modification matcher
type Modification struct {
	After    *time.Time
	Before   *time.Time
	matchers []option.Match
}

//Match matcher parent and info with matcher rules
func (r *Modification) Match(parent string, info os.FileInfo) bool {
	if r.After != nil {
		if !r.After.Before(info.ModTime()) {
			return false
		}
	}
	if r.Before != nil {
		if !r.Before.After(info.ModTime()) {
			return false
		}
	}
	for i := range r.matchers {
		if !r.matchers[i](parent, info) {
			return false
		}
	}
	return true
}

//NewModification creates a modification time matcher
func NewModification(before, after *time.Time, matchers ...option.Match) *Modification {
	return &Modification{
		Before:   before,
		After:    after,
		matchers: matchers,
	}
}
