package option

import "os"

//WalkerMatcher represent on skip assign, if return true skip processing
type WalkerMatcher func(baseURL, relativePath string, info os.FileInfo) bool

//ListMatcher reprsents a list matcher
type ListMatcher func(info os.FileInfo) bool
