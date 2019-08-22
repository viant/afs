package walker

import (
	"context"
	"github.com/viant/afs/storage"
	"github.com/viant/afs/url"
	"io"
	"path"
)

type walker struct {
	storage.Manager
	counter      uint32
	locationName string
}

//Walk traverses URL and calls handler on all file or folder
func (f *walker) Walk(ctx context.Context, URL string, handler storage.OnVisit, options ...storage.Option) error {
	f.counter = 0
	_, URLPath := url.Base(URL, f.Manager.Scheme())
	_, f.locationName = path.Split(URLPath)
	return f.walk(ctx, URL, "", handler, options)
}

func (f *walker) visitResource(ctx context.Context, object storage.Object, URL, parent string, handler storage.OnVisit, options []storage.Option) error {
	var err error
	var reader io.ReadCloser

	if !object.IsDir() {
		if reader, err = f.Download(ctx, object, options...); err != nil {
			return err
		}
		defer func() { _ = reader.Close() }()
	}
	if f.counter == 0 && object.IsDir() && f.locationName == object.Name() {
		//skip base location
		return nil
	}
	f.counter++

	toContinue, err := handler(ctx, URL, parent, object, reader)
	if err != nil || !toContinue {
		return err
	}
	if !object.IsDir() {
		return nil
	}

	relative := object.Name()
	if parent != "" {
		relative = path.Join(parent, object.Name())
	}
	if err = f.walk(ctx, URL, relative, handler, options); err != nil {
		return err
	}

	return nil
}

//Walk traverses URL and calls handler on all file or folder
func (f *walker) walk(ctx context.Context, URL, parent string, handler storage.OnVisit, options []storage.Option) error {
	URL = url.Normalize(URL, f.Scheme())
	resourceURL := URL
	if parent != "" {
		resourceURL = url.Join(URL, parent)
	}
	objects, err := f.List(ctx, resourceURL, options...)
	if err != nil {
		return err
	}

	for i := range objects {
		if i == 0 && objects[i].IsDir() {
			continue
		}
		if err = f.visitResource(ctx, objects[i], URL, parent, handler, options); err != nil {
			break
		}

	}
	return err
}

//New create a walker for supplied manager
func New(manager storage.Manager) storage.Walker {
	return &walker{Manager: manager}
}
