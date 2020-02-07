package content

//Meta represents content meta options
type Meta struct {
	Values map[string]string
}

//NewMeta represents content meta
func NewMeta(kvPairs ...string) *Meta {
	result := &Meta{Values: make(map[string]string)}
	for i := 0; i < len(kvPairs); i += 2 {
		value := ""
		if i+1 < len(kvPairs) {
			value = kvPairs[i+1]
		}
		result.Values[kvPairs[i]] = value
	}
	return result
}
