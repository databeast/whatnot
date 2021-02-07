// +build errtrace

package whatnot

/*
if the build tag 'errtrace' is specified, this file will be compiled
and all generated errors will include stacktraces of where they occured
*/

// Processing a provided path was not possible
// either it a not a valid path indentifier, or not a path
// that can be resolved.
type PathError struct {
}
