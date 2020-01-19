package url

import "strings"

//IsRelative returns true if location is relative path
func IsRelative(location string) bool {
	return !(strings.HasPrefix(location, "/") || strings.Contains(location, "://"))
}
