package matcher

import (
	"os"
	"path"
	"path/filepath"
)

/*
Filepath returns filepath based filepath matcher

// The pattern syntax is:
//
//	pattern:
//		{ term }
//	term:
//		'*'         matches any sequence of non-Separator characters
//		'?'         matches any single non-Separator character
//		'[' [ '^' ] { character-range } ']'
//		            character class (must be non-empty)
//		c           matches character c (c != '*', '?', '\\', '[')
//		'\\' c      matches character c

*/
func Filepath(pattern string) func(parent string, info os.FileInfo) bool {
	return func(parent string, info os.FileInfo) bool {
		name := path.Join(parent, info.Name())
		hasMatch, _ := filepath.Match(pattern, name)
		return hasMatch
	}
}
