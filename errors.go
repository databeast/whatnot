// +build !errtrace

package whatnot

// PathError indicates that processing a provided path was not possible
// either it a not a valid path identifier, or not a path
// that can be resolved.
type PathError struct {
}
