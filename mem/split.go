package mem

import (
	"github.com/viant/afs/url"
	"strings"
)

//SplitPath splits path
func SplitPath(URLPath string) []string {
	var result = make([]string, 0)
	var elements = strings.Split(URLPath, "/")
	if len(elements) == 0 {
		return result
	}
	for _, elem := range elements {
		if elem == "" {
			continue
		}
		result = append(result, elem)
	}
	return result
}

//Split split URL with the last URI element and its parent path
func Split(URL string) (string, string) {
	return url.Split(URL, Scheme)
}
