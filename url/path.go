package url

import (
	"net/url"
	"strings"
)

//Path returns path for an URL or path
func Path(URL string) string {
	location := URL
	if location == "" {
		return "/"
	}
	if strings.Contains(URL, ":/") {
		if parsed, err := url.Parse(URL); err == nil {
			return parsed.Path
		}
	}
	return location
}
