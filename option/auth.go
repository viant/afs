package option

//Auth auth options to force auth, instead of reusing previous auth session
type Auth struct {
	Force bool
}

//NewAuth create an auth option
func NewAuth(force bool) *Auth {
	return &Auth{Force: force}
}
