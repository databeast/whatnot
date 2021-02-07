// +build !errtrace

package whatnot

// Processing a provided path was not possible
// either it a not a valid path indentifier, or not a path
// that can be resolved.
type PathError struct {
}
