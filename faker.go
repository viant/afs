package afs

//NewFaker returns new faker service. All operation uses in memory service
func NewFaker() Service {
	return newService(true)
}
