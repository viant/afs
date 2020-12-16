package asset

import (
	"bytes"
	"context"
	"github.com/viant/afs/storage"
	"github.com/viant/afs/url"
	"os"
	"path"
	"path/filepath"
	"time"
)

//Modify modify supplied assets, links or folders in provided location (for testing purpose)
func Modify(manager storage.Manager, URL string, resources []*Resource) error {
	return modify(manager, URL, resources, false)
}

func modify(manager storage.Manager, URL string, resources []*Resource, recreatedURL bool) error {
	if len(resources) == 0 {
		return nil
	}
	URL = url.Normalize(URL, manager.Scheme())
	ctx := context.Background()
	if recreatedURL {
		_ = manager.Delete(ctx, URL)
	}
	_ = manager.Create(ctx, URL, 0744, true)
	baseURL, URLPath := url.Base(URL, manager.Scheme())

	for _, asset := range resources {
		if !asset.Dir {
			continue
		}
		baseURL, URLPath := url.Base(URL, manager.Scheme())
		resourceURL := url.Join(baseURL, path.Join(URLPath, asset.Name))

		if err := manager.Create(ctx, resourceURL, asset.Mode, true); err != nil {
			return err
		}
	}
	for _, asset := range resources {
		if asset.Dir || asset.Link != "" {
			continue
		}
		resourceURL := url.Join(baseURL, path.Join(URLPath, asset.Name))
		modTime := time.Now()
		if asset.ModTime != nil {
			modTime = *asset.ModTime
		}
		if err := manager.Upload(ctx, resourceURL, asset.Mode, bytes.NewReader(asset.Data), modTime); err != nil {
			return err
		}
	}

	for _, asset := range resources {
		if asset.Link == "" {
			continue
		}
		symlink := filepath.Join(URLPath, asset.Name)
		source := path.Join(URLPath, asset.Link)
		if err := os.Symlink(source, symlink); err != nil {
			return err
		}
	}
	return nil
}
