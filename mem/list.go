package mem

import (
	"context"
	"github.com/viant/afs/option"
	"github.com/viant/afs/storage"
	"os"
)

//List list directory or returns a file info
func (s *storager) List(ctx context.Context, location string, options ...storage.Option) ([]os.FileInfo, error) {
	page := &option.Page{}
	var matcher option.Matcher
	_, _ = option.Assign(options, &page, &matcher)
	root := s.Root
	object, err := root.Lookup(location, 0)
	matcher = option.GetMatcher(matcher)
	if err != nil {
		return nil, err
	}
	if object.IsDir() {
		folder := &Folder{}
		if err = object.Unwrap(&folder); err != nil {
			return nil, err
		}
		var objects = folder.Objects()
		var result = make([]os.FileInfo, len(objects))

		for i := range objects {
			if !matcher(location, objects[i]) {
				continue
			}
			page.Increment()
			if page.ShallSkip() {
				continue
			}
			result[i] = objects[i]
			if page.HasReachedLimit() {
				break
			}
		}
		return result, nil
	}
	if !matcher(location, object) {
		return []os.FileInfo{}, nil
	}
	return []os.FileInfo{object}, nil
}

//Exists checks if location exists
func (s *storager) Exists(ctx context.Context, location string) (bool, error) {
	root := s.Root
	_, err := root.Lookup(location, 0)
	if err != nil {
		return false, nil
	}
	return true, nil

}
