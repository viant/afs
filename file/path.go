package file

import (
	"github.com/viant/afs/url"
	"os"
	"path"
	"strings"
)

var rootElement = []string{"/"}

//Path returns shortest path for specified location, if relative it adds current directory
func Path(location string) string {
	location = url.Path(location)
	isAbsolute := strings.HasPrefix(location, "/")
	if !isAbsolute {
		if currentDirectory, err := os.Getwd(); err == nil {
			location = path.Join(currentDirectory, location)
		}
	}
	location = path.Clean(location)
	elements := append(rootElement, strings.Split(location, "/")...)
	return path.Join(elements...)
}
