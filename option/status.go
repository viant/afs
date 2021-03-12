package option

//Status represents status code
type Status struct {
	Code int
}

func NewStatus() *Status {
	return &Status{}
}