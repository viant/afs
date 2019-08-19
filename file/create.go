package file

import (
	"context"
	"github.com/viant/afs/storage"
	"os"
	"strings"
)

//Create creates a new file or directory
func Create(ctx context.Context, URL string, mode os.FileMode, isDir bool, options ...storage.Option) error {
	filePath := Path(URL)
	if isDir {
		if err := EnsureParentPathExists(filePath, mode); err != nil {
			return err
		}
		return os.MkdirAll(filePath, mode)
	}
	return Upload(ctx, URL, mode, strings.NewReader(""), options...)
}
