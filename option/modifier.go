package option

import (
	"io"
	"os"
)

//Modifier option to modify content
type Modifier func(info os.FileInfo, reader io.Reader) (io.Reader, error)
