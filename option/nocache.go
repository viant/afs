package option

const (
	//NoCacheBaseURL no cache base URL
	NoCacheBaseURL = iota + 1
)

//NoCache represents nocache option
type NoCache struct {
	Source int
}
