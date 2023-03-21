package url

import (
	"path"
	"strings"
)

//Dir returns director
func Dir(URL string) string {
	_, URLPath := Base(URL, "file")
	if strings.HasSuffix(URLPath, "/") {
		URLPath = string(URLPath[:len(URLPath)-1])
	}
	parent, _ := path.Split(URLPath)
	return parent
}
