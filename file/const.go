package file

import "os"

const (
	//DefaultDirOsMode folder mode default
	DefaultDirOsMode = os.ModeDir | 0755
	//DefaultFileOsMode file mode default
	DefaultFileOsMode = os.FileMode(0644)
)
