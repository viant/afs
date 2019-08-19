package url

import (
	"runtime/debug"
	"strings"
)

//Base returns base URL and URL path
func Base(URL string, defaultSchema string) (string, string) {
	schema := Scheme(URL, defaultSchema)
	host := Host(URL)
	path := Path(URL)
	if strings.HasPrefix(URL, "mem:/v") {
		debug.PrintStack()
	}
	return schema + "://" + host, path
}
