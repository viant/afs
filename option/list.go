package option

import (
	"github.com/viant/afs/storage"
)

//GetListOptions returns list options
func GetListOptions(options []storage.Option) (Match, *Page) {
	var matcher Matcher
	var match Match
	page := Page{}
	Assign(options, &match, &page, &matcher)
	if matcher != nil {
		match = matcher.Match
	}
	return GetMatchFunc(match), &page
}
