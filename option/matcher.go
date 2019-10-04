package option

import "os"

//Match represents a matching function
type Match func(parent string, info os.FileInfo) bool

//Matcher represents a matcher
type Matcher interface {
	Match(parent string, info os.FileInfo) bool
}

func DefaultMatch(parent string, info os.FileInfo) bool {
	return true
}

//GetMatchFunc returns supplied matcher or default matcher
func GetMatchFunc(matcher Match) Match {
	if matcher != nil {
		return matcher
	}
	return DefaultMatch
}
