package matcher

import (
	"os"
	"path"
	"path/filepath"
)

//FilepathMatcher returns filepath based filepath matcher
func FilepathMatcher(pattern string) func(baseURL, parent string, info os.FileInfo) bool {
	return func(baseURL, parent string, info os.FileInfo) bool {
		name := path.Join(parent, info.Name())
		hasMatch, _ := filepath.Match(pattern, name)
		return hasMatch
	}
}

//FileMatcher returns filepath based filename matched
func FileMatcher(pattern string) func(info os.FileInfo) bool {
	return func(info os.FileInfo) bool {
		hasMatch, _ := filepath.Match(pattern, info.Name())
		return hasMatch
	}
}
