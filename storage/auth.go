package storage

import "context"

//Authenticator represents an authennticator
type Authenticator interface {
	//Auth authenticate URL scheme with authentication option
	Auth(baseURL string, option ...Option)
}

//AuthTracker represents auth change tracker
type AuthTracker interface {
	//IsAuthChanged return true if auth has changed
	IsAuthChanged(ctx context.Context, baseURL string, options []Option) bool
}

//StoragerAuthTracker represents auth manager
type StoragerAuthTracker interface {

	//FilterAuthOptions filters auth options
	FilterAuthOptions(option []Option) []Option

	//IsAuthChanged return true if auth has changes
	IsAuthChanged(authOptions []Option) bool
}
