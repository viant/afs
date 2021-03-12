package embed

import (
	"context"
	"github.com/viant/afs/file"
	"github.com/viant/afs/storage"
	"io"
	"os"
	"strings"
)

//Open downloads asset for supplied object
func (s *manager) OpenURL(ctx context.Context, URL string, options ...storage.Option) (io.ReadCloser, error) {
	filePath := file.Path(URL)
	filePath = strings.Trim(filePath, "/")
	return os.Open(filePath)
}