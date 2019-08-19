package storage

import "context"

//Mover represents an asset mover
type Mover interface {
	//Move moves source to dest
	Move(ctx context.Context, sourceURL, destURL string, options ...Option) error
}
