package whatnot

// PathElementPermission implements access control over a given PathElement
// and possible all of its descendents
type PathElementPermission struct {
	onElement *PathElement
}

// ApprovedAction resolves if a given role can perform the requested action
func (p *PathElementPermission) ApprovedAction() {

}
