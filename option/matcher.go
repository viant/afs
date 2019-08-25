package option

import "os"

//Matcher represents a matcher
type Matcher func(parent string, info os.FileInfo) bool

func defaultMatcher(parent string, info os.FileInfo) bool {
	return true
}

//GetMatcher returns supplied matcher or default matcher
func GetMatcher(matcher Matcher) Matcher {
	if matcher != nil {
		return matcher
	}
	return defaultMatcher
}
