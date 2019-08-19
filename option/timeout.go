package option

import "time"

//Timeout represents timeout option
type Timeout struct {
	time.Duration
}

//NewTimeout creates a new timeout option
func NewTimeout(durationInMs int) Timeout {
	return Timeout{Duration: time.Millisecond * time.Duration(durationInMs)}
}
