package file

import (
	"context"
	"github.com/pkg/errors"
	"github.com/viant/afs/object"
	"github.com/viant/afs/option"
	"github.com/viant/afs/storage"
	"github.com/viant/afs/url"
	"os"
)

//List list directory or returns a file Info
func List(ctx context.Context, URL string, options ...storage.Option) ([]storage.Object, error) {
	baseURL, filePath := url.Base(URL, Scheme)
	file, err := os.Open(Path(filePath))
	if err != nil {
		return nil, errors.Wrap(err, "unable to open "+filePath)
	}
	var matcher option.ListMatcher
	page := option.Page{}
	_, _ = option.Assign(options, &matcher, &page)
	if matcher == nil {
		matcher = func(info os.FileInfo) bool {
			return true
		}
	}
	defer func() { _ = file.Close() }()
	stat, err := file.Stat()
	if err != nil {
		return nil, err
	}
	if !stat.IsDir() {
		return []storage.Object{
			object.New(URL, stat, nil),
		}, nil
	}
	files, err := file.Readdir(0)
	if err != nil {
		return nil, err
	}

	var result = make([]storage.Object, 0)
	result = append(result, object.New(URL, stat, nil))
	for _, fileInfo := range files {
		if !matcher(fileInfo) {
			continue
		}
		page.Increment()
		if page.ShallSkip() {
			continue
		}
		fileURL := url.Join(baseURL, filePath, fileInfo.Name())
		result = append(result, object.New(fileURL, fileInfo, nil))
		if page.HasReachedLimit() {
			break
		}
	}
	return result, nil
}
