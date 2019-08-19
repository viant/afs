package option

const (
	NoCacheBaseURL = iota + 1
)

type NoCache struct {
	Source int
}
