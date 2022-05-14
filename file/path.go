package file

import (
	"github.com/viant/afs/url"
	"os"
	"path"
	"regexp"
	"runtime"
	"strings"
)

var rootElement = []string{"/"}

//Path returns shortest path for specified location, if relative it adds current directory
func Path(location string) string {
	location = url.Path(location)
	isAbsolute := strings.HasPrefix(location, "/")

	if runtime.GOOS == "windows" {
		root := location
		matched, _ := regexp.MatchString(`^[a-zA-Z]:`, root)
		if matched {
			isAbsolute = true
		}
	}

	if !isAbsolute {
		if currentDirectory, err := os.Getwd(); err == nil {
			location = path.Join(currentDirectory, location)
		}
	}
	location = path.Clean(location)
	elements := append(rootElement, strings.Split(location, "/")...)

	location = path.Join(elements...)
	location = normalize(location)
	return location
}

func normalize(location string) string {
	if runtime.GOOS != "windows" || location == "" {
		return location
	}

	if location[0] == '/' {
		location = location[1:]
	}
	return location
}
