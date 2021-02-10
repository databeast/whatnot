// +build !errtrace

package whatnot

import "errors"

// PathError indicates that processing a provided path was not possible
// either it a not a valid path identifier, or not a path
// that can be resolved.
type PathError struct {
}

type ConfigError struct {
	error
}

func newConfigError(msg string) error {
	return ConfigError{errors.New(msg)}
}
