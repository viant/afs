package url

//Normalize normalizes URL
func Normalize(URL, scheme string) string {
	schema := Scheme(URL, scheme)
	baseURL, URLPath := Base(URL, schema)
	return Join(baseURL, URLPath)
}
