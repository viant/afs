package url

import (
	"strings"
)

//Base returns base URL and URL path
func Base(URL string, defaultSchema string) (string, string) {
	schema := Scheme(URL, defaultSchema)
	schemaExt := SchemeExtensionURL(URL)
	host := Host(URL)
	path := Path(URL)
	if schemaExt != "" {
		schemaExt = strings.Replace(schemaExt, "://", ":", 1)
		base := schemaExt + "/" + schema + "://" + host
		return base, path
	}
	return schema + "://" + host, path
}
