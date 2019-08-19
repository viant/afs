package mem

import (
	"os"
	"path"
)

//parent return parent folder
func (s *storager) parent(location string, dirMode os.FileMode) (*Folder, error) {
	root := s.Root
	parentPath, _ := path.Split(location)
	return root.Folder(parentPath, dirMode)
}
