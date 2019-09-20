package file

import (
	"fmt"
	"os"
	"strings"
)

//NewMode returns a new file mode for supplied attributes
func NewMode(attributes string) (os.FileMode, error) {
	var result os.FileMode
	if len(attributes) != 10 {
		return result, fmt.Errorf("invalid attribute length %v %v", attributes, len(attributes))
	}

	const fileType = "dalTLDpSugct?"
	var fileModePosition = strings.Index(fileType, string(attributes[0]))

	if fileModePosition != -1 {
		result = 1 << uint(32-1-fileModePosition)
	}

	const filePermission = "rwxrwxrwx"
	for i, c := range filePermission {
		if c == rune(attributes[i+1]) {
			result = result | 1<<uint(9-1-i)
		}
	}
	return result, nil

}

//Mode returns mode for file Info
func Mode(info os.FileInfo) os.FileMode {
	mode := info.Mode()
	if mode == 0 {
		mode = 0644
		if info.IsDir() {
			mode = DefaultDirOsMode
		}
	}
	return mode
}
