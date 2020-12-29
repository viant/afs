package url

import (
	"strings"
)

//Join joins base URL with path elements
func Join(baseURL string, elements ...string) string {
	if strings.HasSuffix(baseURL, "://") {
		baseURL += Localhost
	} else if strings.HasSuffix(baseURL, "/") {
		index := strings.LastIndex(baseURL, "/")
		if index != -1 {
			baseURL = baseURL[:index]
		}
	}
	if len(elements) == 0 {
		return baseURL
	}

	for i := range elements {
		elements[i] = strings.Trim(elements[i], "/")
	}
	return baseURL + "/" + strings.Join(elements, "/")
}

//JoinUNC joins base URL with path elements, it support '.' or '..' elements
func JoinUNC(baseURL string, fragments ...string) string {
	if strings.HasSuffix(baseURL, "://") {
		baseURL += Localhost
	} else if strings.HasSuffix(baseURL, "/") {
		index := strings.LastIndex(baseURL, "/")
		if index != -1 {
			baseURL = baseURL[:index]
		}
	}
	if len(fragments) == 0 {
		return baseURL
	}
	schema := Scheme(baseURL, "file")
	baseURL, basePath := Base(baseURL, schema)
	result := strings.Split(basePath, "/")
	for i := range fragments {
		fragment := strings.Trim(fragments[i], "/")

		elements := strings.Split(fragment, "/")

		for _, element := range elements {
			if element == "." || element == "" {
				continue
			}
			if element == ".." {
				if len(result) > 0 {
					result = result[:len(result)-1]
				}
				continue
			}
			result = append(result, element)
		}
	}
	location := strings.Join(result, "/")
	return baseURL + "/" + strings.Trim(location, "/")
}
