package afs

import (
	"context"
	"github.com/viant/afs/file"
	"github.com/viant/afs/option"
	"github.com/viant/afs/storage"
	"github.com/viant/afs/url"
	"io"
	"os"
	"path"
)

//updateDestURL updated dest file
func (s *service) updateDestURL(sourceURL, destURL string) string {
	sourcePath := url.Path(sourceURL)
	_, sourceName := path.Split(sourcePath)
	baseURL, destPath := url.Base(destURL, file.Scheme)
	_, destName := path.Split(destPath)
	if destName == sourceName {
		return destURL
	}
	return url.Join(baseURL, destPath)
}

func (s *service) copy(ctx context.Context, sourceURL, destURL string, srcOptions *option.Source, destOptions *option.Dest,
	walker storage.Walker, uploader storage.BatchUploader) error {
	destURL = s.updateDestURL(sourceURL, destURL)
	object, err := s.Object(ctx, sourceURL, *srcOptions...)

	destOpts := *destOptions
	if err == nil && object.IsDir() {
		err = s.Create(ctx, destURL, object.Mode(), object.IsDir(), destOpts...)
	}
	if err != nil {
		return err
	}

	upload, closer, err := uploader.Uploader(ctx, destURL, destOpts...)
	if err != nil {
		return err
	}
	defer func() {
		_ = closer.Close()
	}()
	return walker.Walk(ctx, sourceURL, func(ctx context.Context, baseURL string, parent string, info os.FileInfo, reader io.Reader) (toContinue bool, err error) {
		err = upload(ctx, parent, info, reader)
		return err == nil, err
	}, *srcOptions...)

}

func (s *service) Copy(ctx context.Context, sourceURL, destURL string, options ...storage.Option) (err error) {
	sourceURL = url.Normalize(sourceURL, file.Scheme)
	destURL = url.Normalize(destURL, file.Scheme)
	sourceOptions := option.NewSource()
	destOptions := option.NewDest()
	var walker storage.Walker
	var uploader storage.BatchUploader
	var matcher option.Matcher
	_, _ = option.Assign(options, &sourceOptions, &destOptions, &matcher, &walker, &uploader)
	if matcher != nil {
		*sourceOptions = append(*sourceOptions, matcher)
	}
	if walker == nil {
		walker = s
	}
	if uploader == nil {
		uploader = s
	}
	return s.copy(ctx, sourceURL, destURL, sourceOptions, destOptions, walker, uploader)
}
