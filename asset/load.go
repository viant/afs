package asset

import (
	"context"
	"github.com/viant/afs/storage"
	"github.com/viant/afs/url"
	"github.com/viant/afs/walker"
	"io"
	"io/ioutil"
	"os"
	"path"
)

//Load loads location resources for supplied manager
func Load(manager storage.Manager, URL string) (map[string]*Resource, error) {
	URL = url.Normalize(URL, manager.Scheme())
	managerWalker, ok := manager.(storage.Walker)
	if !ok {
		managerWalker = walker.New(manager)
	}
	var result = make(map[string]*Resource)
	err := managerWalker.Walk(context.Background(), URL, func(ctx context.Context, baseURL string, parent string, info os.FileInfo, reader io.Reader) (toContinue bool, err error) {
		key := path.Join(parent, info.Name())
		var data []byte
		if !info.IsDir() {
			data, err = ioutil.ReadAll(reader)
		}
		result[key] = New(key, info.Mode(), info.IsDir(), "", data)
		return true, nil
	})
	return result, err
}
