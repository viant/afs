package mem

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/viant/afs/file"
	"github.com/viant/afs/option"
	"github.com/viant/afs/storage"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)


var preconditionErrorMessage  = fmt.Sprintf("precondition failed: %v ", http.StatusPreconditionFailed)

//Upload writes fakeReader TestContent to supplied URL path.
func (s *storager) Upload(ctx context.Context, location string, mode os.FileMode, reader io.Reader, options ...storage.Option) error {
	s.mux.Lock()
	defer s.mux.Unlock()

	parent, err := s.parent(location, file.DefaultDirOsMode)
	generation := &option.Generation{}
	_, ok := option.Assign(options, &generation)
	if ! ok {
		generation = nil
	}
	parent, err = s.parent(location, file.DefaultDirOsMode)
	if err != nil {
		return err
	}
	var data []byte
	if reader != nil {
		if data, err = ioutil.ReadAll(reader); err != nil {
			return err
		}
	}
	modTime := time.Now()
	option.Assign(options, &modTime)
	memFile := NewFile(location, mode, data, modTime)


	if prev, ok := parent.files[memFile.Name()]; ok {
		memFile.generation = prev.generation
	}

	if generation != nil {
		if generation.WhenMatch {
			if generation.Generation != memFile.generation {
				return errors.Errorf(preconditionErrorMessage+" expected: %v, but had: %v", generation.Generation, memFile.generation)
			}
		}  else {
			if generation.Generation == memFile.generation {
				return errors.Errorf(preconditionErrorMessage+" unexpected: %v", generation.Generation)
			}
		}
	}


	memFile.generation++
	return parent.Put(memFile.Object)
}
