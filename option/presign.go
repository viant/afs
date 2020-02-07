package option

import (
	"net/http"
	"time"
)

//TimeToLive represents presign URL
type PreSign struct {
	URL        string
	Header     http.Header
	TimeToLive time.Duration
}

//NewPreSign  creates a presign option
func NewPreSign(timeToLive time.Duration) *PreSign {
	return &PreSign{TimeToLive: timeToLive}
}
