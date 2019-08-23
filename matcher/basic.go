package matcher

import (
	"os"
	"path"
	"regexp"
	"strings"
)

//Basic represents prefix, suffix or regexp matcher
type Basic struct {
	Prefix string
	Suffix string
	Filter string
	filter *regexp.Regexp
}

//Match matcher parent and info with matcher rules
func (r *Basic) Match(parent string, info os.FileInfo) bool {
	location := path.Join(parent, info.Name())
	if r.filter != nil {
		if !r.filter.MatchString(location) {
			return false
		}
	}
	if r.Prefix != "" {
		if !strings.HasPrefix(location, r.Prefix) {
			return false
		}
	}
	if r.Suffix != "" {
		if !strings.HasSuffix(location, r.Suffix) {
			return false
		}
	}
	return true
}

//NewBasic creates basic matcher
func NewBasic(prefix, suffix, filter string) (matcher *Basic, err error) {
	matcher = &Basic{
		Prefix: prefix,
		Suffix: suffix,
		Filter: filter,
	}
	if filter != "" {
		matcher.filter, err = regexp.Compile(filter)
	}
	return matcher, err
}
