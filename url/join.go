package url

import "strings"

func Join(baseURL string, elements ...string) string {
	if strings.HasSuffix(baseURL, "://") {
		baseURL += Localhost
	} else {
		baseURL = strings.Trim(baseURL, "/")
	}
	if len(elements) == 0 {
		return baseURL
	}

	for i := range elements {
		elements[i] = strings.Trim(elements[i], "/")
	}
	return baseURL + "/" + strings.Join(elements, "/")
}
