package option

import "os"

//Matcher represents a matcher
type Matcher func(parent string, info os.FileInfo) bool
