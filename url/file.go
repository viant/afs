package url

import (
	"path/filepath"
	"runtime"
	"strings"
)

// ToFileURL converts an OS path into a file:// URL in a cross‑platform way.
//   - Leaves existing URLs (with a scheme) unchanged
//   - Converts UNC paths (\\server\share\dir) to file://server/share/dir
//   - Converts Windows drive paths (C:\foo\bar) to file:///C:/foo/bar
//   - Converts POSIX absolute paths (/var/tmp) to file:///var/tmp
//   - Converts other paths to file://<path> with normalized slashes
func ToFileURL(p string) string {
	p = strings.TrimSpace(p)
	if p == "" {
		return ""
	}
	if Scheme(p, "") != "" {
		return p
	}

	q := filepath.Clean(p)

	// UNC path on Windows: \\server\share\path
	if strings.HasPrefix(q, `\\`) {
		s := strings.TrimPrefix(q, `\\`)
		s = strings.ReplaceAll(s, `\\`, "/")
		s = strings.ReplaceAll(s, `\`, "/")
		return "file://" + s
	}

	hasDrive := func(x string) bool {
		return len(x) >= 2 && x[1] == ':' && ((x[0] >= 'A' && x[0] <= 'Z') || (x[0] >= 'a' && x[0] <= 'z'))
	}

	if hasDrive(q) || runtime.GOOS == "windows" {
		s := strings.ReplaceAll(q, `\\`, "/")
		s = strings.ReplaceAll(s, `\`, "/")
		// Ensure exactly one leading slash before drive letter → /C:/...
		if hasDrive(s) {
			if strings.HasPrefix(s, "/") {
				return "file://" + s
			}
			return "file:///" + s
		}
		if strings.HasPrefix(s, "/") {
			return "file://" + s
		}
		return "file:///" + s
	}

	// POSIX absolute
	if strings.HasPrefix(q, string(filepath.Separator)) {
		return "file://" + q
	}
	// Relative path
	s := strings.ReplaceAll(q, `\\`, "/")
	s = strings.ReplaceAll(s, `\`, "/")
	return "file://" + s
}
