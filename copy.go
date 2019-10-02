package afs

import (
	"context"
	"github.com/viant/afs/file"
	"github.com/viant/afs/option"
	"github.com/viant/afs/storage"
	"github.com/viant/afs/url"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strings"
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
	sourceExt := path.Ext(sourceName)
	if sourceExt != "" && !strings.Contains(destName, sourceExt) {
		destPath = path.Join(destPath, sourceName)
	}
	return url.Join(baseURL, destPath)
}

func (s *service) copy(ctx context.Context, sourceURL, destURL string, srcOptions *option.Source, destOptions *option.Dest,
	walker storage.Walker, uploader storage.BatchUploader) (err error) {
	object, err := s.Object(ctx, sourceURL, *srcOptions...)
	destOpts := *destOptions

	mappedName := ""
	if err == nil {
		if object.IsDir() {
			err = s.Create(ctx, destURL, object.Mode()|os.ModeDir, object.IsDir(), destOpts...)
		} else {
			destURL, mappedName = url.Split(destURL, file.Scheme)
		}
	}

	if err != nil {
		return err
	}
	upload, closer, err := uploader.Uploader(ctx, destURL, destOpts...)
	if err != nil {
		return err
	}
	defer func() {
		closeErr := closer.Close()
		if err == nil {
			err = closeErr
		}
	}()

	var modifier option.Modifier
	option.Assign(destOpts, &modifier)
	err = walker.Walk(ctx, sourceURL, func(ctx context.Context, baseURL string, parent string, info os.FileInfo, reader io.Reader) (toContinue bool, err error) {
		if mappedName != "" {
			info = file.NewInfo(mappedName, info.Size(), info.Mode(), info.ModTime(), info.IsDir())
		}
		if modifier != nil {
			info, reader, err = modifier(info, ioutil.NopCloser(reader))
			if err != nil {
				return false, err
			}
		}
		err = upload(ctx, parent, info, reader)
		return err == nil, err
	}, *srcOptions...)
	return err

}

func (s *service) Copy(ctx context.Context, sourceURL, destURL string, options ...storage.Option) (err error) {
	sourceURL = url.Normalize(sourceURL, file.Scheme)
	destURL = url.Normalize(destURL, file.Scheme)
	sourceOptions := option.NewSource()
	destOptions := option.NewDest()

	var walker storage.Walker
	var uploader storage.BatchUploader

	match, modifier := option.GetWalkOptions(options)
	option.Assign(options, &sourceOptions, &destOptions, &match, &walker, &uploader, &modifier)
	if match != nil {
		*sourceOptions = append(*sourceOptions, match)
	}
	if modifier != nil {
		*sourceOptions = append(*sourceOptions, modifier)
	}
	if walker == nil {
		walker = s
	}
	if uploader == nil {
		uploader = s
	}

	destURL = s.updateDestURL(sourceURL, destURL)
	return s.copy(ctx, sourceURL, destURL, sourceOptions, destOptions, walker, uploader)
}
