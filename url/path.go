package url

import (
	"strings"
)

//Path returns path for an URL or path
func Path(URL string) string {
	location := URL
	if location == "" {
		return "/"
	}

	if index := strings.Index(URL, "://"); index != -1 {
		location = string(URL[index+3:])
		index := strings.Index(location, "/")
		if index == -1 {
			return ""
		}
		location = string(location[index:])
	}
	return location
}
