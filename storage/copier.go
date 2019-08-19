package storage

import (
	"context"
)

//Copier represents an asset copier
type Copier interface {
	//Copy copies source to dest
	Copy(ctx context.Context, sourceURL, destURL string, options ...Option) error
}
