package option

import "time"

//RefreshInterval represents interval option
type RefreshInterval struct {
	IntervalMs int
}

//Duration returns a duration
func (i *RefreshInterval) Duration() time.Duration {
	return time.Duration(i.IntervalMs) * time.Millisecond
}

//NewRefreshInterval create refresh interval option
func NewRefreshInterval(intervalMs int) *RefreshInterval {
	return &RefreshInterval{IntervalMs: intervalMs}
}
