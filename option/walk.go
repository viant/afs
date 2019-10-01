package option

import (
	"github.com/viant/afs/storage"
)

//GetWalkOptions returns walk options
func GetWalkOptions(options []storage.Option) (Match, Modifier) {
	var match Match
	var matcher Matcher
	var modifier Modifier
	Assign(options, &match, &modifier, &matcher)
	if matcher != nil {
		match = matcher.Match
	}
	match = GetMatchFunc(match)
	return match, modifier
}
