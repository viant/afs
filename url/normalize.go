package url

import (
	"os"
	"strings"
)

//Normalize normalizes URL
func Normalize(URL, scheme string) string {
	if strings.Index(URL, ":") == -1 {
		if URL != "" && URL[0] != '/' {
			if basePath, err := os.Getwd(); err == nil {
				URL = Join(basePath, URL)
			}
		}
	}
	schema := Scheme(URL, scheme)
	baseURL, URLPath := Base(URL, schema)
	return Join(baseURL, URLPath)
}
