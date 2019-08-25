package url

//Base returns base URL and URL path
func Base(URL string, defaultSchema string) (string, string) {
	schema := Scheme(URL, defaultSchema)
	host := Host(URL)
	path := Path(URL)
	return schema + "://" + host, path
}
