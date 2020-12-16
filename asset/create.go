package asset

import (
	"context"
	"github.com/viant/afs/storage"
	"github.com/viant/afs/url"
)

//Create creates supplied assets, links or folders in provided location (for testing purpose)
func Create(manager storage.Manager, URL string, resources []*Resource) error {
	return modify(manager, URL, resources, true)
}

//Cleanup removes supplied locations
func Cleanup(manager storage.Manager, URL string) error {
	ctx := context.Background()
	URL = url.Normalize(URL, manager.Scheme())
	return manager.Delete(ctx, URL)
}
