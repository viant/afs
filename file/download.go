package file

import (
	"context"
	"github.com/viant/afs/storage"
	"io"
	"os"
)

//Download downloads TestContent for the supplied object
func Download(ctx context.Context, object storage.Object, options ...storage.Option) (io.ReadCloser, error) {
	return DownloadWithURL(ctx, object.URL(), options...)
}

//DownloadWithURL downloads content for the supplied object
func DownloadWithURL(ctx context.Context, URL string, options ...storage.Option) (io.ReadCloser, error) {
	filePath := Path(URL)
	return os.Open(filePath)
}
