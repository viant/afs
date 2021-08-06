package option

//ACL represents acl
type ACL struct {
	ACL string
}

//NewACL creates an acl option
func NewACL(acl string) *ACL {
	return &ACL{ACL: acl}
}
