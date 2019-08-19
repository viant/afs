package storage

//Authenticator represents an authennticator
type Authenticator interface {
	//Auth authenticate URL scheme with authentication option
	Auth(baseURL string, option ...Option)
}
