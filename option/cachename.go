package option

//CacheName  cache name option
type CacheName struct {
	Name string
}

//WithCacheName creates cache name option
func WithCacheName(name string) *CacheName {
	return &CacheName{Name: name}
}
