package matcher

import (
	"os"
	"path"
	"regexp"
	"strings"
)

//Basic represents prefix, suffix or regexp matcher
type Basic struct {
	Prefix   string   `json:",omitempty"`
	Suffix   string   `json:",omitempty"`
	Filter    string `json:",omitempty"`
	Exclusion string `json:",omitempty"`

	Directory        *bool `json:",omitempty"`
	compiledFilter   *regexp.Regexp
	comiledExclusion *regexp.Regexp
}

//Match matcher parent and info with matcher rules
func (r *Basic) Match(parent string, info os.FileInfo) bool {

	if r.Directory != nil {
		expectDir := *r.Directory
		if expectDir != info.IsDir() {
			return false
		}
	}
	if r.Filter != "" && r.compiledFilter == nil {
		r.compiledFilter, _ = regexp.Compile(r.Filter)
	}
	location := path.Join(parent, info.Name())
	if r.compiledFilter != nil {
		if !r.compiledFilter.MatchString(location) {
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
	if r.Exclusion != "" && r.comiledExclusion == nil {
		r.comiledExclusion, _ = regexp.Compile(r.Exclusion)
	}
	if r.comiledExclusion != nil {
		if r.comiledExclusion.MatchString(location) {
			return false
		}
	}
	return true
}

//NewBasic creates basic matcher
func NewBasic(prefix, suffix, filter string, dir *bool) (matcher *Basic, err error) {
	matcher = &Basic{
		Prefix:    prefix,
		Suffix:    suffix,
		Filter:    filter,
		Directory: dir,
	}
	if filter != "" {
		matcher.compiledFilter, err = regexp.Compile(filter)
	}
	return matcher, err
}
