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

//IsSchemeEquals returns true if scheme is equals
func IsSchemeEquals(URL1, URL2 string) bool {
	scheme1 := Scheme(URL1, "file")
	scheme2 := Scheme(URL2, "file")
	return scheme1 == scheme2

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
