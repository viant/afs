package url

import "strings"

// IsRelative returns true if location is a relative path (no scheme and not absolute).
// On Windows, treats drive-letter (C:\\...) and UNC (\\\\server\\share) as absolute (returns false).
func IsRelative(location string) bool {
	if location == "" {
		return true
	}
	if strings.Contains(location, "://") {
		return false // has scheme
	}
	if strings.HasPrefix(location, "/") {
		return false // posix absolute
	}
	// UNC (\\server\share)
	if strings.HasPrefix(location, `\\`) {
		return false
	}
	// Windows drive letter (C:\ or C:/)
	if len(location) >= 2 && location[1] == ':' {
		c := location[0]
		if (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') {
			return false
		}
	}
	return true
}
