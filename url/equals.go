package url

import "strings"

//Equals checks if url are the same
func Equals(URL1, URL2 string) bool {
	base1, path1 := Base(URL1, "file")
	base2, path2 := Base(URL2, "file")
	if base1 != base2 {
		return false
	}
	return strings.Trim(path1, "/") == strings.Trim(path2, "/")
}
