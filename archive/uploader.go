package archive

import (
	"context"
	"github.com/viant/afs/asset"
	"io"
	"io/ioutil"
	"os"
	"path"
)

//RewriteUploader represents rewrite uploaderMe
type RewriteUploader struct {
	resources []*asset.Resource
	listener  func(resources []*asset.Resource) error
}

//Upload returns upload handler, and upload closer for batch upload or error
func (r *RewriteUploader) Upload(ctx context.Context, parent string, info os.FileInfo, reader io.Reader) error {
	var data []byte
	var err error
	if !info.IsDir() {
		if data, err = ioutil.ReadAll(reader); err != nil {
			return err
		}
	}
	resource := asset.New(path.Join(parent, info.Name()), info.Mode(), info.IsDir(), "", data)
	resource.FileInfo = info
	r.resources = append(r.resources, resource)
	return nil
}

//Close notifies specified listener
func (r *RewriteUploader) Close() error {
	return r.listener(r.resources)
}

//NewRewriteUploader returns new rewrite uploader
func NewRewriteUploader(listener func(resources []*asset.Resource) error) *RewriteUploader {
	return &RewriteUploader{
		resources: make([]*asset.Resource, 0),
		listener:  listener,
	}
}
