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

//GenerateWithMem generate code that pre loads src location assets into memory storage
func GenerateWithMem(ctx context.Context, src, dest string, useASCII bool, opts ...storage.Option) (err error) {
	fs := afs.New()
	var uploads = make([]string, 0)
	err = fs.Walk(ctx, src, func(ctx context.Context, baseURL string, parent string, info os.FileInfo, reader io.Reader) (toContinue bool, err error) {
		if info.IsDir() {
			return true, nil
		}
		var data Data
		data, err = ioutil.ReadAll(reader)
		destURL := url.Join(dest, path.Join(parent, info.Name()))
		uploads = append(uploads,
			fmt.Sprintf(`
	err = fs.Upload(ctx, "%v", file.DefaultFileOsMode, bytes.NewReader(%v))
	if err != nil {
		log.Printf("failed to upload: %v %v", err)
	}
`, destURL, data.AsBytesLiteral(useASCII), destURL, `%v`))
		return true, nil
	}, opts...)

	if err != nil {
		return err
	}
	if len(uploads) == 0 {
		return nil
	}
	payload := fmt.Sprintf(`package %v
import (
	"bytes"
	"log"
	"github.com/viant/afs"
	"github.com/viant/afs/file"
	"context"
)

func init() {
	fs := afs.New()
	ctx := context.Background()
	var err error
	%v
}

`, Pkg(dest), strings.Join(uploads, "\n"))
	return fs.Upload(ctx, dest, file.DefaultFileOsMode, strings.NewReader(payload))
}
