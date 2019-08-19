package url

import "strings"

//Scheme extracts URL scheme
func Scheme(URL, defaultSchema string) string {
	index := strings.Index(URL, ":")
	if index == -1 {
		return defaultSchema
	}
	return string(URL[:index])
}
