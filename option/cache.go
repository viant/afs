package option

import "strings"

//Cache represents cache option
type Cache struct {
	Name        string
	Compression string
}

func (c *Cache) Init() {
	if strings.HasSuffix(c.Name, ".gz") {
		if c.Compression == "" {
			c.Compression = "gzip"
		}
	} else if c.Compression == "gzip" {
		c.Name += ".gz"
	}
}

//WithCacheName creates cache name option
func WithCacheName(name, compression string) *Cache {
	return &Cache{Name: name, Compression: compression}
}
