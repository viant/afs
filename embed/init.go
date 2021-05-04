package embed

import (
	"github.com/viant/afs"
)

func init() {
	afs.GetRegistry().Register(Scheme, Provider)
}

