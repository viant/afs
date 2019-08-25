package url

import "strings"

//Scheme extracts URL scheme
func Scheme(URL, defaultSchema string) string {
	index := strings.Index(URL, "://")
	if index == -1 {
		return defaultSchema
	}
	schema := string(URL[:index])

	if index := strings.LastIndex(schema, "/"); index != -1 {
		schema = string(schema[index+1:])
	}
	return schema
}

//SchemeExtensionURL extract scheme extension or empty string
func SchemeExtensionURL(URL string) string {
	index := strings.Index(URL, "://")
	if index == -1 {
		return ""
	}
	schema := string(URL[:index])
	if index := strings.LastIndex(schema, "/"); index != -1 {
		extension := string(schema[:index])
		return strings.Replace(extension, ":", "://", 1)
	}
	return ""
}
