package storage

//Options represents options
type Options []Option

//NewOptions returns new options
func NewOptions(options []Option, extraOptions ...Option) []Option {
	return append(options, extraOptions...)
}
