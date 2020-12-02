package parrot

import (
	"context"
	"fmt"
	"github.com/viant/afs"
	"github.com/viant/afs/file"
	"github.com/viant/afs/storage"
	"github.com/viant/afs/url"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

//Generate generate code that maps source file into destination go files
func Generate(ctx context.Context, src, dest string, useASCII bool, opts ...storage.Option) (err error) {
	fs := afs.New()
	return fs.Walk(ctx, src, func(ctx context.Context, baseURL string, parent string, info os.FileInfo, reader io.Reader) (toContinue bool, err error) {
		if info.IsDir() {
			return true, nil
		}
		var data Data
		data, err = ioutil.ReadAll(reader)
		container := info.Name()
		ext := path.Ext(container)
		source := path.Join(parent, info.Name())
		name := "gen"
		if ext != "" {
			container = strings.ToLower(container[:len(container)-len(ext)])
			name = strings.ToLower(ext[1:])
		}
		destURL := url.Join(dest, path.Join(parent, container, name+".go"))
		payload := fmt.Sprintf(`package %v
//%v content from %v
var %v = %v`, container, strings.ToUpper(name), source, strings.ToUpper(name), data.AsBytesLiteral(useASCII))
		err = fs.Upload(ctx, destURL, file.DefaultFileOsMode, strings.NewReader(payload))
		return err == nil, err
	}, opts...)

}
