package option

//BasicAuth represents a basic auth
type BasicAuth interface {
	Credentials() (user, password string)
}

type basicAuth struct {
	user      string
	_password string
}

func (a *basicAuth) Credentials() (user, password string) {
	return a.user, a._password
}

//NewBasicAuth returns credential authenticator
func NewBasicAuth(user, password string) BasicAuth {
	return &basicAuth{user: user, _password: password}
}
