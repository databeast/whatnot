package whatnot

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
)

const (
	DELIMITER = "/"
	ROOT_ID   = "[ROOT]"
)

// PathString is a string rppresentation of all or part of
// a hierchical set of resources in a namespace
type PathString string

type SubPath string

// AbsolutePath is the fully-qualified path to a single PathElement
// from the root of the Namespace it resides in
type AbsolutePath []SubPath

// RelativePath is the path to a single PathElement
// relative to a single PathElement somewhere in its parent chain
type RelativePath []SubPath

// ToAbsolutePath converts a PathString into an AbsolutePath
// (a slice of ordered SubPath sections)
func (m PathString) ToAbsolutePath() AbsolutePath {
	if strings.HasPrefix(string(m), DELIMITER) == false {
		return []SubPath{DELIMITER}
	}
	return AbsolutePath(splitPath(m))
}

func (m PathString) ToRelativePath() RelativePath {
	return RelativePath(splitPath(m))
}

// ToPathString converts an absolute path back into
// a delimited string
func (m AbsolutePath) ToPathString() PathString {
	return PathString(fmt.Sprintf("/%s", joinPath(m)))
}

// SubtractPath removes the right-hand-size RelativePath from the AbsolutePath
func (m AbsolutePath) SubtractPath(path AbsolutePath) PathString {
	return ""
}

func splitPath(path PathString) (sections []SubPath) {
	s := strings.Split(string(path), DELIMITER)

	// was this an absolute path, with a leading slash? if so, remove it
	if len(s) > 0 {
		if s[0] == "" {
			s = s[1:]
		}
	}

	for _, ps := range s {
		sections = append(sections, SubPath(ps))
	}

	// SANITY CHECKS
	return sections
}

func joinPath(sections []SubPath) (newPath PathString) {
	strs := make([]string, len(sections))
	for i, p := range sections {
		strs[i] = string(p)
	}
	newPath = PathString(strings.Join(strs, DELIMITER))
	return
}

// Validate confirms that this SubPath entry is usable
// to construct a valid location within an AbsolutePath
// for a given Path Element
func (m SubPath) Validate() error {
	// hard rule for preventing insanity
	if strings.Contains(string(m), DELIMITER) {
		return errors.Errorf("refusing to access")
	}
	return nil
}
