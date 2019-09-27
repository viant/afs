package option

import (
	"io"
	"os"
)

//Modifier option to modify content,
type Modifier func(info os.FileInfo, reader io.ReadCloser) (os.FileInfo, io.ReadCloser, error)
