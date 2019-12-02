package url

import (
	"strings"
)

const (
	//Localhost default host
	Localhost = "localhost"
)

//Host extract host from URL
func Host(URL string) string {

	index := strings.Index(URL, "://")
	if index == -1 {
		return Localhost
	}
	fragment := string(URL[index+3:])
	index = strings.Index(fragment, "/")
	if index != -1 {
		return string(fragment[0:index])
	}
	return fragment
}
