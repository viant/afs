package mem

import (
	"github.com/viant/afs/file"
	"github.com/viant/afs/storage"
	"github.com/viant/afs/url"
)

type storager struct {
	scheme string
	Root   *Folder
}

func (s *storager) Close() error {
	return nil
}

//NewStorager create a new in memeory storage service
func NewStorager(baseURL string) storage.Storager {
	return &storager{
		Root:   NewFolder(baseURL, file.DefaultDirOsMode),
		scheme: url.Scheme(baseURL, Scheme),
	}
}
