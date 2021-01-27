package whatnot

// PathElementPermission implements access control over a given PathElement
// and possible all of its descendents
type PathElementPermission struct {
	baseElement *PathElement
}
