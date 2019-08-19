package url

import (
	"path"
	"strings"
)

//Split split URL with the last URI element and its parent path
func Split(URL, defaultScheme string) (string, string) {
	baseURL, URLPath := Base(URL, defaultScheme)
	if strings.HasSuffix(URLPath, "/") {
		URLPath = string(URLPath[:len(URLPath)-1])
	}
	parent, name := path.Split(URLPath)
	return Join(baseURL, parent), name
}
