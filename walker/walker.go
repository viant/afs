package walker

import (
	"context"
	"github.com/pkg/errors"
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
func (w *walker) Walk(ctx context.Context, URL string, handler storage.OnVisit, options ...storage.Option) error {
	w.counter = 0
	_, URLPath := url.Base(URL, w.Manager.Scheme())
	_, w.locationName = path.Split(URLPath)
	return w.walk(ctx, URL, "", handler, options)
}

func (w *walker) visitResource(ctx context.Context, object storage.Object, URL, parent string, handler storage.OnVisit, options []storage.Option) error {
	var err error
	var reader io.ReadCloser

	if !object.IsDir() {
		if reader, err = w.Open(ctx, object, options...); err != nil {
			return err
		}
		defer func() { _ = reader.Close() }()
	}
	if w.counter == 0 && object.IsDir() && w.locationName == object.Name() {
		//skip base location
		return nil
	}
	w.counter++

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
	if err = w.walk(ctx, URL, relative, handler, options); err != nil {
		return err
	}

	return nil
}

//Walk traverses URL and calls handler on all file or folder
func (w *walker) walk(ctx context.Context, URL, parent string, handler storage.OnVisit, options []storage.Option) error {
	URL = url.Normalize(URL, w.Scheme())
	resourceURL := URL
	if parent != "" {
		resourceURL = url.Join(URL, parent)
	}
	objects, err := w.List(ctx, resourceURL, options...)
	if err != nil {
		return errors.Wrapf(err, "failed to %T.List %v", w.Manager, resourceURL)
	}
	for i := range objects {
		if objects[i].IsDir() && url.Equals(resourceURL, objects[i].URL()) {
			continue
		}
		if err = w.visitResource(ctx, objects[i], URL, parent, handler, options); err != nil {
			break
		}

	}
	return err
}

//New create a walker for supplied manager
func New(manager storage.Manager) storage.Walker {
	return &walker{Manager: manager}
}
