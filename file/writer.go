package file

import (
	"context"
	"github.com/viant/afs/option"
	"github.com/viant/afs/storage"
	"github.com/viant/afs/url"
	"io"
	"os"
	"path"
)

func NewWriter(_ context.Context, URL string, mode os.FileMode, options ...storage.Option) (io.WriteCloser, error) {
	flagOpt := option.OsFlag(0)
	option.Assign(options, &flagOpt)
	location := url.Path(URL)
	_, err := os.Stat(location)
	exists := err == nil
	flag := os.O_WRONLY
	if !exists {
		flag |= os.O_CREATE
	}
	if flagOpt > 0 {
		flag |= int(flagOpt)
	} else { //by default append  is file exists
		if exists {
			flag |= os.O_APPEND
		}
	}
	if !exists {
		parent, _ := path.Split(location)
		EnsureParentPathExists(parent, DefaultDirOsMode)
	}
	return os.OpenFile(location, flag, mode)
}
