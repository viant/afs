package option

//Proxy represents http proxy
type Proxy struct {
	//URL proxy //URL
	URL string
	//TimeoutMs connection timeout
	TimeoutMs int
	//Fallback if proxy fails retry without proxy
	Fallback bool
}

//NewProxy creates a new proxy
func NewProxy(URL string, timeoutMs int, fallback bool) *Proxy {
	return &Proxy{
		URL:       URL,
		TimeoutMs: timeoutMs,
		Fallback:  fallback,
	}
}
