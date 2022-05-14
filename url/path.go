package url

import (
	"runtime"
	"strings"
)

//Path returns path for an URL or path
func Path(URL string) string {
	location := URL
	if location == "" {
		return "/"
	}

	if runtime.GOOS == "windows" {
		location = strings.ReplaceAll(location, "\\", "/")
	}

	if index := strings.Index(location, "://"); index != -1 {
		location = string(location[index+3:])
		index := strings.Index(location, "/")
		if index == -1 {
			return ""
		}
		location = string(location[index:])
	}
	return location
}
