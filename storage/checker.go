package storage

import "context"

//Checker represents abstraction that check if resource exists
type Checker interface {
	//Exists returns true if resource exists
	Exists(ctx context.Context, URL string, options ...Option) (bool, error)
}
